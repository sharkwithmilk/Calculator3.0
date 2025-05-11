package Middleware

import (
	"Calculator3.0/Internal/User"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strconv"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return User.JwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Ошибка токена", http.StatusUnauthorized)
			return
		}

		userID := int(claims["user_id"].(float64))
		r.Header.Set("User-ID", strconv.Itoa(userID))
		next(w, r)
	}
}
