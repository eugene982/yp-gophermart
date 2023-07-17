package application

import (
	"testing"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/storage/memstore"
)

// Прикидываемся приложением
func newMocApplication(t *testing.T) *Application {

	return &Application{
		accrualSystem: "",
		storage:       memstore.New(),
		passwdHashFn: func(r model.LoginReqest) string {
			return r.Password
		},
	}
}
