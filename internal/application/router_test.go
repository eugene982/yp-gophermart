package application

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eugene982/yp-gophermart/internal/middleware"
	"github.com/eugene982/yp-gophermart/internal/model"
)

func TestRouterMethods(t *testing.T) {

	type want struct {
		code        int
		body        string
		contentType string
	}

	want404 := want{404, "404 page not found\n", "text/plain"}

	tests := []struct {
		method string
		want   want
	}{
		{method: http.MethodDelete, want: want404},
		{method: http.MethodConnect, want: want404},
		{method: http.MethodHead, want: want404},
		{method: http.MethodOptions, want: want404},
		{method: http.MethodPatch, want: want404},
		{method: http.MethodPut, want: want404},
		{method: http.MethodTrace, want: want404},
	}

	app := newMocApplication(t)
	defer app.Close()
	router := app.NewRouter()

	for _, tcase := range tests {
		t.Run(tcase.method, func(t *testing.T) {

			r := httptest.NewRequest(tcase.method, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tcase.want.code, resp.StatusCode)
			assert.Contains(t, resp.Header.Get("Content-Type"), tcase.want.contentType)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tcase.want.body, string(body))
		})
	}
}

func TestApplication_pingHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	tests := []struct {
		name       string
		wantStatus int
	}{
		{name: "ping", wantStatus: 200},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/ping", nil)

			app.pingHandler(w, r)
			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}

func TestApplication_registerUserHadler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

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
				`{"login":"user1","password":"password1"}`,
			},
			wantStatus: 200,
		},
		{
			name: "bad content-type",
			request: request{
				"text/plaun",
				`{"login":"user1","password":"password1"}`,
			},
			wantStatus: 400,
		},
		{
			name: "empty pass",
			request: request{
				"application/json",
				`{"login":"user1"}`,
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
				`{"login":"user1","password":"password1"}`,
			},
			wantStatus: 409,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)

			app.registerUserHadler(w, r)
			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}

func TestApplication_loginUserHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	err := app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user1", PasswordHash: "password"})
	require.NoError(t, err)

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
				`{"login":"user1","password":"password"}`,
			},
			wantStatus: 200,
		},
		{
			name: "bad content-type",
			request: request{
				"text/plaun",
				`{"login":"user1","password":"password"}`,
			},
			wantStatus: 400,
		},
		{
			name: "bad body",
			request: request{
				"application/json",
				`{"login":"user1"}`,
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
			name: "bad password",
			request: request{
				"application/json",
				`{"login":"user1", "password":"user1"}`,
			},
			wantStatus: 401,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)

			app.loginUserHandler(w, r)
			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}

