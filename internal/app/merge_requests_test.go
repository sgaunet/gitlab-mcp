package app

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func newTestApp(mockClient *MockGitLabClient) *App {
	app := NewWithClient("token", "https://gitlab.com/", mockClient)
	app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	return app
}

func TestApp_ListProjectMergeRequests(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *ListMergeRequestsOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService)
		want    []MergeRequest
		wantErr bool
	}{
		{
			name: "successful list with default options",
			opts: &ListMergeRequestsOptions{State: "opened", Limit: 100},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectMergeRequestsOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				mrs.On("ListProjectMergeRequests", int64(123), expectedOpts).Return(
					[]*gitlab.MergeRequest{
						{
							BasicMergeRequest: gitlab.BasicMergeRequest{
								ID:                  1,
								IID:                 10,
								Title:               "Test MR",
								Description:         "Test Description",
								State:               "opened",
								SourceBranch:        "feature",
								TargetBranch:        "main",
								Labels:              gitlab.Labels{"bug"},
								WebURL:              "https://gitlab.com/test/project/-/merge_requests/10",
								DetailedMergeStatus: "mergeable",
								Draft:               false,
								Author:              &gitlab.BasicUser{ID: 1, Username: "user1", Name: "User One"},
								CreatedAt:           &testTime,
								UpdatedAt:           &testTime,
							},
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []MergeRequest{
				{
					ID:           1,
					IID:          10,
					Title:        "Test MR",
					Description:  "Test Description",
					State:        "opened",
					SourceBranch: "feature",
					TargetBranch: "main",
					Labels:       []string{"bug"},
					WebURL:       "https://gitlab.com/test/project/-/merge_requests/10",
					MergeStatus:  "mergeable",
					Draft:        false,
					Author:       map[string]any{"id": int64(1), "username": "user1", "name": "User One"},
					CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
					UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				},
			},
		},
		{
			name: "successful list with filters",
			opts: &ListMergeRequestsOptions{State: "merged", Labels: []string{"bug"}, Author: "dev1", Search: "fix", Limit: 50},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"bug"})
				expectedOpts := &gitlab.ListProjectMergeRequestsOptions{
					State:          gitlab.Ptr("merged"),
					Labels:         &expectedLabels,
					AuthorUsername: gitlab.Ptr("dev1"),
					Search:         gitlab.Ptr("fix"),
					ListOptions:    gitlab.ListOptions{PerPage: 50, Page: 1},
				}

				mrs.On("ListProjectMergeRequests", int64(123), expectedOpts).Return(
					[]*gitlab.MergeRequest{},
					&gitlab.Response{}, nil,
				)
			},
			want: []MergeRequest{},
		},
		{
			name: "project not found",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("project not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "API error on list",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectMergeRequestsOptions{
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				mrs.On("ListProjectMergeRequests", int64(123), expectedOpts).Return(
					([]*gitlab.MergeRequest)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			wantErr: true,
		},
		{
			name: "empty results",
			opts: &ListMergeRequestsOptions{State: "opened", Limit: 100},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectMergeRequestsOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				mrs.On("ListProjectMergeRequests", int64(123), expectedOpts).Return(
					[]*gitlab.MergeRequest{},
					&gitlab.Response{}, nil,
				)
			},
			want: []MergeRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockMRs := &MockMergeRequestsService{}

			tt.setup(mockClient, mockProjects, mockMRs)

			app := newTestApp(mockClient)
			result, err := app.ListProjectMergeRequests("test/project", tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_CreateProjectMergeRequest(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *CreateMergeRequestOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService)
		want    *MergeRequest
		wantErr bool
	}{
		{
			name: "successful create with minimal options",
			opts: &CreateMergeRequestOptions{
				Title:        "New MR",
				SourceBranch: "feature",
				TargetBranch: "main",
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateMergeRequestOptions{
					Title:        gitlab.Ptr("New MR"),
					SourceBranch: gitlab.Ptr("feature"),
					TargetBranch: gitlab.Ptr("main"),
				}

				mrs.On("CreateMergeRequest", int64(123), expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           1,
							IID:          10,
							Title:        "New MR",
							State:        "opened",
							SourceBranch: "feature",
							TargetBranch: "main",
							Author:       &gitlab.BasicUser{ID: 1, Username: "user1", Name: "User One"},
							CreatedAt:    &testTime,
							UpdatedAt:    &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           1,
				IID:          10,
				Title:        "New MR",
				State:        "opened",
				SourceBranch: "feature",
				TargetBranch: "main",
				Author:       map[string]any{"id": int64(1), "username": "user1", "name": "User One"},
				CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name: "successful create with all options",
			opts: &CreateMergeRequestOptions{
				Title:        "Full MR",
				SourceBranch: "feature",
				TargetBranch: "main",
				Description:  "Full description",
				Labels:       []string{"bug", "urgent"},
				AssigneeIDs:  []int64{1, 2},
				ReviewerIDs:  []int64{3},
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"bug", "urgent"})
				expectedOpts := &gitlab.CreateMergeRequestOptions{
					Title:        gitlab.Ptr("Full MR"),
					SourceBranch: gitlab.Ptr("feature"),
					TargetBranch: gitlab.Ptr("main"),
					Description:  gitlab.Ptr("Full description"),
					Labels:       &expectedLabels,
					AssigneeIDs:  &[]int64{1, 2},
					ReviewerIDs:  &[]int64{3},
				}

				mrs.On("CreateMergeRequest", int64(123), expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           2,
							IID:          11,
							Title:        "Full MR",
							Description:  "Full description",
							State:        "opened",
							SourceBranch: "feature",
							TargetBranch: "main",
							Labels:       gitlab.Labels{"bug", "urgent"},
							Author:       &gitlab.BasicUser{ID: 1, Username: "user1", Name: "User One"},
							Assignees:    []*gitlab.BasicUser{{ID: 1, Username: "a1", Name: "Assignee 1"}, {ID: 2, Username: "a2", Name: "Assignee 2"}},
							Reviewers:    []*gitlab.BasicUser{{ID: 3, Username: "r1", Name: "Reviewer 1"}},
							CreatedAt:    &testTime,
							UpdatedAt:    &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           2,
				IID:          11,
				Title:        "Full MR",
				Description:  "Full description",
				State:        "opened",
				SourceBranch: "feature",
				TargetBranch: "main",
				Labels:       []string{"bug", "urgent"},
				Author:       map[string]any{"id": int64(1), "username": "user1", "name": "User One"},
				Assignees:    []map[string]any{{"id": int64(1), "username": "a1", "name": "Assignee 1"}, {"id": int64(2), "username": "a2", "name": "Assignee 2"}},
				Reviewers:    []map[string]any{{"id": int64(3), "username": "r1", "name": "Reviewer 1"}},
				CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:    "nil options",
			opts:    nil,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:    "empty title",
			opts:    &CreateMergeRequestOptions{Title: "", SourceBranch: "feature", TargetBranch: "main"},
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:    "empty source branch",
			opts:    &CreateMergeRequestOptions{Title: "MR", SourceBranch: "", TargetBranch: "main"},
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:    "empty target branch",
			opts:    &CreateMergeRequestOptions{Title: "MR", SourceBranch: "feature", TargetBranch: ""},
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name: "project not found",
			opts: &CreateMergeRequestOptions{Title: "MR", SourceBranch: "feature", TargetBranch: "main"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name: "API error on create",
			opts: &CreateMergeRequestOptions{Title: "MR", SourceBranch: "feature", TargetBranch: "main"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("CreateMergeRequest", int64(123), &gitlab.CreateMergeRequestOptions{
					Title:        gitlab.Ptr("MR"),
					SourceBranch: gitlab.Ptr("feature"),
					TargetBranch: gitlab.Ptr("main"),
				}).Return(
					(*gitlab.MergeRequest)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockMRs := &MockMergeRequestsService{}

			tt.setup(mockClient, mockProjects, mockMRs)

			app := newTestApp(mockClient)
			result, err := app.CreateProjectMergeRequest("test/project", tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_GetMergeRequest(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		mrIID   int64
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService)
		want    *MergeRequest
		wantErr bool
	}{
		{
			name:  "successful get",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("GetMergeRequest", int64(123), int(10), (*gitlab.GetMergeRequestsOptions)(nil)).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           1,
							IID:          10,
							Title:        "Test MR",
							State:        "opened",
							SourceBranch: "feature",
							TargetBranch: "main",
							CreatedAt:    &testTime,
							UpdatedAt:    &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           1,
				IID:          10,
				Title:        "Test MR",
				State:        "opened",
				SourceBranch: "feature",
				TargetBranch: "main",
				CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:    "invalid IID zero",
			mrIID:   0,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:    "invalid IID negative",
			mrIID:   -1,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:  "project not found",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name:  "API error",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("GetMergeRequest", int64(123), int(10), (*gitlab.GetMergeRequestsOptions)(nil)).Return(
					(*gitlab.MergeRequest)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockMRs := &MockMergeRequestsService{}

			tt.setup(mockClient, mockProjects, mockMRs)

			app := newTestApp(mockClient)
			result, err := app.GetMergeRequest("test/project", tt.mrIID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_UpdateMergeRequest(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		mrIID   int64
		opts    *UpdateMergeRequestOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService)
		want    *MergeRequest
		wantErr bool
	}{
		{
			name:  "successful partial update",
			mrIID: 10,
			opts:  &UpdateMergeRequestOptions{Title: "Updated Title"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.UpdateMergeRequestOptions{
					Title: gitlab.Ptr("Updated Title"),
				}

				mrs.On("UpdateMergeRequest", int64(123), int(10), expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           1,
							IID:          10,
							Title:        "Updated Title",
							State:        "opened",
							SourceBranch: "feature",
							TargetBranch: "main",
							CreatedAt:    &testTime,
							UpdatedAt:    &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           1,
				IID:          10,
				Title:        "Updated Title",
				State:        "opened",
				SourceBranch: "feature",
				TargetBranch: "main",
				CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:  "successful full update",
			mrIID: 10,
			opts: &UpdateMergeRequestOptions{
				Title:       "New Title",
				Description: "New desc",
				State:       "closed",
				Labels:      []string{"label1"},
				AssigneeIDs: []int64{1},
				ReviewerIDs: []int64{2},
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"label1"})
				expectedOpts := &gitlab.UpdateMergeRequestOptions{
					Title:       gitlab.Ptr("New Title"),
					Description: gitlab.Ptr("New desc"),
					StateEvent:  gitlab.Ptr("closed"),
					Labels:      &expectedLabels,
					AssigneeIDs: &[]int64{1},
					ReviewerIDs: &[]int64{2},
				}

				mrs.On("UpdateMergeRequest", int64(123), int(10), expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:    1,
							IID:   10,
							Title: "New Title",
							State: "closed",
							Labels: gitlab.Labels{"label1"},
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:        1,
				IID:       10,
				Title:     "New Title",
				State:     "closed",
				Labels:    []string{"label1"},
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:    "invalid IID",
			mrIID:   0,
			opts:    &UpdateMergeRequestOptions{Title: "T"},
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:    "nil options",
			mrIID:   10,
			opts:    nil,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:  "project not found",
			mrIID: 10,
			opts:  &UpdateMergeRequestOptions{Title: "T"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name:  "API error",
			mrIID: 10,
			opts:  &UpdateMergeRequestOptions{Title: "T"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("UpdateMergeRequest", int64(123), int(10), &gitlab.UpdateMergeRequestOptions{
					Title: gitlab.Ptr("T"),
				}).Return(
					(*gitlab.MergeRequest)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockMRs := &MockMergeRequestsService{}

			tt.setup(mockClient, mockProjects, mockMRs)

			app := newTestApp(mockClient)
			result, err := app.UpdateMergeRequest("test/project", tt.mrIID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_MergeMergeRequest(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		mrIID   int64
		opts    *MergeMergeRequestOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService)
		want    *MergeRequest
		wantErr bool
	}{
		{
			name:  "successful merge with no options",
			mrIID: 10,
			opts:  nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("AcceptMergeRequest", int64(123), int(10), &gitlab.AcceptMergeRequestOptions{}).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:    1,
							IID:   10,
							Title: "Merged MR",
							State: "merged",
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:        1,
				IID:       10,
				Title:     "Merged MR",
				State:     "merged",
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:  "successful merge with options",
			mrIID: 10,
			opts: &MergeMergeRequestOptions{
				MergeCommitMessage:       "Merge commit",
				SquashCommitMessage:      "Squash commit",
				Squash:                   true,
				ShouldRemoveSourceBranch: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("AcceptMergeRequest", int64(123), int(10), &gitlab.AcceptMergeRequestOptions{
					MergeCommitMessage:       gitlab.Ptr("Merge commit"),
					SquashCommitMessage:      gitlab.Ptr("Squash commit"),
					Squash:                   gitlab.Ptr(true),
					ShouldRemoveSourceBranch: gitlab.Ptr(true),
				}).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:    1,
							IID:   10,
							Title: "Merged MR",
							State: "merged",
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:        1,
				IID:       10,
				Title:     "Merged MR",
				State:     "merged",
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:    "invalid IID",
			mrIID:   0,
			opts:    nil,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:  "project not found",
			mrIID: 10,
			opts:  nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name:  "API error",
			mrIID: 10,
			opts:  nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("AcceptMergeRequest", int64(123), int(10), &gitlab.AcceptMergeRequestOptions{}).Return(
					(*gitlab.MergeRequest)(nil), (*gitlab.Response)(nil), errors.New("merge failed"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockMRs := &MockMergeRequestsService{}

			tt.setup(mockClient, mockProjects, mockMRs)

			app := newTestApp(mockClient)
			result, err := app.MergeMergeRequest("test/project", tt.mrIID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_GetMergeRequestDiff(t *testing.T) {
	tests := []struct {
		name    string
		mrIID   int64
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService)
		want    string
		wantErr bool
	}{
		{
			name:  "successful diff with mixed file types",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("ListMergeRequestDiffs", int64(123), int(10), (*gitlab.ListMergeRequestDiffsOptions)(nil)).Return(
					[]*gitlab.MergeRequestDiff{
						{
							OldPath: "file.go",
							NewPath: "file.go",
							Diff:    "@@ -1,3 +1,4 @@\n package main\n+import \"fmt\"\n",
						},
						{
							NewPath: "new_file.go",
							NewFile: true,
							Diff:    "@@ -0,0 +1,3 @@\n+package main\n+func hello() {}\n",
						},
						{
							OldPath:     "old_file.go",
							DeletedFile: true,
							Diff:        "@@ -1,2 +0,0 @@\n-package main\n-func old() {}\n",
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: "--- a/file.go\n+++ b/file.go\n@@ -1,3 +1,4 @@\n package main\n+import \"fmt\"\n" +
				"\n--- /dev/null\n+++ b/new_file.go\n@@ -0,0 +1,3 @@\n+package main\n+func hello() {}\n" +
				"\n--- a/old_file.go\n+++ /dev/null\n@@ -1,2 +0,0 @@\n-package main\n-func old() {}\n",
		},
		{
			name:  "empty diffs",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("ListMergeRequestDiffs", int64(123), int(10), (*gitlab.ListMergeRequestDiffsOptions)(nil)).Return(
					[]*gitlab.MergeRequestDiff{},
					&gitlab.Response{}, nil,
				)
			},
			want: "No diffs available for this merge request",
		},
		{
			name:    "invalid IID",
			mrIID:   0,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {},
			wantErr: true,
		},
		{
			name:  "project not found",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name:  "API error",
			mrIID: 10,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				mrs.On("ListMergeRequestDiffs", int64(123), int(10), (*gitlab.ListMergeRequestDiffsOptions)(nil)).Return(
					([]*gitlab.MergeRequestDiff)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockMRs := &MockMergeRequestsService{}

			tt.setup(mockClient, mockProjects, mockMRs)

			app := newTestApp(mockClient)
			result, err := app.GetMergeRequestDiff("test/project", tt.mrIID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_ApproveMergeRequest(t *testing.T) {
	tests := []struct {
		name    string
		mrIID   int64
		opts    *ApproveMergeRequestOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestApprovalsService)
		want    string
		wantErr bool
	}{
		{
			name:  "successful approve without SHA",
			mrIID: 10,
			opts:  &ApproveMergeRequestOptions{},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, approvals *MockMergeRequestApprovalsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequestApprovals").Return(approvals)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				approvals.On("ApproveMergeRequest", int64(123), int64(10), &gitlab.ApproveMergeRequestOptions{}).Return(
					&gitlab.MergeRequestApprovals{
						ApprovedBy:        []*gitlab.MergeRequestApproverUser{{User: &gitlab.BasicUser{ID: 1}}},
						ApprovalsRequired: 1,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: "Merge request 10 approved successfully. Approvals: 1/1",
		},
		{
			name:  "successful approve with SHA",
			mrIID: 10,
			opts:  &ApproveMergeRequestOptions{SHA: "abc123"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, approvals *MockMergeRequestApprovalsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequestApprovals").Return(approvals)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				approvals.On("ApproveMergeRequest", int64(123), int64(10), &gitlab.ApproveMergeRequestOptions{
					SHA: gitlab.Ptr("abc123"),
				}).Return(
					&gitlab.MergeRequestApprovals{
						ApprovedBy:        []*gitlab.MergeRequestApproverUser{{User: &gitlab.BasicUser{ID: 1}}, {User: &gitlab.BasicUser{ID: 2}}},
						ApprovalsRequired: 2,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: "Merge request 10 approved successfully. Approvals: 2/2",
		},
		{
			name:    "invalid IID",
			mrIID:   0,
			opts:    nil,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, approvals *MockMergeRequestApprovalsService) {},
			wantErr: true,
		},
		{
			name:  "project not found",
			mrIID: 10,
			opts:  nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, approvals *MockMergeRequestApprovalsService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name:  "API error",
			mrIID: 10,
			opts:  &ApproveMergeRequestOptions{},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, approvals *MockMergeRequestApprovalsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequestApprovals").Return(approvals)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				approvals.On("ApproveMergeRequest", int64(123), int64(10), &gitlab.ApproveMergeRequestOptions{}).Return(
					(*gitlab.MergeRequestApprovals)(nil), (*gitlab.Response)(nil), errors.New("forbidden"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockApprovals := &MockMergeRequestApprovalsService{}

			tt.setup(mockClient, mockProjects, mockApprovals)

			app := newTestApp(mockClient)
			result, err := app.ApproveMergeRequest("test/project", tt.mrIID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockApprovals.AssertExpectations(t)
		})
	}
}

func TestApp_AddMergeRequestNote(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		mrIID   int64
		opts    *AddMergeRequestNoteOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockNotesService)
		want    *Note
		wantErr bool
	}{
		{
			name:  "successful note creation",
			mrIID: 10,
			opts:  &AddMergeRequestNoteOptions{Body: "LGTM!"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {
				client.On("Projects").Return(projects)
				client.On("Notes").Return(notes)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				notes.On("CreateMergeRequestNote", int64(123), int64(10), &gitlab.CreateMergeRequestNoteOptions{
					Body: gitlab.Ptr("LGTM!"),
				}).Return(
					&gitlab.Note{
						ID:          1,
						Body:        "LGTM!",
						System:      false,
						Author:      gitlab.NoteAuthor{ID: 5, Username: "reviewer", Name: "Reviewer"},
						NoteableID:  10,
						NoteableIID: 10,
						NoteableType: "MergeRequest",
						CreatedAt:   &testTime,
						UpdatedAt:   &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Note{
				ID:   1,
				Body: "LGTM!",
				Author: map[string]any{
					"id":       int64(5),
					"username": "reviewer",
					"name":     "Reviewer",
				},
				Noteable: map[string]any{
					"id":   int64(10),
					"iid":  int64(10),
					"type": "MergeRequest",
				},
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name:    "empty body",
			mrIID:   10,
			opts:    &AddMergeRequestNoteOptions{Body: ""},
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {},
			wantErr: true,
		},
		{
			name:    "nil options",
			mrIID:   10,
			opts:    nil,
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {},
			wantErr: true,
		},
		{
			name:    "invalid IID",
			mrIID:   0,
			opts:    &AddMergeRequestNoteOptions{Body: "test"},
			setup:   func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {},
			wantErr: true,
		},
		{
			name:  "project not found",
			mrIID: 10,
			opts:  &AddMergeRequestNoteOptions{Body: "test"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {
				client.On("Projects").Return(projects)
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("not found"),
				)
			},
			wantErr: true,
		},
		{
			name:  "API error",
			mrIID: 10,
			opts:  &AddMergeRequestNoteOptions{Body: "test"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {
				client.On("Projects").Return(projects)
				client.On("Notes").Return(notes)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				notes.On("CreateMergeRequestNote", int64(123), int64(10), &gitlab.CreateMergeRequestNoteOptions{
					Body: gitlab.Ptr("test"),
				}).Return(
					(*gitlab.Note)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockNotes := &MockNotesService{}

			tt.setup(mockClient, mockProjects, mockNotes)

			app := newTestApp(mockClient)
			result, err := app.AddMergeRequestNote("test/project", tt.mrIID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockNotes.AssertExpectations(t)
		})
	}
}

func TestConvertGitLabMergeRequest(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name  string
		input *gitlab.MergeRequest
		want  MergeRequest
	}{
		{
			name: "full conversion with all fields",
			input: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					ID:                  1,
					IID:                 10,
					Title:               "Full MR",
					Description:         "Description",
					State:               "opened",
					SourceBranch:        "feature",
					TargetBranch:        "main",
					Labels:              gitlab.Labels{"bug", "urgent"},
					WebURL:              "https://gitlab.com/test/-/merge_requests/10",
					DetailedMergeStatus: "mergeable",
					Draft:               true,
					Author:              &gitlab.BasicUser{ID: 1, Username: "user1", Name: "User One"},
					Assignees:           []*gitlab.BasicUser{{ID: 2, Username: "a1", Name: "Assignee"}},
					Reviewers:           []*gitlab.BasicUser{{ID: 3, Username: "r1", Name: "Reviewer"}},
					CreatedAt:           &testTime,
					UpdatedAt:           &testTime,
				},
			},
			want: MergeRequest{
				ID:             1,
				IID:            10,
				Title:          "Full MR",
				Description:    "Description",
				State:          "opened",
				SourceBranch:   "feature",
				TargetBranch:   "main",
				Labels:         []string{"bug", "urgent"},
				WebURL:         "https://gitlab.com/test/-/merge_requests/10",
				MergeStatus:    "mergeable",
				Draft:          true,
				WorkInProgress: true,
				Author:         map[string]any{"id": int64(1), "username": "user1", "name": "User One"},
				Assignees:      []map[string]any{{"id": int64(2), "username": "a1", "name": "Assignee"}},
				Reviewers:      []map[string]any{{"id": int64(3), "username": "r1", "name": "Reviewer"}},
				CreatedAt:      testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:      testTime.Format("2006-01-02T15:04:05Z"),
			},
		},
		{
			name: "nil author and empty collections",
			input: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					ID:    1,
					IID:   10,
					Title: "Minimal MR",
					State: "opened",
				},
			},
			want: MergeRequest{
				ID:    1,
				IID:   10,
				Title: "Minimal MR",
				State: "opened",
			},
		},
		{
			name: "nil timestamps",
			input: &gitlab.MergeRequest{
				BasicMergeRequest: gitlab.BasicMergeRequest{
					ID:        1,
					IID:       10,
					Title:     "No Timestamps",
					State:     "opened",
					Author:    &gitlab.BasicUser{ID: 1, Username: "u", Name: "U"},
					CreatedAt: nil,
					UpdatedAt: nil,
				},
			},
			want: MergeRequest{
				ID:     1,
				IID:    10,
				Title:  "No Timestamps",
				State:  "opened",
				Author: map[string]any{"id": int64(1), "username": "u", "name": "U"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertGitLabMergeRequest(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}
