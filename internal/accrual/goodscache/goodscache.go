package goodscache

import (
	"context"
	"sync"

	goodsmodel "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
)

type GoodsCache struct {
	mu    sync.RWMutex
	rules []goodsmodel.GoodsInfo

	provider GoodsRrovider
}

type GoodsRrovider interface {
	GetAllGoods(ctx context.Context) ([]goodsmodel.GoodsInfo, error)
}

func New(goodsRepo GoodsRrovider) *GoodsCache {
	return &GoodsCache{
		rules:    []goodsmodel.GoodsInfo{},
		provider: goodsRepo,
	}
}

func (c *GoodsCache) Get() []goodsmodel.GoodsInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rules := make([]goodsmodel.GoodsInfo, len(c.rules))
	copy(rules, c.rules)
	return rules
}

func (c *GoodsCache) Set(rules []goodsmodel.GoodsInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = rules
}

func (c *GoodsCache) Load(ctx context.Context) error {
	rules, err := c.provider.GetAllGoods(ctx)
	if err != nil {
		return err
	}
	c.Set(rules)
	return nil
}

func (c *GoodsCache) Add(rule goodsmodel.GoodsInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = append(c.rules, rule)
}
