package Tests

import (
	"Calculator3.0/Internal/User"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Calculator3.0/Internal/DataBase"
	"Calculator3.0/Internal/Orchestrator"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAddExpressionIntegration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	defer db.Close()

	mock.ExpectExec("INSERT INTO expressions").WithArgs(1, "2 + 3", "pending").WillReturnResult(sqlmock.NewResult(1, 1))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 1})
	tokenStr, _ := token.SignedString(User.JwtSecret)
	reqBody, _ := json.Marshal(map[string]string{"expression": "2 + 3"})
	req := httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	req.Header.Set("User-ID", "1")
	rr := httptest.NewRecorder()

	Orchestrator.AddExpressionHandler(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]int
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.Equal(t, 1, resp["id"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExpressionsIntegration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "status", "result"}).AddRow(1, "completed", 5.0)
	mock.ExpectQuery("SELECT id, status, result FROM expressions").WithArgs(1).WillReturnRows(rows)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 1})
	tokenStr, _ := token.SignedString(User.JwtSecret)
	req := httptest.NewRequest("GET", "/api/v1/expressions", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	req.Header.Set("User-ID", "1")
	rr := httptest.NewRecorder()

	Orchestrator.GetExpressionsHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	expressions := resp["expressions"].([]interface{})
	assert.Len(t, expressions, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExpressionByIDIntegration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "status", "result"}).AddRow(1, "completed", 5.0)
	mock.ExpectQuery("SELECT id, status, result FROM expressions").WithArgs(1, 1).WillReturnRows(rows)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": 1})
	tokenStr, _ := token.SignedString(User.JwtSecret)
	req := httptest.NewRequest("GET", "/api/v1/expressions/1", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	req.Header.Set("User-ID", "1")
	rr := httptest.NewRecorder()
	Orchestrator.GetExpressionByIDHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	expr := resp["expression"].(map[string]interface{})
	assert.Equal(t, float64(1), expr["id"])
	assert.Equal(t, "completed", expr["status"])
	assert.Equal(t, float64(5), expr["result"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
