package balance

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceHandler(t *testing.T) {

	type request struct {
		userID      string
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
				"user",
				"application/json",
			},
			wantStatus: 200,
			wantBody:   `{"current":405.05, "withdrawn":100}`,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Content-Type", tcase.request.contentType)
			r = middleware.RequestWithUserID(r, tcase.request.userID)

			mockDB := mocks.NewDatabase(t)
			mockDB.On("ReadBalance", r.Context(), tcase.request.userID).
				Once().
				Return(
					model.BalanceInfo{
						UserID:    tcase.request.userID,
						Current:   40505,
						Withdrawn: 10000,
					},
					nil,
				)

			resp := w.Result()
			defer resp.Body.Close()

			NewBalanceHandler(mockDB).ServeHTTP(w, r)
			assert.Equal(t, tcase.wantStatus, w.Code)

			if tcase.wantBody != "" {
				body, err := io.ReadAll(w.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tcase.wantBody, string(body))
			}
		})
	}
}
