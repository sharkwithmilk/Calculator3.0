package Tests

import (
	"Calculator3.0/Internal/Agent"
	y "Calculator3.0/Pkg"
	"Calculator3.0/Proto"
	"context"
	"google.golang.org/grpc"
	"sync"
	"testing"
)

type mockTaskServiceClient struct {
	fetchTaskFunc    func(ctx context.Context, in *proto.Empty, opts ...grpc.CallOption) (*proto.Task, error)
	submitResultFunc func(ctx context.Context, in *proto.TaskResult, opts ...grpc.CallOption) (*proto.Empty, error)
}

func (m *mockTaskServiceClient) FetchTask(ctx context.Context, in *proto.Empty, opts ...grpc.CallOption) (*proto.Task, error) {
	return m.fetchTaskFunc(ctx, in, opts...)
}

func (m *mockTaskServiceClient) SubmitResult(ctx context.Context, in *proto.TaskResult, opts ...grpc.CallOption) (*proto.Empty, error) {
	return m.submitResultFunc(ctx, in, opts...)
}

func TestFetchTask(t *testing.T) {
	mockClient := &mockTaskServiceClient{
		fetchTaskFunc: func(ctx context.Context, in *proto.Empty, opts ...grpc.CallOption) (*proto.Task, error) {
			return &proto.Task{Id: 1, Arg1: 2, Arg2: 3, Operation: "+"}, nil
		},
	}
	agent := &Agent.Agent{ID: 1, Client: mockClient, WG: &sync.WaitGroup{}}
	task, err := agent.FetchTask()
	if err != nil || task.ID != 1 || task.Arg1 != 2 || task.Arg2 != 3 || task.Operation != "+" {
		t.Errorf("Ожидалась задача {ID: 1, Arg1: 2, Arg2: 3, Op: +}, получено: %+v, ошибка: %v", task, err)
	}
}

func TestProcessTask(t *testing.T) {
	agent := &Agent.Agent{ID: 1, WG: &sync.WaitGroup{}}
	tests := []struct {
		task     *y.Task
		expected float64
		err      bool
	}{
		{task: &y.Task{Arg1: 2, Arg2: 3, Operation: "+"}, expected: 5, err: false},
		{task: &y.Task{Arg1: 10, Arg2: 0, Operation: "/"}, expected: 0, err: true},
	}
	for _, tt := range tests {
		result, err := agent.ProcessTask(tt.task)
		if (err != nil) != tt.err || (!tt.err && result != tt.expected) {
			t.Errorf("Для %+v ожидалось %f (ошибка: %v), получено %f (ошибка: %v)", tt.task, tt.expected, tt.err, result, err)
		}
	}
}

func TestSubmitResult(t *testing.T) {
	var submitted *proto.TaskResult
	mockClient := &mockTaskServiceClient{
		submitResultFunc: func(ctx context.Context, in *proto.TaskResult, opts ...grpc.CallOption) (*proto.Empty, error) {
			submitted = in
			return &proto.Empty{}, nil
		},
	}
	agent := &Agent.Agent{ID: 1, Client: mockClient, WG: &sync.WaitGroup{}}
	agent.SubmitResult(1, 5, nil)
	if submitted.Id != 1 || submitted.Result != 5 || submitted.Error != "" {
		t.Errorf("Ожидалось {Id: 1, Result: 5, Error: ''}, получено: %+v", submitted)
	}
}
