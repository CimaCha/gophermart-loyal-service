package handler

import (
	"context"
	model2 "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/model"
	"log/slog"
	"net/http"
)

const maxBodySize = 32

//go:generate go tool mockgen -source=handler.go -destination=mock/user_service_gen.go -package=mock

type GoodsService interface {
	RegisterGoods(ctx context.Context, goods model2.GoodsInfo) error
}

type Handler struct {
	logger  *slog.Logger
	service GoodsService
}

func New(logger *slog.Logger, goodsService GoodsService) *Handler {
	return &Handler{
		logger:  logger,
		service: goodsService,
	}
}

func (h *Handler) RegisterGoods(writer http.ResponseWriter, request *http.Request) {
	//TODO
}
