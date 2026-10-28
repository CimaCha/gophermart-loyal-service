package worker

import (
	"context"
	"log/slog"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func newTestGate() *gate {
	return newGate(100, slog.New(slog.DiscardHandler))
}

// startWait запускает g.Wait в отдельной горутине и отдаёт канал с результатом
func startWait(ctx context.Context, g *gate) <-chan error {
	done := make(chan error, 1)
	go func() { done <- g.Wait(ctx) }()
	return done
}

// result неблокирующе проверяет, завершился ли Wait
func result(done <-chan error) (finished bool, err error) {
	select {
	case err = <-done:
		return true, err
	default:
		return false, nil
	}
}

// Проверяе что Wait сразу вернул nil и воркер не завис в лупе
func TestGate_Wait_PassesImmediatelyWithoutPause(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g := newTestGate()
		done := startWait(ctx, g)
		synctest.Wait()

		finished, err := result(done)
		require.True(t, finished)
		require.NoError(t, err)
	})
}

// Проверяет что wait не выпустит воркера пока сон не завершится
func TestGate_Wait_BlocksUntilPauseEnds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g := newTestGate()
		g.PauseFor(10 * time.Second)

		done := startWait(ctx, g)
		synctest.Wait()
		finished, _ := result(done)
		require.False(t, finished)

		// Проверяем что после паузы он финишировал
		time.Sleep(10 * time.Second)
		synctest.Wait()
		finished, err := result(done)
		require.True(t, finished)
		require.NoError(t, err)
	})
}

// Проверяет что Wait не выпустит воркера по старому таймеру, если он обнвоился
func TestGate_Wait_RespectsExtendedPause(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g := newTestGate()
		g.PauseFor(10 * time.Second)
		done := startWait(ctx, g)
		synctest.Wait()

		// Ждём 5 секунд из 10 в паузе
		time.Sleep(5 * time.Second)
		// Увеличиваем паузу ещё на 20 секунд
		g.PauseFor(20 * time.Second)
		// Ждём ещё 5 секунд из 10, по итогу предыдущая паузав 10 секунд истекла
		time.Sleep(5 * time.Second)
		synctest.Wait()
		finished, _ := result(done)
		// Проверяем, что wait не отпускает по старому таймеру
		require.False(t, finished)

		time.Sleep(15 * time.Second)
		synctest.Wait()
		finished, err := result(done)
		require.True(t, finished)
		require.NoError(t, err)
	})
}

// Проверяет что по контексту всё отменится
func TestGate_Wait_CancelledDuringPause(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g := newTestGate()
		g.PauseFor(time.Hour)

		done := startWait(ctx, g)
		synctest.Wait()

		cancel()
		synctest.Wait()

		finished, err := result(done)
		require.True(t, finished)
		require.ErrorIs(t, err, context.Canceled)
	})
}

// Проверяет что время паузы не перезапишется маньшей длительностью чем есть сейчас
func TestGate_PauseFor_OnlyExtends(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g := newTestGate()

		g.PauseFor(10 * time.Second)
		g.PauseFor(time.Second)
		assert.Equal(t, 10*time.Second, g.Remaining())

		g.PauseFor(30 * time.Second)
		assert.Equal(t, 30*time.Second, g.Remaining())
	})
}

func TestGate_LowTargetRPS_NeverExceedsTarget(t *testing.T) {
	g := newGate(0.17, slog.New(slog.DiscardHandler))
	assert.Equal(t, 1, g.burst)

	g.PauseFor(time.Second)

	assert.Equal(t, rate.Limit(0.17), g.limiter.Limit())
}
