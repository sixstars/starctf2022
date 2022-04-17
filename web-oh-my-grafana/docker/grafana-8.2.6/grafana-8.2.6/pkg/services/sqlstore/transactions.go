package sqlstore

import (
	"context"
	"errors"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/util/errutil"
	"github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

// WithTransactionalDbSession calls the callback with a session within a transaction.
func (ss *SQLStore) WithTransactionalDbSession(ctx context.Context, callback dbTransactionFunc) error {
	return inTransactionWithRetryCtx(ctx, ss.engine, callback, 0)
}

func (ss *SQLStore) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return ss.inTransactionWithRetry(ctx, fn, 0)
}

func (ss *SQLStore) inTransactionWithRetry(ctx context.Context, fn func(ctx context.Context) error, retry int) error {
	return inTransactionWithRetryCtx(ctx, ss.engine, func(sess *DBSession) error {
		withValue := context.WithValue(ctx, ContextSessionKey{}, sess)
		return fn(withValue)
	}, retry)
}

func inTransactionWithRetry(callback dbTransactionFunc, retry int) error {
	return inTransactionWithRetryCtx(context.Background(), x, callback, retry)
}

func inTransactionWithRetryCtx(ctx context.Context, engine *xorm.Engine, callback dbTransactionFunc, retry int) error {
	sess, err := startSession(ctx, engine, true)
	if err != nil {
		return err
	}

	defer sess.Close()

	err = callback(sess)

	// special handling of database locked errors for sqlite, then we can retry 5 times
	var sqlError sqlite3.Error
	if errors.As(err, &sqlError) && retry < 5 && (sqlError.Code == sqlite3.ErrLocked || sqlError.Code == sqlite3.ErrBusy) {
		if rollErr := sess.Rollback(); rollErr != nil {
			return errutil.Wrapf(err, "Rolling back transaction due to error failed: %s", rollErr)
		}

		time.Sleep(time.Millisecond * time.Duration(10))
		sqlog.Info("Database locked, sleeping then retrying", "error", err, "retry", retry)
		return inTransactionWithRetry(callback, retry+1)
	}

	if err != nil {
		if rollErr := sess.Rollback(); rollErr != nil {
			return errutil.Wrapf(err, "Rolling back transaction due to error failed: %s", rollErr)
		}
		return err
	}
	if err := sess.Commit(); err != nil {
		return err
	}

	if len(sess.events) > 0 {
		for _, e := range sess.events {
			if err = bus.Publish(e); err != nil {
				log.Errorf(3, "Failed to publish event after commit. error: %v", err)
			}
		}
	}

	return nil
}

func inTransaction(callback dbTransactionFunc) error {
	return inTransactionWithRetry(callback, 0)
}

func inTransactionCtx(ctx context.Context, callback dbTransactionFunc) error {
	return inTransactionWithRetryCtx(ctx, x, callback, 0)
}
