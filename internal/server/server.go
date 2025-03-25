// Модуль server отвечает за инициализацию и запуск web - сервера. В нем определены эндпоинты, обработчики и middleware.
package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/handlers"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/service"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type Server struct {
	config    config.Config
	router    *gin.Engine
	storage   storage.Storage
	service   *service.Service
	ctx       context.Context
	lg        *logging.ZapLogger
	secretKey []byte
}

func NewServer(ctx context.Context, c config.Config, strg storage.Storage, srvc *service.Service, lg *logging.ZapLogger) (*Server, error) {
	s := Server{
		config:    c,
		router:    gin.New(),
		service:   srvc,
		ctx:       lg.WithContextFields(ctx, zap.String("name", "server")),
		lg:        lg,
		secretKey: []byte(c.Key),
	}

	s.router.Use(
		gin.Recovery(),
		headers(),
		httpLogger(ctx, lg),
		s.signer(),
		gzip.Gzip(gzip.DefaultCompression),
	)

	s.storage = strg
	return &s, nil
}

// Start - запускат сервер в отдельной горутине с возможностью gracefull shutdown.
func (s *Server) Start(wg *sync.WaitGroup) {
	pprof.Register(s.router)
	s.router.LoadHTMLGlob("internal/server/templates/*.tmpl")
	s.router.POST("/update/:type/:name/:value", handlers.NewUpdateMetricHandler(s.storage))
	s.router.POST("/update/", handlers.NewRestUpdateMetricHandler(s.storage, s.service, s.lg))
	s.router.POST("/updates/", handlers.NewUpdatesRestMetricHandler(s.storage, s.service, s.lg))
	s.router.POST("/value/", handlers.NewShowRestMetricHandler(s.storage, s.lg))
	s.router.GET("/value/:type/:name", handlers.NewShowMetricHandler(s.storage))
	s.router.GET("/ping", handlers.NewPingHandler(s.storage, s.lg))
	s.router.GET("/", handlers.NewRootHandler(s.storage))

	s.lg.DebugCtx(s.ctx, "start", zap.String("config", s.config.String()))

	server := &http.Server{
		Addr:              s.config.Address,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           s.router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.lg.ErrorCtx(s.ctx, "Failed", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-s.ctx.Done()
		s.lg.DebugCtx(s.ctx, "Graceful shutdown")
		if err := server.Shutdown(s.ctx); err != nil {
			s.lg.ErrorCtx(s.ctx, "Shutdown failed", zap.Error(err))
		}
	}()
}

func headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Content-Type") == "application/json" {
			c.Writer.Header().Set("Content-Type", "application/json")
		}

		c.Next()
	}
}

// httpLogger - middleware для логирования запроса/ответа.
func httpLogger(ctx context.Context, lg *logging.ZapLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		var requestID string
		if reqID := c.Request.Header.Get("X-Request-ID"); reqID != "" {
			requestID = reqID
		} else {
			requestID = uuid.NewV4().String()
		}

		ctx = lg.WithContextFields(ctx,
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Reflect("params", c.Params),
			zap.Reflect("headers", c.Request.Header),
		)

		bodyBuff := &bytes.Buffer{}
		if _, err := io.Copy(bodyBuff, c.Request.Body); err != nil {
			lg.ErrorCtx(ctx, "read body error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bodyBuff)
		c.Set("request_id", requestID)

		lg.InfoCtx(ctx, "request", zap.String("body", bodyBuff.String()))

		c.Next()

		status := c.Writer.Status()
		bodySize := c.Writer.Size()
		lg.InfoCtx(ctx, "response",
			zap.String("body", bodyBuff.String()),
			zap.Int("status", status),
			zap.Int("response_size", bodySize),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

var signHeaderKey string = "HashSHA256"

type signResponseReadWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (rw *signResponseReadWriter) Write(b []byte) (int, error) {
	rw.b.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *signResponseReadWriter) Read(b []byte) (int, error) {
	return rw.b.Read(b)
}

// signer - middleware для проверки подписи запроса.
func (s *Server) signer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(s.secretKey) == 0 {
			c.Next()
			return
		}

		rw := &signResponseReadWriter{b: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rw

		cms := crypto.NewCms(hmac.New(sha256.New, s.secretKey))
		base64sign := c.Request.Header.Get(signHeaderKey)
		if base64sign == "" {
			c.Next()
			return
		}

		sign, err := base64.StdEncoding.DecodeString(base64sign)
		if err != nil {
			s.lg.ErrorCtx(c, "base64 decode sign error", zap.Error(err))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		bodyBuff := &bytes.Buffer{}
		if _, err := io.Copy(bodyBuff, c.Request.Body); err != nil {
			s.lg.ErrorCtx(c, "read body error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bodyBuff)

		if eq, err := cms.Verify(bodyBuff, sign); err != nil || !eq {
			s.lg.ErrorCtx(c, "invalid request signature", zap.Error(err))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Next()

		sign, err = cms.Sign(rw)
		if err != nil {
			s.lg.ErrorCtx(c, "response sign failed", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Header.Set(signHeaderKey, base64.StdEncoding.EncodeToString(sign))
	}
}
