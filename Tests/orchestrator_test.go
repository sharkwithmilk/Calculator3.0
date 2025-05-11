package Tests

import (
	"Calculator3.0/Internal/DataBase"
	"Calculator3.0/Internal/Orchestrator"
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddExpressionHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания mock: %v", err)
	}
	defer db.Close()
	DataBase.DB = db

	mock.ExpectExec("INSERT INTO expressions \\(user_id, expression, status\\) VALUES \\(\\?, \\?, \\?\\)").
		WithArgs(1, "2 + 3", "pending").
		WillReturnResult(sqlmock.NewResult(1, 1))

	reqBody := []byte(`{"expression": "2 + 3"}`)
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(reqBody))
	req.Header.Set("User-ID", "1")
	rr := httptest.NewRecorder()
	Orchestrator.AddExpressionHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Ожидался код 201, получен %d", rr.Code)
	}
	var resp map[string]int
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Errorf("Ошибка разбора ответа: %v", err)
	}
	if _, ok := resp["id"]; !ok {
		t.Errorf("ID отсутствует в ответе")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Ожидания mock не выполнены: %v", err)
	}
}

func TestGetExpressionsHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания mock: %v", err)
	}
	defer db.Close()
	DataBase.DB = db

	rows := sqlmock.NewRows([]string{"id", "status", "result"}).
		AddRow(1, "completed", 5.0)
	mock.ExpectQuery("SELECT id, status, result FROM expressions WHERE user_id = \\?").
		WithArgs(1).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/v1/expressions", nil)
	req.Header.Set("User-ID", "1")
	rr := httptest.NewRecorder()
	Orchestrator.GetExpressionsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался код 200, получен %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Ожидания mock не выполнены: %v", err)
	}
}

func TestGetExpressionByIDHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка создания mock: %v", err)
	}
	defer db.Close()
	DataBase.DB = db

	mock.ExpectQuery("SELECT id, status, result FROM expressions WHERE id = \\? AND user_id = \\?").
		WithArgs(999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "result"}))

	req, _ := http.NewRequest("GET", "/api/v1/expressions/999", nil)
	req.Header.Set("User-ID", "1")
	rr := httptest.NewRecorder()
	Orchestrator.GetExpressionByIDHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Ожидался код 404, получен %d", rr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Ожидания mock не выполнены: %v", err)
	}
}
