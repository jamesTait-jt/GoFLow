package log

import "github.com/stretchr/testify/mock"

type TestifyMock struct {
	mock.Mock
}

func (m *TestifyMock) Info(msg string) {
	m.Called(msg)
}

func (m *TestifyMock) Success(msg string) {
	m.Called(msg)
}

func (m *TestifyMock) Warn(msg string) {
	m.Called(msg)
}

func (m *TestifyMock) Error(msg string) {
	m.Called(msg)
}

func (m *TestifyMock) Fatal(msg string) {
	m.Called(msg)
}

func (m *TestifyMock) Waiting(msg string) func(doneMsg string, success bool) {
	m.Called(msg)

	return func(doneMsg string, success bool) {
		m.Called(doneMsg, success)
	}
}
