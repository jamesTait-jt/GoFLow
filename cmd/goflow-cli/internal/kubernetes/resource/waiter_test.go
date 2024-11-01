package resource

import (
	"context"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type mockWatchable struct {
	mock.Mock
}

// Watch is the mock implementation of the Watch method
func (m *mockWatchable) Watch(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
	args := m.Called(ctx, options)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(watch.Interface), args.Error(1)
}
