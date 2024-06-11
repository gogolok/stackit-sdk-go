package wait

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
	"github.com/stackitcloud/stackit-sdk-go/core/wait"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
)

const (
	StateHealthy   = "STATE_HEALTHY"
	StateFailed    = "STATE_FAILED"
	StateDeleting  = "STATE_DELETING"
	StateCreated   = "STATE_CREATED"
	StateUnhealthy = "STATE_UNHEALTHY"
)

type APIClientProjectInterface interface {
	GetServiceStatusExecute(ctx context.Context, projectId string) (*skcf.ProjectResponse, error)
}

type APIClientClusterInterface interface {
	GetClusterExecute(ctx context.Context, projectId, name string) (*skcf.Cluster, error)
	ListClustersExecute(ctx context.Context, projectId string) (*skcf.ListClustersResponse, error)
}

// CreateOrUpdateClusterWaitHandler will wait for cluster creation or update
func CreateOrUpdateClusterWaitHandler(ctx context.Context, a APIClientClusterInterface, projectId, name string) *wait.AsyncActionHandler[skcf.Cluster] {
	handler := wait.New(func() (waitFinished bool, response *skcf.Cluster, err error) {
		s, err := a.GetClusterExecute(ctx, projectId, name)
		if err != nil {
			return false, nil, err
		}
		state := *s.Status.Aggregated

		if state == StateHealthy {
			return true, s, nil
		}

		if state == StateFailed {
			return true, s, fmt.Errorf("create failed")
		}

		return false, nil, nil
	})
	handler.SetTimeout(45 * time.Minute)
	return handler
}

// DeleteClusterWaitHandler will wait for cluster deletion
func DeleteClusterWaitHandler(ctx context.Context, a APIClientClusterInterface, projectId, name string) *wait.AsyncActionHandler[skcf.ListClustersResponse] {
	handler := wait.New(func() (waitFinished bool, response *skcf.ListClustersResponse, err error) {
		s, err := a.ListClustersExecute(ctx, projectId)
		if err != nil {
			return false, nil, err
		}
		items := *s.Items
		for i := range items {
			n := items[i].Name
			if n != nil && *n == name {
				return false, nil, nil
			}
		}
		return true, s, nil
	})
	handler.SetTimeout(45 * time.Minute)
	return handler
}

// EnableServiceWaitHandler will wait for service enablement
func EnableServiceWaitHandler(ctx context.Context, a APIClientProjectInterface, projectId string) *wait.AsyncActionHandler[skcf.ProjectResponse] {
	handler := wait.New(func() (waitFinished bool, response *skcf.ProjectResponse, err error) {
		s, err := a.GetServiceStatusExecute(ctx, projectId)
		if err != nil {
			return false, nil, err
		}
		state := *s.State
		switch state {
		case StateDeleting, StateFailed:
			return false, nil, fmt.Errorf("received state: %s for project Id: %s", state, projectId)
		case StateCreated:
			return true, s, nil
		}
		return false, nil, nil
	})
	handler.SetTimeout(15 * time.Minute)
	return handler
}

// DisableServiceWaitHandler will wait for service disablement
func DisableServiceWaitHandler(ctx context.Context, a APIClientProjectInterface, projectId string) *wait.AsyncActionHandler[struct{}] {
	handler := wait.New(func() (waitFinished bool, response *struct{}, err error) {
		_, err = a.GetServiceStatusExecute(ctx, projectId)
		if err == nil {
			return false, nil, nil
		}
		oapiErr, ok := err.(*oapierror.GenericOpenAPIError) //nolint:errorlint //complaining that error.As should be used to catch wrapped errors, but this error should not be wrapped
		if !ok {
			return false, nil, fmt.Errorf("could not convert error to oapierror.GenericOpenAPIError in delete wait.AsyncHandler, %w", err)
		}
		if oapiErr.StatusCode == http.StatusNotFound || oapiErr.StatusCode == http.StatusForbidden {
			return true, nil, nil
		}
		return false, nil, err
	})
	handler.SetTimeout(15 * time.Minute)
	return handler
}
