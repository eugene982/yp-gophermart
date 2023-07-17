// Хранение в базе данных postres
package pgstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

type PgxStore struct {
	db *sqlx.DB
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*PgxStore)(nil)

// Функция конструктор
func New(db *sqlx.DB) (*PgxStore, error) {
	err := db.Ping()
	if err != nil {
		return nil, err
	}

	if err = createTablesIfNonExists(db); err != nil {
		return nil, err
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(3 * time.Minute)

	return &PgxStore{db}, nil
}

// Закрытие соединения
func (p *PgxStore) Close() error {
	return p.db.Close()
}

// Пинг к базе
func (p *PgxStore) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Установка уникального соответствия
func (p *PgxStore) WriteUser(ctx context.Context, data model.UserInfo) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO users (user_id, passwd_hash) 
		VALUES(:user_id, :passwd_hash);`

	if _, err = tx.NamedExecContext(ctx, query, data); err != nil {
		return errWriteConflict(err)
	}
	return tx.Commit()
}

// Чтение данных пользователя
func (p *PgxStore) ReadUsers(ctx context.Context, userIDs ...string) (res []model.UserInfo, err error) {
	res = make([]model.UserInfo, 0)
	query := `SELECT * FROM users`

	if len(userIDs) == 0 {
		err = p.db.SelectContext(ctx, &res, query)
	} else {
		var args []any
		query, args, err = sqlx.In(query+`
			WHERE user_id IN (?) LIMIT ?`, userIDs, len(userIDs))
		if err != nil {
			return nil, err
		}
		err = p.db.SelectContext(ctx, &res, p.db.Rebind(query), args...)
	}
	return
}

// Запись закаказа
func (p *PgxStore) WriteOrder(ctx context.Context, data model.OrderInfo) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// добавление записи о заказе
	query := `
		INSERT INTO orders (user_id, order_id, status, uploaded_at) 
		VALUES(:user_id, :order_id, :status, :uploaded_at);`
	if _, err = tx.NamedExecContext(ctx, query, data); err != nil {
		return errWriteConflict(err)
	}
	return tx.Commit()
}

// читаем заказы указанного пользователя по списку номеров, если номера не указаны - читаем всё
func (p *PgxStore) ReadOrders(ctx context.Context, userID string, nums ...int64) (res []model.OrderInfo, err error) {
	res = make([]model.OrderInfo, 0)

	if len(nums) == 0 {
		query := `
			SELECT * FROM orders 
			WHERE user_id = $1;`
		err = p.db.SelectContext(ctx, &res, query, userID)
	} else {
		var (
			query string
			args  []any
		)
		query, args, err = sqlx.In(`
			SELECT * FROM orders
			WHERE user_id = ? AND order_id IN (?) LIMIT ?;`, userID, nums, len(nums))
		if err != nil {
			return nil, err
		}
		err = p.db.SelectContext(ctx, &res, p.db.Rebind(query), args...)
	}
	return
}

// читаем заказы всех пользователей указанных статусов
func (p *PgxStore) ReadOrdersWithStatus(ctx context.Context, status []string, limit int) (res []model.OrderInfo, err error) {
	res = make([]model.OrderInfo, 0)
	query := `SELECT * FROM orders %s`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	if len(status) == 0 {
		err = p.db.SelectContext(ctx, &res, fmt.Sprintf(query, ""))
	} else {
		var args []any
		query, args, err = sqlx.In(fmt.Sprintf(query, `WHERE status IN (?)`),
			status)
		if err != nil {
			return nil, err
		}
		err = p.db.SelectContext(ctx, &res, p.db.Rebind(query), args...)
	}
	return
}

// Запись сведений о лояльности
func (p *PgxStore) WriteLoyalty(ctx context.Context, data []model.LoyaltyInfo) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// добавление записи о заказе
	query := `
		INSERT INTO loyalty (user_id, order_id, is_accrual, points, uploaded_at) 
		VALUES(:user_id, :order_id, :is_accrual, :points, :uploaded_at);`
	if _, err = tx.NamedExecContext(ctx, query, data); err != nil {
		return err
	}
	return tx.Commit()
}

// читаем данные лояльности
func (p *PgxStore) ReadLoyalty(ctx context.Context, userID string, accrual bool) (res []model.LoyaltyInfo, err error) {
	res = make([]model.LoyaltyInfo, 0)

	query := `
		SELECT * FROM loyalty 
		WHERE user_id = $1 AND is_accrual = $2;`
	err = p.db.SelectContext(ctx, &res, query, userID, accrual)

	return
}

// читаем баланс пользователя
func (p *PgxStore) ReadBalances(ctx context.Context, userIDs ...string) (res []model.BalanceInfo, err error) {
	res = make([]model.BalanceInfo, 0)
	query := `
		SELECT
			user_id,
			SUM(CASE WHEN is_accrual THEN points ELSE -points END) AS current,
			SUM(CASE WHEN is_accrual THEN 0 ELSE points END) AS withdrawn 
		FROM loyalty %s
		GROUP BY user_id;`

	if len(userIDs) == 0 {
		err = p.db.SelectContext(ctx, &res, fmt.Sprintf(query, ""))
	} else {
		var args []any
		query, args, err = sqlx.In(fmt.Sprintf(query, `WHERE user_id IN (?)`),
			userIDs)
		if err != nil {
			return nil, err
		}
		err = p.db.SelectContext(ctx, &res, p.db.Rebind(query), args...)
	}
	return
}

// Обновление сведений заказа, добавление записей о начислении скидок
func (p *PgxStore) UpdateOrders(ctx context.Context, orders []model.OrderInfo, accrues []model.LoyaltyInfo) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	query := `
		UPDATE orders SET user_id=:user_id, status=:status, uploaded_at=:uploaded_at  
		WHERE order_id = :order_id;`
	for _, o := range orders {
		_, err = tx.NamedExecContext(ctx, query, o)
		if err != nil {
			return err
		}
	}

	query = `
		INSERT INTO loyalty (user_id, order_id, is_accrual, points, uploaded_at) 
		VALUES(:user_id, :order_id, :is_accrual, :points, :uploaded_at);`
	for _, a := range accrues {
		_, err = tx.NamedExecContext(ctx, query, a)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// При первом запуске база может быть пустая
func createTablesIfNonExists(db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR (100) PRIMARY KEY,
			passwd_hash TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS orders (
			order_id BIGINT PRIMARY KEY,
			user_id VARCHAR (100) NOT NULL,
			status VARCHAR (20) NOT NULL,
			uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
		CREATE INDEX IF NOT EXISTS user_idx 
		ON orders (user_id);
		CREATE INDEX IF NOT EXISTS status_idx 
		ON orders (status);

		CREATE TABLE IF NOT EXISTS loyalty (
			user_id VARCHAR (100) NOT NULL,
			order_id       BIGINT NOT NULL,
			is_accrual     BOOL NOT NULL,
			points numeric(15, 2) NOT NULL,
			uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
		CREATE INDEX IF NOT EXISTS user_idx 
		ON loyalty (user_id);
		`
	_, err := db.Exec(query)
	return err
}

func errWriteConflict(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return storage.ErrWriteConflict
	}
	return err
}
