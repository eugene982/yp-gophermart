package memstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage"
)

type MemStore struct {
	mx      sync.Mutex
	users   map[string]model.UserInfo
	orders  map[int64]model.OrderInfo
	loyalty []model.LoyaltyInfo
}

// Утверждение типа, ошибка компиляции
var _ storage.Storage = (*MemStore)(nil)

// конструктор нового хранилища
func New() *MemStore {
	return &MemStore{
		users:   make(map[string]model.UserInfo),
		orders:  make(map[int64]model.OrderInfo),
		loyalty: make([]model.LoyaltyInfo, 0),
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
func (m *MemStore) ReadUsers(ctx context.Context, userIDs ...string) (res []model.UserInfo, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	res = make([]model.UserInfo, 0)

	if len(userIDs) == 0 {
		for _, v := range m.users {
			res = append(res, v)
		}
	} else {
		for _, id := range userIDs {
			if v, ok := m.users[id]; ok {
				res = append(res, v)
			}
		}
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
func (m *MemStore) WriteLoyalty(ctx context.Context, data []model.LoyaltyInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	m.loyalty = append(m.loyalty, data...)
	return nil
}

// чтение лояльности
func (m *MemStore) ReadLoyalty(ctx context.Context, userID string, accrual bool) ([]model.LoyaltyInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	res := make([]model.LoyaltyInfo, 0)
	for _, v := range m.loyalty {
		if v.UserID == userID && v.IsAccrual == accrual {
			res = append(res, v)
		}
	}
	return res, nil
}

// чтение баланса
func (m *MemStore) ReadBalances(ctx context.Context, userIDs ...string) ([]model.BalanceInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	userBalense := make(map[string]model.BalanceInfo)
	for _, l := range m.loyalty {
		if !hasValue(l.UserID, userIDs) {
			continue
		}
		b := userBalense[l.UserID]
		if l.IsAccrual {
			b.Current += l.Points
		} else {
			b.Withdrawn += l.Points
			b.Current -= l.Points
		}
		userBalense[l.UserID] = b
	}

	res := make([]model.BalanceInfo, 0, len(userBalense))
	for _, b := range userBalense {
		res = append(res, b)
	}
	return res, nil
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
func (m *MemStore) UpdateOrders(ctx context.Context, orders []model.OrderInfo, accrues []model.LoyaltyInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, o := range orders {
		if _, ok := m.orders[o.OrderID]; !ok {
			return fmt.Errorf("error update order %d, not found", o.OrderID)
		}
	}
	for _, o := range orders {
		m.orders[o.OrderID] = o
	}

	m.loyalty = append(m.loyalty, accrues...)
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
