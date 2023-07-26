package application

// import (
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/eugene982/yp-gophermart/internal/services/database/mocks"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestRouterMethods(t *testing.T) {

// 	type want struct {
// 		code        int
// 		body        string
// 		contentType string
// 	}

// 	want404 := want{404, "404 page not found\n", "text/plain"}

// 	tests := []struct {
// 		method string
// 		want   want
// 	}{
// 		{method: http.MethodDelete, want: want404},
// 		{method: http.MethodConnect, want: want404},
// 		{method: http.MethodHead, want: want404},
// 		{method: http.MethodOptions, want: want404},
// 		{method: http.MethodPatch, want: want404},
// 		{method: http.MethodPut, want: want404},
// 		{method: http.MethodTrace, want: want404},
// 	}

// 	mockDB := mocks.NewDatabase(t)
// 	router := newRouter(mockDB)

// 	for _, tcase := range tests {
// 		t.Run(tcase.method, func(t *testing.T) {

// 			r := httptest.NewRequest(tcase.method, "/", nil)
// 			w := httptest.NewRecorder()

// 			router.ServeHTTP(w, r)
// 			resp := w.Result()
// 			defer resp.Body.Close()

// 			assert.Equal(t, tcase.want.code, resp.StatusCode)
// 			assert.Contains(t, resp.Header.Get("Content-Type"), tcase.want.contentType)

// 			body, err := io.ReadAll(resp.Body)
// 			require.NoError(t, err)
// 			assert.Equal(t, tcase.want.body, string(body))
// 		})
// 	}
// }
