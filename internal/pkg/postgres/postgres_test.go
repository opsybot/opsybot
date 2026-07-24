package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/opsybot/opsybot/internal/config"
)

func newMockClient(t *testing.T) (Client, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return Client{db}, mock
}

func TestQuerierReturnsPoolOutsideTx(t *testing.T) {
	c, _ := newMockClient(t)
	q := c.Querier(context.Background())
	if q != Querier(c.DB) {
		t.Fatalf("Querier outside tx = %T, want the base pool", q)
	}
}

func TestWithTxCommitsAndQuerierJoins(t *testing.T) {
	c, mock := newMockClient(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := c.WithTx(context.Background(), func(ctx context.Context) error {
		q := c.Querier(ctx)
		if _, ok := q.(*sql.Tx); !ok {
			t.Fatalf("Querier inside WithTx = %T, want *sql.Tx", q)
		}
		_, err := q.ExecContext(ctx, "UPDATE users SET name = $1", "x")
		return err
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	c, mock := newMockClient(t)
	sentinel := errors.New("boom")
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := c.WithTx(context.Background(), func(context.Context) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx error = %v, want %v", err, sentinel)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithTxJoinsActiveTxWithoutNestedBegin(t *testing.T) {
	c, mock := newMockClient(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO a").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO b").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := c.WithTx(context.Background(), func(ctx context.Context) error {
		if _, err := c.Querier(ctx).ExecContext(ctx, "INSERT INTO a VALUES (1)"); err != nil {
			return err
		}
		return c.WithTx(ctx, func(ctx context.Context) error {
			_, err := c.Querier(ctx).ExecContext(ctx, "INSERT INTO b VALUES (1)")
			return err
		})
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithTxBeginError(t *testing.T) {
	c, mock := newMockClient(t)
	sentinel := errors.New("no conn")
	mock.ExpectBegin().WillReturnError(sentinel)

	err := c.WithTx(context.Background(), func(context.Context) error { return nil })
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx error = %v, want %v", err, sentinel)
	}
}

func TestWithTxCommitError(t *testing.T) {
	c, mock := newMockClient(t)
	sentinel := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(sentinel)

	err := c.WithTx(context.Background(), func(context.Context) error { return nil })
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithTx error = %v, want %v", err, sentinel)
	}
}

func TestDSNURLWins(t *testing.T) {
	got := dsn(config.Postgres{URL: "postgres://u:p@h:1/d", Host: "ignored"})
	if got != "postgres://u:p@h:1/d" {
		t.Fatalf("dsn = %q, want the explicit URL", got)
	}
}

func TestDSNFromDiscreteFields(t *testing.T) {
	got := dsn(config.Postgres{
		Host: "db.internal", Port: 6543, User: "ops", Password: "p@ss:word",
		Database: "opsybot", SSLMode: "require",
	})
	want := "postgres://ops:p%40ss%3Aword@db.internal:6543/opsybot?sslmode=require"
	if got != want {
		t.Fatalf("dsn = %q, want %q", got, want)
	}
}

func TestDSNWithoutPassword(t *testing.T) {
	got := dsn(config.Postgres{Host: "localhost", Port: 5432, User: "ops", Database: "d", SSLMode: "disable"})
	want := "postgres://ops@localhost:5432/d?sslmode=disable"
	if got != want {
		t.Fatalf("dsn = %q, want %q", got, want)
	}
}

func TestWithSavepointReleasesOnSuccess(t *testing.T) {
	c, mock := newMockClient(t)
	mock.ExpectBegin()
	mock.ExpectExec("SAVEPOINT probe").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO events").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("RELEASE SAVEPOINT probe").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := c.WithTx(context.Background(), func(ctx context.Context) error {
		return c.WithSavepoint(ctx, "probe", func(ctx context.Context) error {
			_, err := c.Querier(ctx).ExecContext(ctx, "INSERT INTO events DEFAULT VALUES")
			return err
		})
	})
	if err != nil {
		t.Fatalf("WithSavepoint: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithSavepointRollsBackAndKeepsTxUsable(t *testing.T) {
	c, mock := newMockClient(t)
	sentinel := errors.New("duplicate key")
	mock.ExpectBegin()
	mock.ExpectExec("SAVEPOINT probe").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO events").WillReturnError(sentinel)
	mock.ExpectExec("ROLLBACK TO SAVEPOINT probe").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id FROM events").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("ev-1"))
	mock.ExpectCommit()

	err := c.WithTx(context.Background(), func(ctx context.Context) error {
		inner := c.WithSavepoint(ctx, "probe", func(ctx context.Context) error {
			_, err := c.Querier(ctx).ExecContext(ctx, "INSERT INTO events DEFAULT VALUES")
			return err
		})
		if !errors.Is(inner, sentinel) {
			t.Fatalf("inner err = %v, want the insert error", inner)
		}
		row := c.Querier(ctx).QueryRowContext(ctx, "SELECT id FROM events")
		var id string
		return row.Scan(&id)
	})
	if err != nil {
		t.Fatalf("tx stayed usable after savepoint rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithSavepointRejectsUnsafeName(t *testing.T) {
	c, mock := newMockClient(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := c.WithTx(context.Background(), func(ctx context.Context) error {
		return c.WithSavepoint(ctx, "probe; DROP TABLE users", func(context.Context) error {
			t.Fatal("callback must not run for an unsafe savepoint name")
			return nil
		})
	})
	if err == nil {
		t.Fatal("want an error for an unsafe savepoint name")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWithSavepointOutsideTxRunsDirectly(t *testing.T) {
	c, mock := newMockClient(t)
	mock.ExpectExec("INSERT INTO events").WillReturnResult(sqlmock.NewResult(0, 1))

	err := c.WithSavepoint(context.Background(), "probe", func(ctx context.Context) error {
		_, err := c.Querier(ctx).ExecContext(ctx, "INSERT INTO events DEFAULT VALUES")
		return err
	})
	if err != nil {
		t.Fatalf("WithSavepoint outside tx: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
