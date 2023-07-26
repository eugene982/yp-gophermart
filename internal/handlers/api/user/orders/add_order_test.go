package orders

// import (
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/eugene982/yp-gophermart/internal/middleware"
// 	"github.com/eugene982/yp-gophermart/internal/model"
// 	"github.com/eugene982/yp-gophermart/internal/services/database"
// 	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestAddOrder(t *testing.T) {

// 	type request struct {
// 		login       string
// 		contentType string
// 		body        string
// 	}
// 	tests := []struct {
// 		name       string
// 		request    request
// 		wantStatus int
// 	}{
// 		{
// 			name: "add new",
// 			request: request{
// 				`{"login":"user","password":"password"}`,
// 				"text/plain",
// 				`12345678903`,
// 			},
// 			wantStatus: 202,
// 		},
// 		{
// 			name: "exist",
// 			request: request{
// 				`{"login":"user","password":"password"}`,
// 				"text/plain",
// 				`12345678903`,
// 			},
// 			wantStatus: 200,
// 		},
// 		{
// 			name: "bad content-type",
// 			request: request{
// 				`{"login":"user","password":"password"}`,
// 				"application/json",
// 				`12345678903`,
// 			},
// 			wantStatus: 400,
// 		},
// 		{
// 			name: "unautotize",
// 			request: request{
// 				`{"login":"user","password":"-"}`,
// 				"text/plain",
// 				`12345678903`,
// 			},
// 			wantStatus: 401,
// 		},
// 		{
// 			name: "write conflict",
// 			request: request{
// 				`{"login":"user","password":"password"}`,
// 				"text/plain",
// 				`12345678903`,
// 			},
// 			wantStatus: 409,
// 		},
// 		{
// 			name: "bad order",
// 			request: request{
// 				`{"login":"user","password":"password"}`,
// 				"text/plain",
// 				`12345678900`,
// 			},
// 			wantStatus: 422,
// 		},
// 		{
// 			name: "bad order",
// 			request: request{
// 				`{"login":"user","password":"password"}`,
// 				"text/plain",
// 				`12345678903`,
// 			},
// 			wantStatus: 500,
// 		},
// 	}
// 	for _, tcase := range tests {
// 		t.Run(tcase.name, func(t *testing.T) {

// 			mockDB := mocks.NewDatabase(t)
// 			database.DB = mockDB

// 			// auth
// 			w := httptest.NewRecorder()
// 			r := httptest.NewRequest("POST", "/",
// 				strings.NewReader(tcase.request.login))
// 			r.Header.Set("Content-Type", "application/json")

// 			mockDB.On("ReadUser", r.Context(), "user").
// 				Once().
// 				Return(model.UserInfo{UserID: "user", PasswordHash: "password"}, nil)
// 			Login(w, r)

// 			resp := w.Result()
// 			if resp.StatusCode == 401 {
// 				require.Equal(t, tcase.wantStatus, resp.StatusCode)
// 			} else {
// 				require.Equal(t, 200, resp.StatusCode)
// 			}
// 			defer resp.Body.Close()
// 			Cookie := w.Header().Get("Set-Cookie")

// 			w = httptest.NewRecorder()
// 			r = httptest.NewRequest("POST", "/",
// 				strings.NewReader(tcase.request.body))
// 			r.Header.Set("Content-Type", tcase.request.contentType)
// 			r.Header.Set("Cookie", Cookie)

// 			middleware.CookieAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 				switch tcase.wantStatus {
// 				case 200:
// 					mockDB.On("WriteNewOrder", r.Context(), "user", int64(12345678903)).
// 						Once().
// 						Return(database.ErrWriteConflict)
// 					mockDB.On("ReadOrders", r.Context(), "user", int64(12345678903)).
// 						Return(
// 							[]model.OrderInfo{{UserID: "user", OrderID: int64(12345678903)}},
// 							nil,
// 						)
// 				case 202:
// 					mockDB.On("WriteNewOrder", r.Context(), "user", int64(12345678903)).
// 						Once().
// 						Return(nil)
// 				case 409:
// 					mockDB.On("WriteNewOrder", r.Context(), "user", int64(12345678903)).
// 						Once().
// 						Return(database.ErrWriteConflict)
// 					mockDB.On("ReadOrders", r.Context(), "user", int64(12345678903)).
// 						Once().
// 						Return(
// 							[]model.OrderInfo{},
// 							nil,
// 						)
// 				case 500:
// 					mockDB.On("WriteNewOrder", r.Context(), "user", int64(12345678903)).
// 						Once().
// 						Return(fmt.Errorf("mock write error"))

// 				}

// 				AddOrder(w, r)
// 			})).ServeHTTP(w, r)

// 			resp = w.Result()
// 			defer resp.Body.Close()

// 			assert.Equal(t, tcase.wantStatus, w.Code)
// 		})
// 	}
// }
