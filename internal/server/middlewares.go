package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/crypto"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

var signHeaderKey string = "HashSHA256"

// signResponseReadWriter wraps gin.ResponseWriter to capture response body for signing
type signResponseReadWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

// Write implements the io.Writer interface
func (rw *signResponseReadWriter) Write(b []byte) (int, error) {
	rw.b.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Read implements the io.Reader interface
func (rw *signResponseReadWriter) Read(b []byte) (int, error) {
	return rw.b.Read(b)
}

func signerMiddleware(lg *logging.ZapLogger, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(cfg.Key) == 0 {
			c.Next()
			return
		}

		rw := &signResponseReadWriter{b: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = rw

		cms := crypto.NewCms(hmac.New(sha256.New, []byte(cfg.Key)))
		base64sign := c.Request.Header.Get(signHeaderKey)
		if base64sign == "" {
			c.Next()
			return
		}

		sign, decodeErr := base64.StdEncoding.DecodeString(base64sign)
		if decodeErr != nil {
			lg.ErrorCtx(c, "base64 decode sign error", zap.Error(decodeErr))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		bodyBuff := &bytes.Buffer{}
		if _, copyErr := io.Copy(bodyBuff, c.Request.Body); copyErr != nil {
			lg.ErrorCtx(c, "read body error", zap.Error(copyErr))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bodyBuff)

		if eq, verifyErr := cms.Verify(bodyBuff, sign); verifyErr != nil || !eq {
			lg.ErrorCtx(c, "invalid request signature", zap.Error(verifyErr))
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Next()

		sign, err := cms.Sign(rw)
		if err != nil {
			lg.ErrorCtx(c, "response sign failed", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Header.Set(signHeaderKey, base64.StdEncoding.EncodeToString(sign))
	}
}

type Decrypter interface {
	Decrypt(ciphertext string) (string, error)
}

func decrypterMiddleware(lg *logging.ZapLogger, cfg *config.Config, decrypter Decrypter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.PrivateKey == nil {
			c.Next()
			return
		}

		bodyBuff := &bytes.Buffer{}
		if _, err := io.Copy(bodyBuff, c.Request.Body); err != nil {
			lg.ErrorCtx(c, "read body error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		decrypted, err := decrypter.Decrypt(bodyBuff.String())
		if err != nil {
			lg.ErrorCtx(c, "decrypt error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBufferString(decrypted))

		c.Next()
	}
}

func headerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Content-Type") == "application/json" {
			c.Writer.Header().Set("Content-Type", "application/json")
		}
	}
}

func httpLoggerMiddleware(lg *logging.ZapLogger) gin.HandlerFunc {
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

		ctx := lg.WithContextFields(c.Request.Context(),
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
