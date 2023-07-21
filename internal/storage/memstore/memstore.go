package memstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

type MemStore struct {
	mx         sync.Mutex
	users      map[string]model.UserInfo
	orders     map[int64]model.OrderInfo
	operations []model.OperationsInfo
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*MemStore)(nil)

// конструктор нового хранилища
func New() *MemStore {
	return &MemStore{
		users:      make(map[string]model.UserInfo),
		orders:     make(map[int64]model.OrderInfo),
		operations: make([]model.OperationsInfo, 0),
	}
}

// освобождение ресурсов
func (*MemStore) Close() error {
	return nil
}

// Пинг
func (*MemStore) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Добавление пользователя
func (m *MemStore) WriteUser(ctx context.Context, data model.UserInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.users[data.UserID]; ok {
		return storage.ErrWriteConflict
	}
	m.users[data.UserID] = data
	return nil
}

// Чтение списка пользователей
func (m *MemStore) ReadUser(ctx context.Context, userID string) (res model.UserInfo, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	res, ok := m.users[userID]
	if !ok {
		err = storage.ErrNoContent
	}
	return
}

// запись сведений о заказе
func (m *MemStore) WriteOrder(ctx context.Context, data model.OrderInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.orders[data.OrderID]; ok {
		return storage.ErrWriteConflict
	}

	m.orders[data.OrderID] = data
	return nil
}

// чтение заказов
func (m *MemStore) ReadOrders(ctx context.Context, userID string, nums ...int64) ([]model.OrderInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	res := make([]model.OrderInfo, 0)
	add := func(order model.OrderInfo) {
		if order.UserID == userID {
			res = append(res, order)
		}
	}

	if len(nums) == 0 {
		for _, o := range m.orders {
			add(o)
		}
	} else {
		for _, n := range nums {
			if o, ok := m.orders[n]; ok {
				add(o)
			}
		}
	}
	return res, nil
}

// запись начисления и списания баллов лояльности
func (m *MemStore) WriteOperations(ctx context.Context, data []model.OperationsInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	m.operations = append(m.operations, data...)
	return nil
}

// чтение лояльности
func (m *MemStore) ReadOperations(ctx context.Context, userID string, isAccrual bool) ([]model.OperationsInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	res := make([]model.OperationsInfo, 0)
	for _, v := range m.operations {
		if v.UserID == userID && v.IsAccrual == isAccrual {
			res = append(res, v)
		}
	}
	return res, nil
}

// чтение баланса
func (m *MemStore) ReadBalance(ctx context.Context, userID string) (res model.BalanceInfo, err error) {
	select {
	case <-ctx.Done():
		return res, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, l := range m.operations {
		if l.UserID != userID {
			continue
		}
		if l.IsAccrual {
			res.Current += l.Points
		} else {
			res.Withdrawn += l.Points
			res.Current -= l.Points
		}
	}
	return
}

// чтение заказов указанных статусов
func (m *MemStore) ReadOrdersWithStatus(ctx context.Context, status []string, limit int) ([]model.OrderInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	res := make([]model.OrderInfo, 0)

	for _, o := range m.orders {
		if hasValue(o.Status, status) {
			res = append(res, o)
		}
		if limit > 0 && len(res) == limit {
			break
		}
	}
	return res, nil
}

// обновление заказов
func (m *MemStore) UpdateOrderAccrual(ctx context.Context, order model.OrderInfo, accrual int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.orders[order.OrderID]; !ok {
		return fmt.Errorf("error update order %d, not found", order.OrderID)
	}
	m.orders[order.OrderID] = order

	if accrual != 0 {
		m.operations = append(m.operations, model.OperationsInfo{
			UserID:    order.UserID,
			OrderID:   order.OrderID,
			IsAccrual: true,
			Points:    accrual,
		})
	}
	return nil
}

// Проверка на вхождение значение в список
func hasValue[T comparable](val T, list []T) bool {
	if len(list) == 0 {
		return true
	}
	for _, v := range list {
		if val == v {
			return true
		}
	}
	return false
}
