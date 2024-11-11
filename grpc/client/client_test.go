//go:build unit

package client

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func Test_GoFlowService_Push(t *testing.T) {
	t.Run("Successfully pushes task", func(t *testing.T) {
		// Arrange
		mockClient := new(mockGoFlowClient)
		mockLogger := new(log.TestifyMock)
		service := &GoFlowService{mockClient, time.Second, mockLogger}

		taskType := "example-task"
		payload := "example-payload"
		expectedID := "12345"

		mockClient.On("PushTask", mock.Anything, &pb.PushTaskRequest{TaskType: taskType, Payload: payload}).
			Once().
			Return(&pb.PushTaskReply{Id: expectedID}, nil)

		// Act
		taskID, err := service.Push(taskType, payload)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedID, taskID)
		mockClient.AssertExpectations(t)
	})

	t.Run("Returns error if push task fails", func(t *testing.T) {
		// Arrange
		mockClient := new(mockGoFlowClient)
		mockLogger := new(log.TestifyMock)
		service := &GoFlowService{mockClient, time.Second, mockLogger}

		taskType := "example-task"
		payload := "example-payload"
		expectedError := errors.New("failed to push task")

		mockClient.On("PushTask", mock.Anything, &pb.PushTaskRequest{TaskType: taskType, Payload: payload}).
			Once().
			Return(nil, expectedError)

		// Act
		taskID, err := service.Push(taskType, payload)

		// Assert
		assert.Empty(t, taskID)
		assert.EqualError(t, err, "failed to push task: failed to push task")
		mockClient.AssertExpectations(t)
	})
}

func Test_GoFlowService_Get(t *testing.T) {
	t.Run("Successfully retrieves task result", func(t *testing.T) {
		// Arrange
		mockClient := new(mockGoFlowClient)
		mockLogger := new(log.TestifyMock)
		service := &GoFlowService{mockClient, time.Second, mockLogger}

		taskID := "12345"
		expectedResult := "task result"

		mockClient.On("GetResult", mock.Anything, &pb.GetResultRequest{TaskID: taskID}).
			Once().
			Return(&pb.GetResultReply{Result: expectedResult}, nil)

		// Act
		result, err := service.Get(taskID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Returns error if get result fails", func(t *testing.T) {
		// Arrange
		mockClient := new(mockGoFlowClient)
		mockLogger := new(log.TestifyMock)
		service := &GoFlowService{mockClient, time.Second, mockLogger}

		taskID := "12345"
		expectedError := errors.New("result not found")

		mockClient.On("GetResult", mock.Anything, &pb.GetResultRequest{TaskID: taskID}).
			Once().
			Return(nil, expectedError)

		// Act
		result, err := service.Get(taskID)

		// Assert
		assert.Empty(t, result)
		assert.EqualError(t, err, "could not get result for taskID '12345': result not found")
		mockClient.AssertExpectations(t)
	})

	t.Run("Returns error on timeout", func(t *testing.T) {
		// Arrange
		mockClient := new(mockGoFlowClient)
		mockLogger := new(log.TestifyMock)
		service := &GoFlowService{mockClient, time.Second, mockLogger}

		taskID := "12345"
		expectedError := context.DeadlineExceeded

		mockClient.On("GetResult", mock.Anything, &pb.GetResultRequest{TaskID: taskID}).
			Once().
			After(service.requestTimeout*10).
			Return(nil, expectedError)

		// Act
		result, err := service.Get(taskID)

		// Assert
		assert.Empty(t, result)
		assert.EqualError(t, err, "could not get result for taskID '12345': context deadline exceeded")
		mockClient.AssertExpectations(t)
	})
}

type mockGoFlowClient struct {
	mock.Mock
}

func (m *mockGoFlowClient) PushTask(ctx context.Context, req *pb.PushTaskRequest, _ ...grpc.CallOption) (*pb.PushTaskReply, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pb.PushTaskReply), args.Error(1)
}

func (m *mockGoFlowClient) GetResult(ctx context.Context, req *pb.GetResultRequest, _ ...grpc.CallOption) (*pb.GetResultReply, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pb.GetResultReply), args.Error(1)
}
