package Orchestrator

import (
	"Calculator3.0/Internal/DataBase"
	t "Calculator3.0/Pkg"
	pb "Calculator3.0/Proto"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Expression struct {
	ID         int     `json:"id"`
	Status     string  `json:"status"`
	Result     float64 `json:"result,omitempty"`
	Tokens     []t.Token
	UserID     int    //поле для пользовател
	Expression string `json:"expression"` //исход выражение
}

var (
	expressions = make(map[int]*Expression)
	tasks       = make(chan t.Task, 100)
	mutex       sync.Mutex
	exprID      = 1
	taskID      = 1
)

type TaskServer struct {
	pb.UnimplementedTaskServiceServer
}

func (s *TaskServer) FetchTask(ctx context.Context, _ *pb.Empty) (*pb.Task, error) {
	select {
	case task := <-tasks:
		return &pb.Task{
			Id:            int32(task.ID),
			Arg1:          task.Arg1,
			Arg2:          task.Arg2,
			Operation:     task.Operation,
			OperationTime: int32(task.OperationTime),
		}, nil
	default:
		return nil, fmt.Errorf("no task available")
	}
}

func (s *TaskServer) SubmitResult(ctx context.Context, res *pb.TaskResult) (*pb.Empty, error) {
	var exprID int
	DataBase.DB.QueryRow("SELECT id FROM expressions WHERE id = ? LIMIT 1", res.Id).Scan(&exprID)

	if res.Error != "" {
		DataBase.DB.Exec("UPDATE expressions SET status = 'error' WHERE id = ?", exprID)
		fmt.Printf("[Оркестратор] Ошибка при вычислении выражения ID %d: %s\n", res.Id, res.Error)
	} else {
		DataBase.DB.Exec("UPDATE expressions SET status = 'completed', result = ? WHERE id = ?", res.Result, exprID)
		fmt.Printf("[Оркестратор] Выражение ID %d вычислено: %f\n", res.Id, res.Result)
	}
	return &pb.Empty{}, nil
}

func StartGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &TaskServer{})
	fmt.Println("gRPC server running on :50051")
	if err := s.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	go StartGRPCServer() // Start gRPC server in a goroutine
	go CreateTasks()

	http.HandleFunc("/api/v1/calculate", AddExpressionHandler)
	http.HandleFunc("/api/v1/expressions", GetExpressionsHandler)
	http.HandleFunc("/api/v1/expressions/", GetExpressionByIDHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Оркестратор запущен на порту", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("HTTP server failed: %v\n", err)
	}
}
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("User-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	//Сохр выражения
	result, err := DataBase.DB.Exec("INSERT INTO expressions (user_id, expression, status) VALUES (?, ?, ?)", userID, req.Expression, "pending")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	id, _ := result.LastInsertId()

	fmt.Printf("[Оркестратор] Добавлено выражение ID %d: %s\n", id, req.Expression)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": int(id)})
}

func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("User-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := DataBase.DB.Query("SELECT id, status, result FROM expressions WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []struct {
		ID     int     `json:"id"`
		Status string  `json:"status"`
		Result float64 `json:"result,omitempty"`
	}
	for rows.Next() {
		var expr struct {
			ID     int     `json:"id"`
			Status string  `json:"status"`
			Result float64 `json:"result,omitempty"`
		}
		var result sql.NullFloat64
		if err := rows.Scan(&expr.ID, &expr.Status, &result); err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		if result.Valid {
			expr.Result = result.Float64
		}
		list = append(list, expr)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": list})
}

func GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("User-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/api/v1/expressions/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var expr struct {
		ID     int     `json:"id"`
		Status string  `json:"status"`
		Result float64 `json:"result,omitempty"`
	}
	var result sql.NullFloat64
	err = DataBase.DB.QueryRow("SELECT id, status, result FROM expressions WHERE id = ? AND user_id = ?", id, userID).Scan(&expr.ID, &expr.Status, &result)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Expression not found", http.StatusNotFound)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}
	if result.Valid {
		expr.Result = result.Float64
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		select {
		case task := <-tasks:
			json.NewEncoder(w).Encode(map[string]t.Task{"task": task})
		default:
			http.Error(w, "No task available", http.StatusNotFound)
		}
	} else if r.Method == http.MethodPost {
		var res struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
			Error  string  `json:"error,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
			return
		}

		mutex.Lock()
		defer mutex.Unlock()

		for _, expr := range expressions {
			if expr.ID == res.ID {
				if res.Error != "" {
					expr.Status = "error"
					fmt.Printf("[Оркестратор] Ошибка при вычислении выражения ID %d: %s\n", res.ID, res.Error)
				} else {
					expr.Status = "completed"
					expr.Result = res.Result
					fmt.Printf("[Оркестратор] Выражение ID %d вычислено: %f\n", res.ID, res.Result)
				}
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		http.Error(w, "Task not found", http.StatusNotFound)
	}
}

func CreateTasks() {
	for {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		rows, _ := DataBase.DB.Query("SELECT id, expression FROM expressions WHERE status = 'pending'")
		defer rows.Close()

		for rows.Next() {
			var id int
			var expression string
			rows.Scan(&id, &expression)
			tokens := t.Tokenize(expression)
			root := (&t.Parser{Tokens: tokens}).ParseExpression()
			task := t.Task{
				ID:            taskID,
				Arg1:          root.Left.Value,
				Arg2:          root.Right.Value,
				Operation:     root.Operator,
				OperationTime: rand.Intn(500),
			}
			taskID++
			tasks <- task
			DataBase.DB.Exec("UPDATE expressions SET status = 'in_progress' WHERE id = ?", id)
			fmt.Printf("[Оркестратор] Создана задача ID %d для выражения %d\n", task.ID, id)
		}
	}
}
