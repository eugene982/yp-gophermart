package orders

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrders(t *testing.T) {

	type request struct {
		UserID      string
		contentType string
	}
	tests := []struct {
		name       string
		request    request
		wantStatus int
		wantBody   string
	}{
		{
			name: "no content",
			request: request{
				"user",
				"application/json",
			},
			wantStatus: 204,
			wantBody:   "",
		},
		{
			name: "OK",
			request: request{
				"user",
				"application/json",
			},
			wantStatus: 200,
			wantBody:   `[{"accrual":505.05, "number":"12345678903", "status":"NEW", "uploaded_at":"2000-12-31T00:00:00Z"}]`,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			mockDB := mocks.NewDatabase(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Content-Type", tcase.request.contentType)
			r = middleware.RequestWithUserID(r, tcase.request.UserID)

			resp := w.Result()
			defer resp.Body.Close()

			switch tcase.wantStatus {
			case 204:
				mockDB.On("ReadOrders", r.Context(), tcase.request.UserID).
					Return(
						[]model.OrderInfo{},
						nil)
			case 200:
				mockDB.On("ReadOrders", r.Context(), tcase.request.UserID).
					Once().
					Return(
						[]model.OrderInfo{{
							UserID:     tcase.request.UserID,
							OrderID:    12345678903,
							Status:     "NEW",
							UploadedAt: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)}},
						nil)
				mockDB.On("ReadAccruals", r.Context(), tcase.request.UserID).
					Once().
					Return(
						[]model.OperationsInfo{{
							UserID:     "user2",
							OrderID:    12345678903,
							IsAccrual:  true,
							Points:     50505,
							UploadedAt: time.Date(2000, 12, 31, 1, 0, 0, 0, time.UTC)}},
						nil)
			}

			NewGetOrdersHandler(mockDB).ServeHTTP(w, r)

			assert.Equal(t, tcase.wantStatus, w.Code)

			if tcase.wantBody != "" {
				body, err := io.ReadAll(w.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tcase.wantBody, string(body))
			}
		})
	}
}
