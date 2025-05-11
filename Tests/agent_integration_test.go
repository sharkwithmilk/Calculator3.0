package Tests

import (
	"context"
	"net"
	"sync"
	"testing"

	"Calculator3.0/Internal/Agent"
	"Calculator3.0/Proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func setupGRPCServer() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterTaskServiceServer(s, &mockTaskServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

type mockTaskServer struct {
	proto.UnimplementedTaskServiceServer
}

func (m *mockTaskServer) FetchTask(ctx context.Context, empty *proto.Empty) (*proto.Task, error) {
	return &proto.Task{Id: 1, Arg1: 2, Arg2: 3, Operation: "+"}, nil
}

func (m *mockTaskServer) SubmitResult(ctx context.Context, result *proto.TaskResult) (*proto.Empty, error) {
	return &proto.Empty{}, nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestAgentFetchTaskIntegration(t *testing.T) {
	setupGRPCServer()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := proto.NewTaskServiceClient(conn)
	agent := &Agent.Agent{ID: 1, Client: client, WG: &sync.WaitGroup{}}

	task, err := agent.FetchTask()
	assert.NoError(t, err)
	assert.Equal(t, 1, task.ID)
	assert.Equal(t, 2.0, task.Arg1)
	assert.Equal(t, 3.0, task.Arg2)
	assert.Equal(t, "+", task.Operation)
}

func TestAgentSubmitResultIntegration(t *testing.T) {
	setupGRPCServer()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()

	client := proto.NewTaskServiceClient(conn)
	agent := &Agent.Agent{ID: 1, Client: client, WG: &sync.WaitGroup{}}

	agent.SubmitResult(1, 5.0, nil)
}
