package tcp

import (
	"net"
	"sync/atomic"
	"time"
)

type trackedConn struct {
	net.Conn
	serverStats *ServerStats // Указатель для атомарного обновления статистики сервера
	idleTimeout time.Duration
}

func newTrackedConn(conn net.Conn, stats *ServerStats, idleTimeout time.Duration) *trackedConn {
	return &trackedConn{
		Conn:        conn,
		serverStats: stats,
		idleTimeout: idleTimeout,
	}
}

// Read перехватывает чтение, обновляет статистику и сбрасывает тайм-аут.
func (t *trackedConn) Read(b []byte) (n int, err error) {
	// Сброс тайм-аута перед чтением
	if t.idleTimeout > 0 {
		err = t.Conn.SetDeadline(time.Now().Add(t.idleTimeout))
		if err != nil {
			return 0, err
		}
	}

	n, err = t.Conn.Read(b)
	if n > 0 {
		atomic.AddInt64(&t.serverStats.BytesRead, int64(n))
		// Обновляем LastActivity сервера атомарно (хотя это может быть дорого)
		// Лучше обновлять его менее часто или только в acceptConnections/handleConnection exit.
		// Для примера оставим так.
		// Consider using atomic.StorePointer for time.Time if needed, but direct assignment
		// is often okay if races on read are acceptable for non-critical stats.
		// For simplicity, we omit the atomic update for time here. A dedicated stats goroutine might be better.
	}
	// Сброс тайм-аута после успешного чтения (или если он не был установлен)
	if t.idleTimeout > 0 && err == nil {
		err = t.Conn.SetDeadline(time.Time{}) // Убираем тайм-аут до следующей операции
		if err != nil {
			// Log or handle error if necessary, but often ignored after successful read
		}
	}
	return n, err
}

// Write перехватывает запись, обновляет статистику и сбрасывает тайм-аут.
func (t *trackedConn) Write(b []byte) (n int, err error) {
	// Сброс тайм-аута перед записью
	if t.idleTimeout > 0 {
		err = t.Conn.SetDeadline(time.Now().Add(t.idleTimeout))
		if err != nil {
			return 0, err
		}
	}

	n, err = t.Conn.Write(b)
	if n > 0 {
		atomic.AddInt64(&t.serverStats.BytesWritten, int64(n))
		// Update LastActivity (see comment in Read)
	}
	// Сброс тайм-аута после успешной записи
	if t.idleTimeout > 0 && err == nil {
		err = t.Conn.SetDeadline(time.Time{}) // Убираем тайм-аут до следующей операции
		if err != nil {
			// Log or handle error if necessary
		}
	}
	return n, err
}

// Close закрывает соединение.
func (t *trackedConn) Close() error {
	return t.Conn.Close()
}
