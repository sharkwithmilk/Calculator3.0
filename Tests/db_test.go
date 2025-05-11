package Tests

import (
	"Calculator3.0/Internal/DataBase"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestInsertUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
	_, err := DataBase.DB.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)", "test", "hash")
	if err != nil {
		t.Errorf("Ошибка при добавлении пользователя: %v", err)
	}
}

func TestUpdateExpressionStatus(t *testing.T) {
	db, mock, _ := sqlmock.New()
	DataBase.DB = db
	mock.ExpectExec("UPDATE expressions").WithArgs("completed", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	_, err := DataBase.DB.Exec("UPDATE expressions SET status = ? WHERE id = ?", "completed", 1)
	if err != nil {
		t.Errorf("Ошибка при обновлении статуса: %v", err)
	}
}
