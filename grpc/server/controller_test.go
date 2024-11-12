//go:build unit

package server

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	pb "github.com/jamesTait-jt/goflow/grpc/proto"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_GoFlowServiceController_PushTask(t *testing.T) {
	t.Run("Logs the request, pushes the task to GoFlow and returns the task ID", func(t *testing.T) {
		// Arrange
		svc := new(mockGoFlowService)
		logger := new(log.TestifyMock)

		controller := NewGoFlowServiceController(svc, logger)

		ctx := context.Background()
		req := &pb.PushTaskRequest{
			TaskType: "task-type",
			Payload:  "12345",
		}

		shouldLog := "Received push task: [task-type] [12345]"
		logger.On("Info", shouldLog).Once()

		taskID := "task-id"
		svc.On("PushTask", req.TaskType, req.Payload).Once().Return(taskID, nil)

		expectedReply := &pb.PushTaskReply{Id: taskID}

		// Act
		resp, err := controller.PushTask(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedReply, resp)

		svc.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if the push failed", func(t *testing.T) {
		// Arrange
		svc := new(mockGoFlowService)
		logger := new(log.TestifyMock)

		controller := NewGoFlowServiceController(svc, logger)

		ctx := context.Background()
		req := &pb.PushTaskRequest{
			TaskType: "task-type",
			Payload:  "12345",
		}

		shouldLog := "Received push task: [task-type] [12345]"
		logger.On("Info", shouldLog).Once()

		pushTaskErr := errors.New("couldn't push task")
		svc.On("PushTask", req.TaskType, req.Payload).Once().Return("", pushTaskErr)

		// Act
		resp, err := controller.PushTask(ctx, req)

		// Assert
		assert.EqualError(t, err, pushTaskErr.Error())
		assert.Nil(t, resp)

		svc.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

func Test_GoFlowServiceController_GetResult(t *testing.T) {
	t.Run("Logs the request, gets the result from GoFlow and returns the result", func(t *testing.T) {
		// Arrange
		svc := new(mockGoFlowService)
		logger := new(log.TestifyMock)

		controller := NewGoFlowServiceController(svc, logger)

		ctx := context.Background()
		req := &pb.GetResultRequest{
			TaskID: "task-id",
		}

		shouldLog := "Received get result: [task-id]"
		logger.On("Info", shouldLog).Once()

		result := task.Result{}
		svc.On("GetResult", req.TaskID).Once().Return(result, true, nil)

		parsedResult, _ := json.Marshal(result)
		expectedReply := &pb.GetResultReply{Result: string(parsedResult)}

		// Act
		resp, err := controller.GetResult(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedReply, resp)

		svc.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if GetResult returns an error", func(t *testing.T) {
		// Arrange
		svc := new(mockGoFlowService)
		logger := new(log.TestifyMock)

		controller := NewGoFlowServiceController(svc, logger)

		ctx := context.Background()
		req := &pb.GetResultRequest{
			TaskID: "failing-task-id",
		}

		shouldLog := "Received get result: [failing-task-id]"
		logger.On("Info", shouldLog).Once()

		getResultErr := errors.New("couldnt get result")
		svc.On("GetResult", req.TaskID).Once().Return(task.Result{}, false, getResultErr)

		// Act
		resp, err := controller.GetResult(ctx, req)

		// Assert
		assert.EqualError(t, err, getResultErr.Error())
		assert.Nil(t, resp)

		svc.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if the task is not complete or does not exist", func(t *testing.T) {
		// Arrange
		svc := new(mockGoFlowService)
		logger := new(log.TestifyMock)

		controller := NewGoFlowServiceController(svc, logger)

		ctx := context.Background()
		req := &pb.GetResultRequest{
			TaskID: "nonexistent-task-id",
		}

		shouldLog := "Received get result: [nonexistent-task-id]"
		logger.On("Info", shouldLog).Once()

		svc.On("GetResult", req.TaskID).Once().Return(task.Result{}, false, nil)

		// Act
		resp, err := controller.GetResult(ctx, req)

		// Assert
		assert.EqualError(t, err, "task not complete or didnt exist")
		assert.Nil(t, resp)

		svc.AssertExpectations(t)
		logger.AssertExpectations(t)
	})

	t.Run("Returns an error if marshaling the result fails", func(t *testing.T) {
		// Arrange
		svc := new(mockGoFlowService)
		logger := new(log.TestifyMock)

		controller := NewGoFlowServiceController(svc, logger)

		ctx := context.Background()
		req := &pb.GetResultRequest{
			TaskID: "task-id",
		}

		shouldLog := "Received get result: [task-id]"
		logger.On("Info", shouldLog).Once()

		result := task.Result{
			Payload: make(chan bool),
		}
		svc.On("GetResult", req.TaskID).Once().Return(result, true, nil)

		// Act
		resp, err := controller.GetResult(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal result")
		assert.Nil(t, resp)

		svc.AssertExpectations(t)
		logger.AssertExpectations(t)
	})
}

type mockGoFlowService struct {
	mock.Mock
}

func (m *mockGoFlowService) PushTask(taskType string, payload any) (string, error) {
	args := m.Called(taskType, payload)
	return args.String(0), args.Error(1)
}

func (m *mockGoFlowService) GetResult(taskID string) (task.Result, bool, error) {
	args := m.Called(taskID)
	return args.Get(0).(task.Result), args.Bool(1), args.Error(2)
}
