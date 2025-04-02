package mitigator

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/core/tcp"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/errors"
	"github.com/RRWM1rr0rB/faraway_lib/backend/golang/logging"

	"app-server/app/internal/config"
)

func (c *Controller) HandleConnection(ctx context.Context, cfg *config.TCPConfig) error {
	server, err := tcp.NewServer(
		cfg.Addr,
		func(conn net.Conn) {
			go func() {
				defer conn.Close()
				ctx, cancel := context.WithTimeout(ctx, cfg.HandlerTimeout)
				defer cancel()
				c.handleClient(ctx, conn, cfg)
			}()
		},
		nil,
	)

	// Настройка TLS, если включено
	var tlsConfig *tls.Config
	if cfg.EnableTLS {
		var err error
		tlsConfig, err = tcp.ServerTLSConfig(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to create TLS config")
		}
	}

	// Создание TCP-сервера с middleware
	server, err := tcp.NewServer(
		cfg.Addr,
		func(conn net.Conn) {
			c.handleClient(ctx, conn, cfg)
		},
		tlsConfig,
		tcp.WithMiddleware(middleware),
		tcp.WithServerLogger(serverLogger),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create TCP server")
	}

	// Запуск сервера
	if err := server.Start(); err != nil {
		return errors.Wrap(err, "failed to start TCP server")
	}

	// Ожидание завершения контекста
	<-ctx.Done()
	logger.Info("Shutting down TCP server")

	// Остановка сервера с таймаутом
	if err := server.StopWithTimeout(5 * time.Second); err != nil {
		logger.Error("Error stopping server", logging.ErrAttr(err))
	}

	return nil
}
