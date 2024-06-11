package wait

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
	"github.com/stackitcloud/stackit-sdk-go/core/utils"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
)

// Used for testing cluster operations
type apiClientClusterMocked struct {
	getFails      bool
	name          string
	resourceState string
}

func (a *apiClientClusterMocked) GetClusterExecute(_ context.Context, _, _ string) (*skcf.Cluster, error) {
	if a.getFails {
		return nil, &oapierror.GenericOpenAPIError{
			StatusCode: http.StatusInternalServerError,
		}
	}
	rs := skcf.ClusterStatusState(a.resourceState)

	return &skcf.Cluster{
		Name: utils.Ptr("cluster"),
		Status: &skcf.ClusterStatus{
			Aggregated: &rs,
		},
	}, nil
}

func (a *apiClientClusterMocked) ListClustersExecute(_ context.Context, _ string) (*skcf.ListClustersResponse, error) {
	if a.getFails {
		return nil, &oapierror.GenericOpenAPIError{
			StatusCode: http.StatusInternalServerError,
		}
	}
	rs := skcf.ClusterStatusState(a.resourceState)
	return &skcf.ListClustersResponse{
		Items: &[]skcf.Cluster{
			{
				Name: utils.Ptr("cluster"),
				Status: &skcf.ClusterStatus{
					Aggregated: &rs,
				},
			},
		},
	}, nil
}

// Used for testing cluster operations
type apiClientProjectMocked struct {
	getFails      bool
	getNotFound   bool
	resourceState string
}

func (a *apiClientProjectMocked) GetServiceStatusExecute(_ context.Context, _ string) (*skcf.ProjectResponse, error) {
	if a.getFails {
		return nil, &oapierror.GenericOpenAPIError{
			StatusCode: http.StatusInternalServerError,
		}
	}
	if a.getNotFound {
		return nil, &oapierror.GenericOpenAPIError{
			StatusCode: http.StatusNotFound,
		}
	}
	rs := skcf.ProjectState(a.resourceState)
	return &skcf.ProjectResponse{
		ProjectId: utils.Ptr("pid"),
		State:     &rs,
	}, nil
}

func TestCreateOrUpdateClusterWaitHandler(t *testing.T) {
	tests := []struct {
		desc          string
		getFails      bool
		resourceState string
		wantErr       bool
		wantResp      bool
	}{
		{
			desc:          "create_succeeded",
			getFails:      false,
			resourceState: StateHealthy,
			wantErr:       false,
			wantResp:      true,
		},
		{
			desc:          "create_failed",
			getFails:      false,
			resourceState: StateFailed,
			wantErr:       true,
			wantResp:      true,
		},
		{
			desc:     "get_fails",
			getFails: true,
			wantErr:  true,
			wantResp: false,
		},
		{
			desc:          "timeout",
			getFails:      false,
			resourceState: "ANOTHER STATE",
			wantErr:       true,
			wantResp:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			name := "cluster"

			apiClient := &apiClientClusterMocked{
				getFails:      tt.getFails,
				name:          name,
				resourceState: tt.resourceState,
			}
			var wantRes *skcf.Cluster
			rs := skcf.ClusterStatusState(tt.resourceState)
			if tt.wantResp {
				wantRes = &skcf.Cluster{
					Name: &name,
					Status: &skcf.ClusterStatus{
						Aggregated: &rs,
					},
				}
			}

			handler := CreateOrUpdateClusterWaitHandler(context.Background(), apiClient, "", name)

			gotRes, err := handler.SetTimeout(10 * time.Millisecond).WaitWithContext(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("handler error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(gotRes, wantRes) {
				t.Fatalf("handler gotRes = %+v, want %+v", gotRes, wantRes)
			}
		})
	}
}

func TestCreateProjectWaitHandler(t *testing.T) {
	tests := []struct {
		desc          string
		getFails      bool
		resourceState string
		wantErr       bool
		wantResp      bool
	}{
		{
			desc:          "create_succeeded",
			getFails:      false,
			resourceState: StateCreated,
			wantErr:       false,
			wantResp:      true,
		},
		{
			desc:     "create_failed",
			getFails: false,
			wantErr:  true,
			wantResp: false,
		},
		{
			desc:     "get_fails",
			getFails: true,
			wantErr:  true,
			wantResp: false,
		},
		{
			desc:          "timeout",
			getFails:      false,
			resourceState: "ANOTHER STATE",
			wantErr:       true,
			wantResp:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			apiClient := &apiClientProjectMocked{
				getFails:      tt.getFails,
				resourceState: tt.resourceState,
			}
			var wantRes *skcf.ProjectResponse
			rs := skcf.ProjectState(tt.resourceState)
			if tt.wantResp {
				wantRes = &skcf.ProjectResponse{
					ProjectId: utils.Ptr("pid"),
					State:     &rs,
				}
			}

			handler := EnableServiceWaitHandler(context.Background(), apiClient, "")

			gotRes, err := handler.SetTimeout(10 * time.Millisecond).WaitWithContext(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("handler error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(gotRes, wantRes) {
				t.Fatalf("handler gotRes = %+v, want %+v", gotRes, wantRes)
			}
		})
	}
}

func TestDeleteProjectWaitHandler(t *testing.T) {
	tests := []struct {
		desc          string
		getFails      bool
		getNotFound   bool
		wantErr       bool
		resourceState string
	}{
		{
			desc:        "delete_succeeded",
			getFails:    false,
			getNotFound: true,
			wantErr:     false,
		},
		{
			desc:     "get_fails",
			getFails: true,
			wantErr:  true,
		},
		{
			desc:          "timeout",
			getFails:      false,
			wantErr:       true,
			resourceState: "ANOTHER STATE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			apiClient := &apiClientProjectMocked{
				getFails:      tt.getFails,
				getNotFound:   tt.getNotFound,
				resourceState: tt.resourceState,
			}

			handler := DisableServiceWaitHandler(context.Background(), apiClient, "")

			_, err := handler.SetTimeout(10 * time.Millisecond).WaitWithContext(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("handler error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