func TestApplication_addOrderHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user1", PasswordHash: "password"})
	err := app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user2", PasswordHash: "password"})
	require.NoError(t, err)

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
				`{"login":"user1","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 202,
		},
		{
			name: "exist",
			request: request{
				`{"login":"user1","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 200,
		},
		{
			name: "bad content-type",
			request: request{
				`{"login":"user1","password":"password"}`,
				"application/json",
				`12345678903`,
			},
			wantStatus: 400,
		},
		{
			name: "unautotize",
			request: request{
				`{"login":"user3","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 401,
		},
		{
			name: "write conflict",
			request: request{
				`{"login":"user2","password":"password"}`,
				"text/plain",
				`12345678903`,
			},
			wantStatus: 409,
		},
		{
			name: "bad order",
			request: request{
				`{"login":"user1","password":"password"}`,
				"text/plain",
				`12345678900`,
			},
			wantStatus: 422,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			// auth
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.login))
			r.Header.Set("Content-Type", "application/json")

			app.loginUserHandler(w, r)

			resp := w.Result()
			if resp.StatusCode == 401 {
				require.Equal(t, tcase.wantStatus, resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}
			defer resp.Body.Close()

			Cookie := w.Header().Get("Set-Cookie")

			w = httptest.NewRecorder()
			r = httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)
			r.Header.Set("Cookie", Cookie)

			middleware.CookieAuth(
				http.HandlerFunc(app.addOrderHandler)).
				ServeHTTP(w, r)
			resp = w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tcase.wantStatus, w.Code)
		})
	}
}

func TestApplication_getOrdersHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user1", PasswordHash: "password"})
	err := app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user2", PasswordHash: "password"})
	require.NoError(t, err)

	err = app.storage.WriteOrder(context.TODO(), model.OrderInfo{
		UserID: "user2", OrderID: 12345678903, Status: "NEW", UploadedAt: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)

	err = app.storage.WriteLoyalty(context.TODO(), []model.LoyaltyInfo{{UserID: "user2", OrderID: 12345678903,
		IsAccrual: true, Points: 505.05, UploadedAt: time.Date(2000, 12, 31, 1, 0, 0, 0, time.UTC),
	}})
	require.NoError(t, err)

	type request struct {
		login       string
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
				`{"login":"user1","password":"password"}`,
				"application/json",
			},
			wantStatus: 204,
			wantBody:   "",
		},
		{
			name: "OK",
			request: request{
				`{"login":"user2","password":"password"}`,
				"application/json",
			},
			wantStatus: 200,
			wantBody:   `[{"accrual":505.05, "number":"12345678903", "status":"NEW", "uploaded_at":"2000-12-31T00:00:00Z"}]`,
		},
		{
			name: "unautorize",
			request: request{
				`{"login":"user1","password":"-"}`,
				"application/json",
			},
			wantStatus: 401,
			wantBody:   "",
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			// auth
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.login))
			r.Header.Set("Content-Type", "application/json")

			app.loginUserHandler(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode == 401 {
				require.Equal(t, tcase.wantStatus, resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}

			Cookie := w.Header().Get("Set-Cookie")

			w = httptest.NewRecorder()
			r = httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Content-Type", tcase.request.contentType)
			r.Header.Set("Cookie", Cookie)

			middleware.CookieAuth(
				http.HandlerFunc(app.getOrdersHandler)).
				ServeHTTP(w, r)
			resp = w.Result()
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

func TestApplication_getBalanceHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	err := app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user1", PasswordHash: "password"})
	require.NoError(t, err)

	err = app.storage.WriteOrder(context.TODO(), model.OrderInfo{UserID: "user1", OrderID: 12345678903,
		Status: "PROCESSED", UploadedAt: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)

	err = app.storage.WriteLoyalty(context.TODO(), []model.LoyaltyInfo{
		{UserID: "user1", OrderID: 12345678903, IsAccrual: true, Points: 505.05,
			UploadedAt: time.Date(2000, 12, 31, 1, 0, 0, 0, time.UTC)},
		{UserID: "user1", OrderID: 12345678903, IsAccrual: false, Points: 100.0,
			UploadedAt: time.Date(2000, 12, 31, 2, 0, 0, 0, time.UTC)},
	})
	require.NoError(t, err)

	type request struct {
		login       string
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
				`{"login":"user1","password":"password"}`,
				"application/json",
			},
			wantStatus: 200,
			wantBody:   `{"current":405.05, "withdrawn":100}`,
		},
		{
			name: "unautorize",
			request: request{
				`{"login":"user1","password":"-"}`,
				"application/json",
			},
			wantStatus: 401,
			wantBody:   "",
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			// auth
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.login))
			r.Header.Set("Content-Type", "application/json")

			app.loginUserHandler(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode == 401 {
				require.Equal(t, tcase.wantStatus, resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}

			Cookie := w.Header().Get("Set-Cookie")

			w = httptest.NewRecorder()
			r = httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Content-Type", tcase.request.contentType)
			r.Header.Set("Cookie", Cookie)

			middleware.CookieAuth(
				http.HandlerFunc(app.getBalanceHandler)).
				ServeHTTP(w, r)
			resp = w.Result()
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

func TestApplication_withdrawHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	err := app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user1", PasswordHash: "password"})
	require.NoError(t, err)

	err = app.storage.WriteOrder(context.TODO(), model.OrderInfo{UserID: "user1", OrderID: 12345678903,
		Status: "PROCESSED", UploadedAt: time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)

	err = app.storage.WriteLoyalty(context.TODO(), []model.LoyaltyInfo{
		{UserID: "user1", OrderID: 12345678903, IsAccrual: true, Points: 505.05,
			UploadedAt: time.Date(2000, 12, 31, 1, 0, 0, 0, time.UTC)},
		{UserID: "user1", OrderID: 12345678903, IsAccrual: false, Points: 100.0,
			UploadedAt: time.Date(2000, 12, 31, 2, 0, 0, 0, time.UTC)},
	})
	require.NoError(t, err)

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
			name: "OK",
			request: request{
				`{"login":"user1","password":"password"}`,
				"application/json",
				`{"order":"12345678903", "sum":300}`,
			},
			wantStatus: 200,
		},
		{
			name: "unautorize",
			request: request{
				`{"login":"user1","password":"-"}`,
				"application/json",
				`{"order":"12345678903", "sum":500}`,
			},
			wantStatus: 401,
		},
		{
			name: "payment required",
			request: request{
				`{"login":"user1","password":"password"}`,
				"application/json",
				`{"order":"12345678903", "sum":500}`,
			},
			wantStatus: 402,
		},
		{
			name: "bad order",
			request: request{
				`{"login":"user1","password":"password"}`,
				"application/json",
				`{"order":"12345678900", "sum":1000}`,
			},
			wantStatus: 422,
		},
	}
	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			// auth
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.login))
			r.Header.Set("Content-Type", "application/json")

			app.loginUserHandler(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode == 401 {
				require.Equal(t, tcase.wantStatus, resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}

			Cookie := w.Header().Get("Set-Cookie")

			w = httptest.NewRecorder()
			r = httptest.NewRequest("POST", "/", strings.NewReader(tcase.request.body))
			r.Header.Set("Content-Type", tcase.request.contentType)
			r.Header.Set("Cookie", Cookie)

			middleware.CookieAuth(
				http.HandlerFunc(app.withdrawHandler)).
				ServeHTTP(w, r)
			resp = w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tcase.wantStatus, w.Code)

		})
	}
}

func TestApplication_getWithdrawalsHandler(t *testing.T) {

	app := newMocApplication(t)
	defer app.Close()

	err := app.storage.WriteUser(context.TODO(), model.UserInfo{UserID: "user1", PasswordHash: "password"})
	require.NoError(t, err)

	err = app.storage.WriteOrder(context.TODO(), model.OrderInfo{UserID: "user1", OrderID: 12345678903,
		Status: "PROCESSED", UploadedAt: time.Date(2000, 12, 31, 1, 0, 0, 0, time.UTC)})
	require.NoError(t, err)

	err = app.storage.WriteLoyalty(context.TODO(), []model.LoyaltyInfo{
		{UserID: "user1", OrderID: 12345678903, IsAccrual: true, Points: 505.05,
			UploadedAt: time.Date(2000, 12, 31, 1, 0, 0, 0, time.UTC)},
		{UserID: "user1", OrderID: 12345678903, IsAccrual: false, Points: 100.0,
			UploadedAt: time.Date(2000, 12, 31, 2, 0, 0, 0, time.UTC)},
	})
	require.NoError(t, err)

	type request struct {
		login       string
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
				`{"login":"user1","password":"password"}`,
				"application/json",
			},
			wantStatus: 200,
			wantBody:   `[{"order":"12345678903", "processed_at":"2000-12-31T02:00:00Z", "sum":100}]`,
		},
		{
			name: "unautorize",
			request: request{
				`{"login":"user1","password":"-"}`,
				"application/json",
			},
			wantStatus: 401,
			wantBody:   "",
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {

			// auth
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/",
				strings.NewReader(tcase.request.login))
			r.Header.Set("Content-Type", "application/json")

			app.loginUserHandler(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode == 401 {
				require.Equal(t, tcase.wantStatus, resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}

			Cookie := w.Header().Get("Set-Cookie")

			w = httptest.NewRecorder()
			r = httptest.NewRequest("GET", "/", nil)
			r.Header.Set("Content-Type", tcase.request.contentType)
			r.Header.Set("Cookie", Cookie)

			middleware.CookieAuth(
				http.HandlerFunc(app.getWithdrawalsHandler)).
				ServeHTTP(w, r)
			resp = w.Result()
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
