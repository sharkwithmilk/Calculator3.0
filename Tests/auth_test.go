package Tests

import (
	"Calculator3.0/Internal/Middleware"
	"Calculator3.0/Internal/User"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareValidToken(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 1})
	tokenStr, _ := token.SignedString(User.JwtSecret)
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler := Middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался код 200, получен %d", rr.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rr := httptest.NewRecorder()
	handler := Middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {})
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Ожидался код 401, получен %d", rr.Code)
	}
}
