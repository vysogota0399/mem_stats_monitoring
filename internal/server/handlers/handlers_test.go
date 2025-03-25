package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap/zapcore"
)

func start(ctx context.Context, server *http.Server) {
	// Запуск сервера
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// Обработка остановки сервера
	go func(ctx context.Context) {
		if err := server.Shutdown(ctx); err != nil {
			panic(err)
		}
	}(ctx)
}

func Example() {
	// Инициализация конфигурации.
	cfg := config.Config{
		LogLevel: -1,
		Address:  `127.0.0.1:8080`,
	}

	// Инициализация хранилища.
	storage := storage.NewMemory()

	// Инициализация логгера.
	logger, _ := logging.MustZapLogger(zapcore.Level(cfg.LogLevel))

	// Инициализация сервисов
	services := service.New(storage)

	// Инициализкация хендлера
	h := handlers.NewUpdatesRestMetricHandler(storage, services, logger)

	// Инициализация сервера.
	router := gin.Default()
	router.POST("/updates/", h)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}
	start(ctx, server)

	// Проверка работы хендлера
	client := http.Client{}
	response, _ := client.Post(
		fmt.Sprintf("http://%s/updates/", cfg.Address),
		`application/json`,
		bytes.NewBuffer(
			[]byte("[{\"id\": \"test\",\"type\": \"counter\",\"delta\": 1}]"),
		),
	)
	response.Body.Close()
}
