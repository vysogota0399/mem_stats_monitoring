package clients

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type IpSetter struct {
	lg *logging.ZapLogger
}

func NewIpSetter(lg *logging.ZapLogger) *IpSetter {
	return &IpSetter{lg: lg}
}

func (ips IpSetter) Call(r *http.Request) error {
	conn, err := net.Dial("udp", r.Host)
	if err != nil {
		return fmt.Errorf("ip_setter: failed to deal udp connection with %s error", r.Host)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			ips.lg.ErrorCtx(context.Background(), "close connection error", zap.Error(err))
		}
	}()

	r.Header.Add(XRealIPHeader, conn.LocalAddr().(*net.UDPAddr).IP.String())
	return nil
}
