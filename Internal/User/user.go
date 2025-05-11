package User

import (
	"Calculator3.0/Internal/DataBase"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

var JwtSecret = []byte("pesyn")

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	var passwordHash string
	var userID int
	err := DataBase.DB.QueryRow("SELECT id, password_hash FROM users WHERE login = ?", req.Login).Scan(&userID, &passwordHash)
	if err != nil {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	//Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	//Ген JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), //Токен 24 часа
	})
	tokenString, err := token.SignedString(JwtSecret)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	//Проверка сущ пользователя
	var exists int
	err := DataBase.DB.QueryRow("SELECT COUNT(*) FROM users WHERE login = ?", req.Login).Scan(&exists)
	if err != nil || exists > 0 {
		http.Error(w, "Пользователь уже существует", http.StatusConflict)
		return
	}

	//Хеш пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	//Сохр пользователя
	_, err = DataBase.DB.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)", req.Login, hash)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Регистрация успешна"})
}
