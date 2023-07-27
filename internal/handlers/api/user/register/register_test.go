package register

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/services/database"
	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
	"github.com/eugene982/yp-gophermart/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	hasher := utils.HasherFunc(func(lr model.LoginReqest) string {
		return lr.Password
	})

	db := mocks.NewDatabase(t)
	call := db.On("WriteUser", context.TODO(), model.UserInfo{UserID: "user", PasswordHash: "password"})

	type request struct {
		contentType string
		body        string
	}
	tests := []struct {
		name       string
		request    request
		wantStatus int
	}{
		{
			name: "Ok",
			request: request{
				"application/json",
				`{"login":"user","password":"password"}`,
			},
			wantStatus: 200,
		},
		{
			name: "bad content-type",
			request: request{
				"text/plaun",
				`{"login":"user","password":"password"}`,
			},
			wantStatus: 400,
		},
		{
			name: "empty pass",
			request: request{
				"application/json",
				`{"login":"user"}`,
			},
			wantStatus: 400,
		},
		{
			name: "emty login",
			request: request{
				"application/json",
				`{"login":"","password":""}`,
			},
			wantStatus: 400,
		},
		{
			name: "user conflict",
			request: request{
				"application/json",
				`{"login":"user","password":"password"}`,
			},
			wantStatus: 409,
		},
		{
			name: "internal error",
			request: request{
				"application/json",
				`{"login":"user","password":"password"}`,
			},
			wantStatus: 500,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			switch tcase.wantStatus {
			case 409:
				call.Return(database.ErrWriteConflict)
			case 500:
				call.Return(fmt.Errorf("mock write error"))
			default:
				call.Return(nil)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)

			NewRegisterHandler(db, hasher).ServeHTTP(w, r)
			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}
