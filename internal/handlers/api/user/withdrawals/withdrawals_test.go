package withdrawals

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

func TestWithdrawalsHandler(t *testing.T) {

	type request struct {
		contentType string
	}
	tests := []struct {
		name       string
		request    request
		wantStatus int
		wantBody   string
	}{
		{
			name: "OK",
			request: request{
				"application/json",
			},
			wantStatus: 200,
			wantBody:   `[{"order":"12345678903", "processed_at":"2000-12-31T02:00:00Z", "sum":100}]`,
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			userID := "user"
			orderID := int64(12345678903)
			mockDB := mocks.NewDatabase(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Content-Type", tcase.request.contentType)
			r = middleware.RequestWithUserID(r, userID)

			mockDB.On("ReadWithdraws", r.Context(), userID).
				Once().
				Return([]model.OperationsInfo{{
					UserID:     userID,
					OrderID:    orderID,
					IsAccrual:  false,
					Points:     10000,
					UploadedAt: time.Date(2000, 12, 31, 2, 0, 0, 0, time.UTC),
				}},
					nil)

			NewWithdrawalsHandler(mockDB).ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tcase.wantStatus, w.Code)

			if tcase.wantBody != "" {
				body, err := io.ReadAll(w.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tcase.wantBody, string(body))
			}
		})
	}
}
