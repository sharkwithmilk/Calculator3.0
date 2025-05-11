package Agent

import (
	t "Calculator3.0/Pkg"
	pb "Calculator3.0/Proto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	ID     int
	Conn   *grpc.ClientConn
	Client pb.TaskServiceClient
	WG     *sync.WaitGroup
}

func (a *Agent) FetchTask() (*t.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := a.Client.FetchTask(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	return &t.Task{
		ID:            int(resp.Id),
		Arg1:          resp.Arg1,
		Arg2:          resp.Arg2,
		Operation:     resp.Operation,
		OperationTime: int(resp.OperationTime),
	}, nil
}

func getOperationTime(operation string) int {
	var envVar string
	switch operation {
	case "+":
		envVar = "TIME_ADDITION_MS"
	case "-":
		envVar = "TIME_SUBTRACTION_MS"
	case "*":
		envVar = "TIME_MULTIPLICATION_MS"
	case "/":
		envVar = "TIME_DIVISION_MS"
	}

	timeMs, err := strconv.Atoi(os.Getenv(envVar))
	if err != nil {
		return 100 // Default value
	}
	return timeMs
}

func (a *Agent) ProcessTask(task *t.Task) (float64, error) {
	time.Sleep(time.Duration(getOperationTime(task.Operation)) * time.Millisecond)
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2, nil
	case "-":
		return task.Arg1 - task.Arg2, nil
	case "*":
		return task.Arg1 * task.Arg2, nil
	case "/":
		if task.Arg2 != 0 {
			return task.Arg1 / task.Arg2, nil
		}
		return 0, fmt.Errorf("division by zero")
	}
	return 0, fmt.Errorf("unknown operation")
}

func (a *Agent) SubmitResult(taskID int, result float64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.TaskResult{
		Id:     int32(taskID),
		Result: result,
	}
	if err != nil {
		req.Error = err.Error()
	}

	_, submitErr := a.Client.SubmitResult(ctx, req)
	if submitErr != nil {
		log.Printf("[Агент %d] Ошибка отправки результата: %v", a.ID, submitErr)
	} else {
		log.Printf("[Агент %d] Результат отправлен: %d -> %f, Ошибка: %v", a.ID, taskID, result, err)
	}
}

func (a *Agent) Run() {
	defer a.WG.Done()
	for {
		task, err := a.FetchTask()
		if err != nil {
			log.Printf("[Агент %d] Нет задач, жду...: %v", a.ID, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("[Агент %d] Получена задача: %v", a.ID, task)
		result, calcErr := a.ProcessTask(task)
		a.SubmitResult(task.ID, result, calcErr)
	}
}

func NewAgent(id int, wg *sync.WaitGroup) *Agent {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("[Агент %d] Не удалось подключиться к gRPC серверу: %v", id, err)
	}
	client := pb.NewTaskServiceClient(conn)
	return &Agent{ID: id, Conn: conn, Client: client, WG: wg}
}
