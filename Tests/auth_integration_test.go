package Tests

import (
	"bytes"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"

	"Calculator3.0/Internal/DataBase"
	"Calculator3.0/Internal/User"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUserIntegration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка инициализации sqlmock: %v", err)
	}
	DataBase.DB = db
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE login = ?").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec("INSERT INTO users").
		WithArgs("testuser", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	reqBody, _ := json.Marshal(map[string]string{
		"login":    "testuser",
		"password": "password",
	})
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	User.RegisterHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Ожидался код ответа 200 OK")
	assert.Contains(t, rr.Body.String(), "Регистрация успешна", "Ожидалось сообщение об успешной регистрации")
	assert.NoError(t, mock.ExpectationsWereMet(), "Ожидания мока не выполнены")
}

func TestLoginUserIntegration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	defer db.Close()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, password_hash FROM users").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow(1, string(hashedPassword)))

	reqBody, _ := json.Marshal(map[string]string{"login": "testuser", "password": "password"})
	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	User.LoginHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterDuplicateUserIntegration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	reqBody, _ := json.Marshal(map[string]string{"login": "testuser", "password": "password"})
	req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	User.RegisterHandler(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "Пользователь уже существует")
	assert.NoError(t, mock.ExpectationsWereMet())
}
