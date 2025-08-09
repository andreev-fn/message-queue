package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"server/internal/eventbus"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var _ eventbus.PubSubDriver = (*PubSubDriver)(nil)

type PubSubDriver struct {
	db *sql.DB
}

func NewPubSubDriver(db *sql.DB) *PubSubDriver {
	return &PubSubDriver{db}
}

func (d PubSubDriver) Listen(ctx context.Context, channels []string, h eventbus.DriverEventHandler) error {
	var lastErr error

	// always return ErrBadConn, so the connection won't return to pool, and listeners will be freed
	err := doRawPgx(ctx, d.db, func(conn *pgx.Conn) error {
		for _, channel := range channels {
			if _, err := conn.Exec(ctx, "LISTEN "+pgx.Identifier{channel}.Sanitize()); err != nil {
				lastErr = fmt.Errorf("conn.Exec: %w", err)
				return driver.ErrBadConn
			}
		}

		for {
			notification, err := conn.WaitForNotification(ctx)
			if err != nil {
				lastErr = fmt.Errorf("conn.WaitForNotification: %w", err)
				return driver.ErrBadConn
			}

			h(notification.Channel, notification.Payload)
		}
	})
	if !errors.Is(err, driver.ErrBadConn) {
		// doRawPgx failed itself
		return err
	}

	return lastErr
}

func (d PubSubDriver) Publish(channel string, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return doRawPgx(ctx, d.db, func(conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "SELECT pg_notify($1, $2)", channel, message); err != nil {
			return fmt.Errorf("conn.Exec: %w", err)
		}
		return nil
	})
}

func doRawPgx(ctx context.Context, db *sql.DB, f func(conn *pgx.Conn) error) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Raw(func(driverConn any) error {
		conn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return errors.New("driver is not pgx")
		}
		pgxConn := conn.Conn()
		return f(pgxConn)
	})
}
