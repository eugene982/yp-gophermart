package login

// import (
// 	"context"
// 	"fmt"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/eugene982/yp-gophermart/internal/model"
// 	"github.com/eugene982/yp-gophermart/internal/services/database"
// 	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
// 	"github.com/stretchr/testify/assert"
// )

// func TestLogin(t *testing.T) {

// 	mockDB := mocks.NewDatabase(t)
// 	database.DB = mockDB

// 	call := mockDB.On("ReadUser", context.TODO(), "user")

// 	type request struct {
// 		contentType string
// 		body        string
// 	}
// 	tests := []struct {
// 		name       string
// 		request    request
// 		wantStatus int
// 	}{
// 		{
// 			name: "Ok",
// 			request: request{
// 				"application/json",
// 				`{"login":"user","password":"password"}`,
// 			},
// 			wantStatus: 200,
// 		},
// 		{
// 			name: "bad content-type",
// 			request: request{
// 				"text/plaun",
// 				`{"login":"user","password":"password"}`,
// 			},
// 			wantStatus: 400,
// 		},
// 		{
// 			name: "bad body",
// 			request: request{
// 				"application/json",
// 				`{"login":"user"}`,
// 			},
// 			wantStatus: 400,
// 		},
// 		{
// 			name: "emty login",
// 			request: request{
// 				"application/json",
// 				`{"login":"","password":""}`,
// 			},
// 			wantStatus: 400,
// 		},
// 		{
// 			name: "bad password",
// 			request: request{
// 				"application/json",
// 				`{"login":"user", "password":"user"}`,
// 			},
// 			wantStatus: 401,
// 		},
// 		{
// 			name: "internal error",
// 			request: request{
// 				"application/json",
// 				`{"login":"user","password":"password"}`,
// 			},
// 			wantStatus: 500,
// 		},
// 	}
// 	for _, tcase := range tests {
// 		t.Run(tcase.name, func(t *testing.T) {

// 			switch tcase.wantStatus {
// 			case 500:
// 				call.Return(model.UserInfo{}, fmt.Errorf("mock write error"))
// 			default:
// 				call.Return(model.UserInfo{UserID: "user", PasswordHash: "password"}, nil)
// 			}

// 			w := httptest.NewRecorder()
// 			r := httptest.NewRequest("POST", "/",
// 				strings.NewReader(tcase.request.body))
// 			r.Header.Set("Content-Type", tcase.request.contentType)

// 			Login(w, r)
// 			assert.Equal(t, tcase.wantStatus, w.Code)
// 		})
// 	}
// }
