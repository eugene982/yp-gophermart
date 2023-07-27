package withdraw

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawHandler(t *testing.T) {

	userID := "user"
	orderID := int64(12345678903)
	points := 50505

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
			name: "OK",
			request: request{
				"application/json",
				`{"order":"12345678903", "sum":505.05}`,
			},
			wantStatus: 200,
		},
		{
			name: "payment required",
			request: request{
				"application/json",
				`{"order":"12345678903", "sum":600}`,
			},
			wantStatus: 402,
		},
		{
			name: "bad order",
			request: request{
				"application/json",
				`{"order":"12345678900", "sum":1000}`,
			},
			wantStatus: 422,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			mockDB := mocks.NewDatabase(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/", strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)
			r = middleware.RequestWithUserID(r, userID)

			if tcase.wantStatus != 422 {
				mockDB.On("ReadBalance", r.Context(), userID).
					Once().
					Return(
						model.BalanceInfo{
							UserID:    userID,
							Current:   points,
							Withdrawn: 0,
						},
						nil)
			}

			if tcase.wantStatus == 200 {
				mockDB.On("WriteWithdraw", r.Context(), userID, orderID, points).
					Once().
					Return(nil)
			}

			NewWithdrawHandler(mockDB).ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tcase.wantStatus, w.Code)

		})
	}
}
