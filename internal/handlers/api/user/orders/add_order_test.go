package orders

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/services/database"
	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddOrder(t *testing.T) {

	type request struct {
		login       string
		contentType string
		body        string
	}
	tests := []struct {
		name       string
		request    request
		wantStatus int
	}{
		{
			name: "add new",
			request: request{
				`{"login":"user","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 202,
		},
		{
			name: "exist",
			request: request{
				`{"login":"user","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 200,
		},
		{
			name: "bad content-type",
			request: request{
				`{"login":"user","password":"password"}`,
				"application/json",
				`12345678903`,
			},
			wantStatus: 400,
		},
		{
			name: "write conflict",
			request: request{
				`{"login":"user","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 409,
		},
		{
			name: "bad order",
			request: request{
				`{"login":"user","password":"password"}`,
				"text/plain",
				`12345678900`,
			},
			wantStatus: 422,
		},
		{
			name: "bad order",
			request: request{
				`{"login":"user","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 500,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			mockDB := mocks.NewDatabase(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)

			userID := "user"
			orderID := int64(12345678903)
			r = middleware.RequestWithUserID(r, userID)

			switch tcase.wantStatus {
			case 200:
				mockDB.On("WriteNewOrder", r.Context(), userID, orderID).
					Once().
					Return(database.ErrWriteConflict)
				mockDB.On("ReadOrders", r.Context(), userID, orderID).
					Return(
						[]model.OrderInfo{{UserID: userID, OrderID: orderID}},
						nil,
					)
			case 202:
				mockDB.On("WriteNewOrder", r.Context(), userID, orderID).
					Once().
					Return(nil)
			case 409:
				mockDB.On("WriteNewOrder", r.Context(), userID, orderID).
					Once().
					Return(database.ErrWriteConflict)
				mockDB.On("ReadOrders", r.Context(), userID, orderID).
					Once().
					Return(
						[]model.OrderInfo{},
						nil,
					)
			case 500:
				mockDB.On("WriteNewOrder", r.Context(), userID, orderID).
					Once().
					Return(fmt.Errorf("mock write error"))
			}

			NewAddOrderHandler(mockDB).ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}
