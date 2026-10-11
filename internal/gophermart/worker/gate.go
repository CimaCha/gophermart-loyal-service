package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type gate struct {
	mu    sync.Mutex
	until time.Time

	warmStart, warmEnd time.Time
	// true если на данный момент идёт прогрев и false если нет
	// чтобы каждый раз не бегать в increaseWarmUpLimit() и не лочить по факту свободных воркеров на mutex
	warming atomic.Bool

	targetRPS float64
	burst     int
	limiter   *rate.Limiter

	l *slog.Logger
}

func newGate(targetRPS float64, logger *slog.Logger) *gate {
	burst := max(int(targetRPS), 1)
	return &gate{
		targetRPS: targetRPS,
		burst:     burst,
		limiter:   rate.NewLimiter(rate.Limit(targetRPS), burst),
		l:         logger,
	}
}

func (g *gate) floor() float64 { return min(1, g.targetRPS) }

// PauseFor - добавляет длительность к паузе
func (g *gate) PauseFor(d time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// если новое время НЕ позже чем старое, то просто скипаем
	// если новое время позже чем старое, значит ставим его как текущее
	end := time.Now().Add(d)
	if !end.After(g.until) {
		return
	}
	g.until = end
	g.warmStart = end      // начала прогрева ставим в конец сна
	g.warmEnd = end.Add(d) // т.к. прогрев длится ровно столько же сколько и сон, ставим конец прогрева спустя время сна

	g.limiter.SetBurst(1) // ставим burst в 1, чтобы резко не выпустить всех воркеров на волю
	g.limiter.SetLimit(rate.Limit(g.floor()))
	g.warming.Store(true) // говорим что прогрев начат
}

// Remaining возвращает либо 0 либо сколько осталось времени до конца until
func (g *gate) Remaining() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()
	return max(time.Until(g.until), 0)
}

// Wait блокируется до конца паузы или отмены ctx, затем плавно начинает пропускать воркеров
func (g *gate) Wait(ctx context.Context) error {
	for {
		d := g.Remaining()
		if d > 0 {
			t := time.NewTimer(d)
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
			}
			continue // отправляем воркера перепроверить remaining
		}

		// если сейчас идёт прогрев отправляем воркеров прогрессировать наш прогрев
		if g.warming.Load() {
			g.advanceWarmUp()
		}

		if err := g.limiter.Wait(ctx); err != nil {
			g.l.Error(
				"gate rate limiter wait failed",
				"err", err,
			)
			return err
		}
		if g.Remaining() == 0 {
			return nil
		}
	}
}

func (g *gate) advanceWarmUp() {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	// Если время прогрева вышло возвращаем на максимальную скорость
	if !now.Before(g.warmEnd) {
		g.limiter.SetLimit(rate.Limit(g.targetRPS))
		g.limiter.SetBurst(g.burst)
		g.warming.Store(false)
		return
	}

	// прогресс равен пропорционально прошедшему времени
	// если прошло половину времени - коэфициент 0.5
	// делим сколько прошло с момента старта на сколько длился прогрев от начала до конца прогрева
	progress := float64(now.Sub(g.warmStart)) / float64(g.warmEnd.Sub(g.warmStart))
	progress = min(max(progress, 0), 1)

	// Линейно увеличиваем RPS в зависимости от прогресса. В качестве старта 5% RPS до 100% от текущего targetRPS
	rps := g.targetRPS * (0.05 + 0.95*progress)
	g.limiter.SetLimit(rate.Limit(max(rps, g.floor())))
}
