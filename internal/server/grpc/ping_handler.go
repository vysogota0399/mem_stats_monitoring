package grpc

import (
	"context"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PingHandler struct {
	strg storages.Storage
	lg   *logging.ZapLogger
}

func NewPingHandler(strg storages.Storage, lg *logging.ZapLogger) *PingHandler {
	return &PingHandler{strg: strg, lg: lg}
}

func (h *PingHandler) Ping(ctx context.Context, params *emptypb.Empty) (*emptypb.Empty, error) {
	if err := h.strg.Ping(ctx); err != nil {
		h.lg.ErrorCtx(ctx, "ping error", zap.Error(err))

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
