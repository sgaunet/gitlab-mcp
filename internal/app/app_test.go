package app

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/api/client-go"
)

func TestApp_ValidateConnection(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockGitLabClient, *MockUsersService)
		wantErr bool
	}{
		{
			name: "successful validation",
			setup: func(client *MockGitLabClient, users *MockUsersService) {
				client.On("Users").Return(users)
				users.On("CurrentUser").Return(&gitlab.User{ID: 1}, &gitlab.Response{}, nil)
			},
			wantErr: false,
		},
		{
			name: "validation fails",
			setup: func(client *MockGitLabClient, users *MockUsersService) {
				client.On("Users").Return(users)
				users.On("CurrentUser").Return((*gitlab.User)(nil), (*gitlab.Response)(nil), errors.New("invalid token"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockUsers := &MockUsersService{}

			tt.setup(mockClient, mockUsers)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)

			err := app.ValidateConnection()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
			mockUsers.AssertExpectations(t)
		})
	}
}

func TestApp_ListProjectIssues(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *ListIssuesOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockIssuesService)
		want    []Issue
		wantErr bool
	}{
		{
			name: "successful list with default options",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), expectedOpts).Return(
					[]*gitlab.Issue{
						{
							ID:          1,
							IID:         10,
							Title:       "Test Issue",
							Description: "Test Description",
							State:       "opened",
							Labels:      []string{"bug", "high-priority"},
							Assignees:   []*gitlab.IssueAssignee{{ID: 1, Username: "user1", Name: "User One"}},
							CreatedAt:   &testTime,
							UpdatedAt:   &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []Issue{
				{
					ID:          1,
					IID:         10,
					Title:       "Test Issue",
					Description: "Test Description",
					State:       "opened",
					Labels:      []string{"bug", "high-priority"},
					Assignees:   []map[string]interface{}{{"id": int64(1), "username": "user1", "name": "User One"}},
					CreatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
					UpdatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "successful list with custom options",
			opts: &ListIssuesOptions{State: "closed", Limit: 50},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("closed"),
					ListOptions: gitlab.ListOptions{PerPage: 50, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), expectedOpts).Return(
					[]*gitlab.Issue{},
					&gitlab.Response{}, nil,
				)
			},
			want:    []Issue{},
			wantErr: false,
		},
		{
			name: "successful list with label filter",
			opts: &ListIssuesOptions{State: "opened", Labels: "bug", Limit: 100},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"bug"})
				expectedOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					Labels:      &expectedLabels,
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), expectedOpts).Return(
					[]*gitlab.Issue{
						{
							ID:          1,
							IID:         10,
							Title:       "Bug Issue",
							Description: "Bug description",
							State:       "opened",
							Labels:      []string{"bug"},
							Assignees:   []*gitlab.IssueAssignee{},
							CreatedAt:   &testTime,
							UpdatedAt:   &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []Issue{
				{
					ID:          1,
					IID:         10,
					Title:       "Bug Issue",
					Description: "Bug description",
					State:       "opened",
					Labels:      []string{"bug"},
					Assignees:   []map[string]interface{}{},
					CreatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
					UpdatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "successful list with multiple labels",
			opts: &ListIssuesOptions{State: "opened", Labels: "bug, priority-high, needs-review", Limit: 50},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"bug", "priority-high", "needs-review"})
				expectedOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					Labels:      &expectedLabels,
					ListOptions: gitlab.ListOptions{PerPage: 50, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), expectedOpts).Return(
					[]*gitlab.Issue{
						{
							ID:          2,
							IID:         20,
							Title:       "Critical Bug",
							Description: "High priority bug needing review",
							State:       "opened",
							Labels:      []string{"bug", "priority-high", "needs-review"},
							Assignees:   []*gitlab.IssueAssignee{},
							CreatedAt:   &testTime,
							UpdatedAt:   &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []Issue{
				{
					ID:          2,
					IID:         20,
					Title:       "Critical Bug",
					Description: "High priority bug needing review",
					State:       "opened",
					Labels:      []string{"bug", "priority-high", "needs-review"},
					Assignees:   []map[string]interface{}{},
					CreatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
					UpdatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "project not found",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("project not found"),
				)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "issues list fails",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), expectedOpts).Return(
					([]*gitlab.Issue)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockIssues := &MockIssuesService{}

			tt.setup(mockClient, mockProjects, mockIssues)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			result, err := app.ListProjectIssues("test/project", tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockIssues.AssertExpectations(t)
		})
	}
}

func TestApp_CreateProjectIssue(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *CreateIssueOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockIssuesService)
		want    *Issue
		wantErr bool
	}{
		{
			name: "successful create with minimal options",
			opts: &CreateIssueOptions{Title: "New Issue"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateIssueOptions{
					Title:       gitlab.Ptr("New Issue"),
					Description: gitlab.Ptr(""),
				}

				issues.On("CreateIssue", int64(123), expectedOpts).Return(
					&gitlab.Issue{
						ID:          2,
						IID:         11,
						Title:       "New Issue",
						Description: "",
						State:       "opened",
						Labels:      []string{},
						Assignees:   []*gitlab.IssueAssignee{},
						CreatedAt:   &testTime,
						UpdatedAt:   &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Issue{
				ID:          2,
				IID:         11,
				Title:       "New Issue",
				Description: "",
				State:       "opened",
				Labels:      []string{},
				Assignees:   []map[string]interface{}{},
				CreatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name: "successful create with all options",
			opts: &CreateIssueOptions{
				Title:       "Full Issue",
				Description: "Full description",
				Labels:      []string{"bug", "priority-high"},
				Assignees:   []int64{1, 2},
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				mockLabels := &MockLabelsService{}
				client.On("Projects").Return(projects).Times(2) // Once for create, once for validation
				client.On("Issues").Return(issues)
				client.On("Labels").Return(mockLabels)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(2)

				// Mock for label validation
				listOpts := &gitlab.ListLabelsOptions{
					WithCounts:            gitlab.Ptr(false),
					IncludeAncestorGroups: gitlab.Ptr(false),
					ListOptions:           gitlab.ListOptions{PerPage: 100, Page: 1},
				}
				mockLabels.On("ListLabels", int64(123), listOpts).Return(
					[]*gitlab.Label{
						{ID: 1, Name: "bug"},
						{ID: 2, Name: "priority-high"},
						{ID: 3, Name: "enhancement"},
					},
					&gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"bug", "priority-high"})
				expectedOpts := &gitlab.CreateIssueOptions{
					Title:       gitlab.Ptr("Full Issue"),
					Description: gitlab.Ptr("Full description"),
					Labels:      &expectedLabels,
					AssigneeIDs: &[]int64{1, 2},
				}

				issues.On("CreateIssue", int64(123), expectedOpts).Return(
					&gitlab.Issue{
						ID:          3,
						IID:         12,
						Title:       "Full Issue",
						Description: "Full description",
						State:       "opened",
						Labels:      []string{"bug", "priority-high"},
						Assignees: []*gitlab.IssueAssignee{
							{ID: 1, Username: "user1", Name: "User One"},
							{ID: 2, Username: "user2", Name: "User Two"},
						},
						CreatedAt: &testTime,
						UpdatedAt: &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Issue{
				ID:          3,
				IID:         12,
				Title:       "Full Issue",
				Description: "Full description",
				State:       "opened",
				Labels:      []string{"bug", "priority-high"},
				Assignees: []map[string]interface{}{
					{"id": int64(1), "username": "user1", "name": "User One"},
					{"id": int64(2), "username": "user2", "name": "User Two"},
				},
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockIssuesService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty title",
			opts:    &CreateIssueOptions{Title: ""},
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockIssuesService) {},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockIssues := &MockIssuesService{}

			tt.setup(mockClient, mockProjects, mockIssues)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			result, err := app.CreateProjectIssue("test/project", tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockIssues.AssertExpectations(t)
		})
	}
}

func TestApp_CreateProjectIssue_LabelValidation(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name            string
		validateLabels  bool
		issueLabels     []string
		setup           func(*MockGitLabClient, *MockProjectsService, *MockIssuesService, *MockLabelsService)
		wantErr         bool
		wantErrContains string
	}{
		{
			name:           "validation disabled - should succeed with non-existent labels",
			validateLabels: false,
			issueLabels:    []string{"non-existent-label"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService, labels *MockLabelsService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"non-existent-label"})
				expectedOpts := &gitlab.CreateIssueOptions{
					Title:       gitlab.Ptr("Test Issue"),
					Description: gitlab.Ptr(""),
					Labels:      &expectedLabels,
				}

				issues.On("CreateIssue", int64(123), expectedOpts).Return(
					&gitlab.Issue{
						ID:        1,
						IID:       1,
						Title:     "Test Issue",
						State:     "opened",
						Labels:    []string{}, // GitLab ignores non-existent labels
						CreatedAt: &testTime,
						UpdatedAt: &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			wantErr: false,
		},
		{
			name:           "validation enabled - should succeed with existing labels",
			validateLabels: true,
			issueLabels:    []string{"bug", "enhancement"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService, labels *MockLabelsService) {
				client.On("Projects").Return(projects).Times(2) // Once for create, once for validation
				client.On("Issues").Return(issues)
				client.On("Labels").Return(labels)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(2)

				// Mock for label validation
				listOpts := &gitlab.ListLabelsOptions{
					WithCounts:            gitlab.Ptr(false),
					IncludeAncestorGroups: gitlab.Ptr(false),
					ListOptions:           gitlab.ListOptions{PerPage: 100, Page: 1},
				}
				labels.On("ListLabels", int64(123), listOpts).Return(
					[]*gitlab.Label{
						{ID: 1, Name: "bug"},
						{ID: 2, Name: "enhancement"},
						{ID: 3, Name: "documentation"},
					},
					&gitlab.Response{}, nil,
				)

				// Mock for issue creation
				expectedLabels := gitlab.LabelOptions([]string{"bug", "enhancement"})
				expectedOpts := &gitlab.CreateIssueOptions{
					Title:       gitlab.Ptr("Test Issue"),
					Description: gitlab.Ptr(""),
					Labels:      &expectedLabels,
				}

				issues.On("CreateIssue", int64(123), expectedOpts).Return(
					&gitlab.Issue{
						ID:        1,
						IID:       1,
						Title:     "Test Issue",
						State:     "opened",
						Labels:    []string{"bug", "enhancement"},
						CreatedAt: &testTime,
						UpdatedAt: &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			wantErr: false,
		},
		{
			name:           "validation enabled - should fail with non-existent labels",
			validateLabels: true,
			issueLabels:    []string{"bug", "non-existent-label"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService, labels *MockLabelsService) {
				client.On("Projects").Return(projects)
				client.On("Labels").Return(labels)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				// Mock for label validation
				listOpts := &gitlab.ListLabelsOptions{
					WithCounts:            gitlab.Ptr(false),
					IncludeAncestorGroups: gitlab.Ptr(false),
					ListOptions:           gitlab.ListOptions{PerPage: 100, Page: 1},
				}
				labels.On("ListLabels", int64(123), listOpts).Return(
					[]*gitlab.Label{
						{ID: 1, Name: "bug"},
						{ID: 2, Name: "enhancement"},
					},
					&gitlab.Response{}, nil,
				)
			},
			wantErr:         true,
			wantErrContains: "non-existent-label",
		},
		{
			name:           "validation enabled - case insensitive matching should succeed",
			validateLabels: true,
			issueLabels:    []string{"BUG", "Enhancement"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService, labels *MockLabelsService) {
				client.On("Projects").Return(projects).Times(2)
				client.On("Issues").Return(issues)
				client.On("Labels").Return(labels)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(2)

				// Mock for label validation
				listOpts := &gitlab.ListLabelsOptions{
					WithCounts:            gitlab.Ptr(false),
					IncludeAncestorGroups: gitlab.Ptr(false),
					ListOptions:           gitlab.ListOptions{PerPage: 100, Page: 1},
				}
				labels.On("ListLabels", int64(123), listOpts).Return(
					[]*gitlab.Label{
						{ID: 1, Name: "bug"},
						{ID: 2, Name: "enhancement"},
					},
					&gitlab.Response{}, nil,
				)

				// Mock for issue creation
				expectedLabels := gitlab.LabelOptions([]string{"BUG", "Enhancement"})
				expectedOpts := &gitlab.CreateIssueOptions{
					Title:       gitlab.Ptr("Test Issue"),
					Description: gitlab.Ptr(""),
					Labels:      &expectedLabels,
				}

				issues.On("CreateIssue", int64(123), expectedOpts).Return(
					&gitlab.Issue{
						ID:        1,
						IID:       1,
						Title:     "Test Issue",
						State:     "opened",
						Labels:    []string{"BUG", "Enhancement"},
						CreatedAt: &testTime,
						UpdatedAt: &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockIssues := &MockIssuesService{}
			mockLabels := &MockLabelsService{}

			tt.setup(mockClient, mockProjects, mockIssues, mockLabels)

			app := NewWithClientAndValidation("token", "https://gitlab.com/", mockClient, tt.validateLabels)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			opts := &CreateIssueOptions{
				Title:  "Test Issue",
				Labels: tt.issueLabels,
			}

			result, err := app.CreateProjectIssue("test/project", opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockIssues.AssertExpectations(t)
			mockLabels.AssertExpectations(t)
		})
	}
}

func TestApp_ListProjectLabels(t *testing.T) {
	tests := []struct {
		name    string
		opts    *ListLabelsOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockLabelsService)
		want    []Label
		wantErr bool
	}{
		{
			name: "successful list with default options",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, labels *MockLabelsService) {
				client.On("Projects").Return(projects)
				client.On("Labels").Return(labels)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListLabelsOptions{
					WithCounts:            gitlab.Ptr(false),
					IncludeAncestorGroups: gitlab.Ptr(false),
					ListOptions:           gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				labels.On("ListLabels", int64(123), expectedOpts).Return(
					[]*gitlab.Label{
						{
							ID:                     1,
							Name:                   "bug",
							Color:                  "#FF0000",
							TextColor:              "#FFFFFF",
							Description:            "Bug label",
							OpenIssuesCount:        5,
							ClosedIssuesCount:      2,
							OpenMergeRequestsCount: 1,
							Subscribed:             true,
							Priority:               10,
							IsProjectLabel:         true,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []Label{
				{
					ID:                     1,
					Name:                   "bug",
					Color:                  "#FF0000",
					TextColor:              "#FFFFFF",
					Description:            "Bug label",
					OpenIssuesCount:        5,
					ClosedIssuesCount:      2,
					OpenMergeRequestsCount: 1,
					Subscribed:             true,
					Priority:               10,
					IsProjectLabel:         true,
				},
			},
			wantErr: false,
		},
		{
			name: "successful list with custom options",
			opts: &ListLabelsOptions{WithCounts: true, Search: "bug", Limit: 50},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, labels *MockLabelsService) {
				client.On("Projects").Return(projects)
				client.On("Labels").Return(labels)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListLabelsOptions{
					WithCounts:            gitlab.Ptr(true),
					IncludeAncestorGroups: gitlab.Ptr(false),
					Search:                gitlab.Ptr("bug"),
					ListOptions:           gitlab.ListOptions{PerPage: 50, Page: 1},
				}

				labels.On("ListLabels", int64(123), expectedOpts).Return(
					[]*gitlab.Label{},
					&gitlab.Response{}, nil,
				)
			},
			want:    []Label{},
			wantErr: false,
		},
		{
			name: "project not found",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, labels *MockLabelsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("project not found"),
				)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockLabels := &MockLabelsService{}

			tt.setup(mockClient, mockProjects, mockLabels)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			result, err := app.ListProjectLabels("test/project", tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockLabels.AssertExpectations(t)
		})
	}
}

func TestNewWithClient(t *testing.T) {
	mockClient := &MockGitLabClient{}

	app := NewWithClient("test-token", "https://gitlab.example.com/", mockClient)

	require.NotNil(t, app)
	assert.Equal(t, "test-token", app.GitLabToken)
	assert.Equal(t, "https://gitlab.example.com/", app.GitLabURI)
	assert.Equal(t, mockClient, app.client)
	assert.NotNil(t, app.logger)
}

func TestApp_GetAPIURL(t *testing.T) {
	mockClient := &MockGitLabClient{}

	app := NewWithClient("test-token", "https://gitlab.example.com", mockClient)

	expected := "https://gitlab.example.com/api/v4"
	assert.Equal(t, expected, app.GetAPIURL())
}

func TestApp_SetLogger(t *testing.T) {
	mockClient := &MockGitLabClient{}
	app := NewWithClient("test-token", "https://gitlab.com/", mockClient)

	logger := slog.New(slog.NewTextHandler(nil, nil))
	app.SetLogger(logger)

	assert.Equal(t, logger, app.logger)
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single label",
			input: "bug",
			want:  []string{"bug"},
		},
		{
			name:  "multiple labels",
			input: "bug,priority-high,needs-review",
			want:  []string{"bug", "priority-high", "needs-review"},
		},
		{
			name:  "labels with spaces",
			input: " bug , priority-high , needs-review ",
			want:  []string{"bug", "priority-high", "needs-review"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "only commas",
			input: ",,,",
			want:  []string{},
		},
		{
			name:  "labels with empty elements",
			input: "bug,,priority-high,,",
			want:  []string{"bug", "priority-high"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLabels(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestApp_UpdateProjectIssue(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name     string
		issueIID int64
		opts     *UpdateIssueOptions
		setup    func(*MockGitLabClient, *MockProjectsService, *MockIssuesService)
		want     *Issue
		wantErr  bool
	}{
		{
			name:     "successful update with all options",
			issueIID: 10,
			opts: &UpdateIssueOptions{
				Title:       "Updated Title",
				Description: "Updated description",
				State:       "closed",
				Labels:      []string{"bug", "fixed"},
				Assignees:   []int64{1, 2},
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedLabels := gitlab.LabelOptions([]string{"bug", "fixed"})
				expectedOpts := &gitlab.UpdateIssueOptions{
					Title:       gitlab.Ptr("Updated Title"),
					Description: gitlab.Ptr("Updated description"),
					StateEvent:  gitlab.Ptr("closed"),
					Labels:      &expectedLabels,
					AssigneeIDs: &[]int64{1, 2},
				}

				issues.On("UpdateIssue", int64(123), int64(10), expectedOpts).Return(
					&gitlab.Issue{
						ID:          3,
						IID:         10,
						Title:       "Updated Title",
						Description: "Updated description",
						State:       "closed",
						Labels:      []string{"bug", "fixed"},
						Assignees: []*gitlab.IssueAssignee{
							{ID: 1, Username: "user1", Name: "User One"},
							{ID: 2, Username: "user2", Name: "User Two"},
						},
						CreatedAt: &testTime,
						UpdatedAt: &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Issue{
				ID:          3,
				IID:         10,
				Title:       "Updated Title",
				Description: "Updated description",
				State:       "closed",
				Labels:      []string{"bug", "fixed"},
				Assignees: []map[string]interface{}{
					{"id": int64(1), "username": "user1", "name": "User One"},
					{"id": int64(2), "username": "user2", "name": "User Two"},
				},
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name:     "successful update with partial options",
			issueIID: 5,
			opts: &UpdateIssueOptions{
				Title: "Just updating title",
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.UpdateIssueOptions{
					Title: gitlab.Ptr("Just updating title"),
				}

				issues.On("UpdateIssue", int64(123), int64(5), expectedOpts).Return(
					&gitlab.Issue{
						ID:          4,
						IID:         5,
						Title:       "Just updating title",
						Description: "Original description",
						State:       "opened",
						Labels:      []string{},
						Assignees:   []*gitlab.IssueAssignee{},
						CreatedAt:   &testTime,
						UpdatedAt:   &testTime,
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Issue{
				ID:          4,
				IID:         5,
				Title:       "Just updating title",
				Description: "Original description",
				State:       "opened",
				Labels:      []string{},
				Assignees:   []map[string]interface{}{},
				CreatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name:     "invalid issue IID",
			issueIID: 0,
			opts:     &UpdateIssueOptions{Title: "Test"},
			setup:    func(*MockGitLabClient, *MockProjectsService, *MockIssuesService) {},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "nil options",
			issueIID: 1,
			opts:     nil,
			setup:    func(*MockGitLabClient, *MockProjectsService, *MockIssuesService) {},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "project not found",
			issueIID: 1,
			opts:     &UpdateIssueOptions{Title: "Test"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("project not found"),
				)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:     "update fails",
			issueIID: 1,
			opts:     &UpdateIssueOptions{Title: "Test"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.UpdateIssueOptions{
					Title: gitlab.Ptr("Test"),
				}

				issues.On("UpdateIssue", int64(123), int64(1), expectedOpts).Return(
					(*gitlab.Issue)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockIssues := &MockIssuesService{}

			tt.setup(mockClient, mockProjects, mockIssues)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			result, err := app.UpdateProjectIssue("test/project", tt.issueIID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockIssues.AssertExpectations(t)
		})
	}
}
func TestApp_AddIssueNote(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *AddIssueNoteOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockNotesService)
		want    *Note
		wantErr bool
	}{
		{
			name: "successful note creation",
			opts: &AddIssueNoteOptions{Body: "This is a test note"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {
				client.On("Projects").Return(projects)
				client.On("Notes").Return(notes)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateIssueNoteOptions{
					Body: gitlab.Ptr("This is a test note"),
				}

				notes.On("CreateIssueNote", int64(123), int64(10), expectedOpts).Return(
					&gitlab.Note{
						ID:           1,
						Body:         "This is a test note",
						System:       false,
						Author:       gitlab.NoteAuthor{ID: 1, Username: "testuser", Name: "Test User"},
						CreatedAt:    &testTime,
						UpdatedAt:    &testTime,
						NoteableID:   50,
						NoteableIID:  10,
						NoteableType: "Issue",
					}, &gitlab.Response{}, nil,
				)
			},
			want: &Note{
				ID:        1,
				Body:      "This is a test note",
				System:    false,
				Author:    map[string]interface{}{"id": int64(1), "username": "testuser", "name": "Test User"},
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				Noteable:  map[string]interface{}{"id": int64(50), "iid": int64(10), "type": "Issue"},
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockNotesService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty body",
			opts:    &AddIssueNoteOptions{Body: ""},
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockNotesService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "project not found",
			opts: &AddIssueNoteOptions{Body: "Test note"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "invalid/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("project not found"),
				)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "create note API error",
			opts: &AddIssueNoteOptions{Body: "Test note"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, notes *MockNotesService) {
				client.On("Projects").Return(projects)
				client.On("Notes").Return(notes)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateIssueNoteOptions{
					Body: gitlab.Ptr("Test note"),
				}

				notes.On("CreateIssueNote", int64(123), int64(10), expectedOpts).Return(
					(*gitlab.Note)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockNotes := &MockNotesService{}

			tt.setup(mockClient, mockProjects, mockNotes)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)

			projectPath := "test/project"
			if tt.name == "project not found" {
				projectPath = "invalid/project"
			}
			got, err := app.AddIssueNote(projectPath, 10, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockNotes.AssertExpectations(t)
		})
	}
}

func TestApp_AddIssueNote_InvalidIssueIID(t *testing.T) {
	app := NewWithClient("token", "https://gitlab.com/", &MockGitLabClient{})
	opts := &AddIssueNoteOptions{Body: "Test note"}

	// Test negative IID
	got, err := app.AddIssueNote("test/project", -1, opts)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidIssueIID, err)
	assert.Nil(t, got)

	// Test zero IID
	got, err = app.AddIssueNote("test/project", 0, opts)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidIssueIID, err)
	assert.Nil(t, got)
}


func TestApp_GetProjectDescription(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		setup       func(*MockGitLabClient, *MockProjectsService)
		want        *ProjectInfo
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful get project description",
			projectPath: "test/project",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:          123,
						Name:        "Test Project",
						Path:        "project",
						Description: "This is a test project description",
					}, &gitlab.Response{}, nil,
				)
			},
			want: &ProjectInfo{
				ID:          123,
				Name:        "Test Project",
				Path:        "project",
				Description: "This is a test project description",
			},
			wantErr: false,
		},
		{
			name:        "project not found",
			projectPath: "nonexistent/project",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "nonexistent/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("404 Project Not Found"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to get project",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}

			tc.setup(mockClient, mockProjects)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))

			got, err := app.GetProjectDescription(tc.projectPath)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
		})
	}
}

func TestApp_UpdateProjectDescription(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		description string
		setup       func(*MockGitLabClient, *MockProjectsService)
		want        *ProjectInfo
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful update project description",
			projectPath: "test/project",
			description: "Updated project description",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects).Times(2)

				// First call to get project ID
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:   123,
						Name: "Test Project",
						Path: "project",
					}, &gitlab.Response{}, nil,
				)

				// Second call to update project
				expectedOpts := &gitlab.EditProjectOptions{
					Description: gitlab.Ptr("Updated project description"),
				}
				projects.On("EditProject", int64(123), expectedOpts).Return(
					&gitlab.Project{
						ID:          123,
						Name:        "Test Project",
						Path:        "project",
						Description: "Updated project description",
						Topics:      []string{"topic1", "topic2"},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &ProjectInfo{
				ID:          123,
				Name:        "Test Project",
				Path:        "project",
				Description: "Updated project description",
				Topics:      []string{"topic1", "topic2"},
			},
			wantErr: false,
		},
		{
			name:        "project not found",
			projectPath: "nonexistent/project",
			description: "New description",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "nonexistent/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("404 Project Not Found"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to get project",
		},
		{
			name:        "update fails",
			projectPath: "test/project",
			description: "New description",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects).Times(2)

				// First call to get project ID
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:   123,
						Name: "Test Project",
						Path: "project",
					}, &gitlab.Response{}, nil,
				)

				// Second call to update project fails
				expectedOpts := &gitlab.EditProjectOptions{
					Description: gitlab.Ptr("New description"),
				}
				projects.On("EditProject", int64(123), expectedOpts).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("403 Forbidden"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to update project description",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}

			tc.setup(mockClient, mockProjects)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))

			got, err := app.UpdateProjectDescription(tc.projectPath, tc.description)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
		})
	}
}

func TestApp_GetProjectTopics(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		setup       func(*MockGitLabClient, *MockProjectsService)
		want        *ProjectInfo
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful get project topics",
			projectPath: "test/project",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:     123,
						Name:   "Test Project",
						Path:   "project",
						Topics: []string{"golang", "mcp", "gitlab"},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &ProjectInfo{
				ID:     123,
				Name:   "Test Project",
				Path:   "project",
				Topics: []string{"golang", "mcp", "gitlab"},
			},
			wantErr: false,
		},
		{
			name:        "project with no topics",
			projectPath: "test/project",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:     123,
						Name:   "Test Project",
						Path:   "project",
						Topics: []string{},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &ProjectInfo{
				ID:     123,
				Name:   "Test Project",
				Path:   "project",
				Topics: []string{},
			},
			wantErr: false,
		},
		{
			name:        "project not found",
			projectPath: "nonexistent/project",
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "nonexistent/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("404 Project Not Found"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to get project",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}

			tc.setup(mockClient, mockProjects)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))

			got, err := app.GetProjectTopics(tc.projectPath)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
		})
	}
}

func TestApp_UpdateProjectTopics(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		topics      []string
		setup       func(*MockGitLabClient, *MockProjectsService)
		want        *ProjectInfo
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful update project topics",
			projectPath: "test/project",
			topics:      []string{"golang", "api", "mcp"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects).Times(2)

				// First call to get project ID
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:   123,
						Name: "Test Project",
						Path: "project",
					}, &gitlab.Response{}, nil,
				)

				// Second call to update project
				expectedTopics := []string{"golang", "api", "mcp"}
				expectedOpts := &gitlab.EditProjectOptions{
					Topics: &expectedTopics,
				}
				projects.On("EditProject", int64(123), expectedOpts).Return(
					&gitlab.Project{
						ID:          123,
						Name:        "Test Project",
						Path:        "project",
						Description: "Test description",
						Topics:      []string{"golang", "api", "mcp"},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &ProjectInfo{
				ID:          123,
				Name:        "Test Project",
				Path:        "project",
				Description: "Test description",
				Topics:      []string{"golang", "api", "mcp"},
			},
			wantErr: false,
		},
		{
			name:        "clear all topics",
			projectPath: "test/project",
			topics:      []string{},
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects).Times(2)

				// First call to get project ID
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:   123,
						Name: "Test Project",
						Path: "project",
					}, &gitlab.Response{}, nil,
				)

				// Second call to update project with empty topics
				expectedTopics := []string{}
				expectedOpts := &gitlab.EditProjectOptions{
					Topics: &expectedTopics,
				}
				projects.On("EditProject", int64(123), expectedOpts).Return(
					&gitlab.Project{
						ID:          123,
						Name:        "Test Project",
						Path:        "project",
						Description: "Test description",
						Topics:      []string{},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &ProjectInfo{
				ID:          123,
				Name:        "Test Project",
				Path:        "project",
				Description: "Test description",
				Topics:      []string{},
			},
			wantErr: false,
		},
		{
			name:        "project not found",
			projectPath: "nonexistent/project",
			topics:      []string{"topic1"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "nonexistent/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("404 Project Not Found"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to get project",
		},
		{
			name:        "update fails",
			projectPath: "test/project",
			topics:      []string{"topic1"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService) {
				client.On("Projects").Return(projects).Times(2)

				// First call to get project ID
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{
						ID:   123,
						Name: "Test Project",
						Path: "project",
					}, &gitlab.Response{}, nil,
				)

				// Second call to update project fails
				expectedTopics := []string{"topic1"}
				expectedOpts := &gitlab.EditProjectOptions{
					Topics: &expectedTopics,
				}
				projects.On("EditProject", int64(123), expectedOpts).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("403 Forbidden"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to update project topics",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}

			tc.setup(mockClient, mockProjects)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))

			got, err := app.UpdateProjectTopics(tc.projectPath, tc.topics)

			if tc.wantErr {
				require.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
		})
	}
}

func TestApp_CreateGroupEpic(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *CreateEpicOptions
		setup   func(*MockGitLabClient, *MockGroupsService, *MockEpicsService)
		want    *Epic
		wantErr bool
		errType error
	}{
		{
			name: "successful create with all optional fields",
			opts: &CreateEpicOptions{
				Title:        "Test Epic",
				Description:  "Test Description",
				Labels:       []string{"epic", "high-priority"},
				StartDate:    "2024-03-01",
				DueDate:      "2024-06-30",
				Confidential: true,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				startDate := gitlab.ISOTime(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
				dueDate := gitlab.ISOTime(time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC))
				fixed := true
				labels := gitlab.LabelOptions{"epic", "high-priority"}
				confidential := true

				expectedOpts := &gitlab.CreateEpicOptions{
					Title:            gitlab.Ptr("Test Epic"),
					Description:      gitlab.Ptr("Test Description"),
					Labels:           &labels,
					StartDateIsFixed: &fixed,
					StartDateFixed:   &startDate,
					DueDateIsFixed:   &fixed,
					DueDateFixed:     &dueDate,
					Confidential:     &confidential,
				}

				epics.On("CreateEpic", int64(456), expectedOpts).Return(
					func() *gitlab.Epic {
						startDate := gitlab.ISOTime(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
						dueDate := gitlab.ISOTime(time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC))
						return &gitlab.Epic{
							ID:          123,
							IID:         5,
							GroupID:     456,
							Title:       "Test Epic",
							Description: "Test Description",
							State:       "opened",
							WebURL:      "https://gitlab.com/groups/test/group/-/epics/5",
							Author: &gitlab.EpicAuthor{
								ID:       1,
								Username: "testuser",
								Name:     "Test User",
							},
							StartDate: &startDate,
							DueDate:   &dueDate,
							Labels:    gitlab.LabelOptions{"epic", "high-priority"},
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						}
					}(),
					&gitlab.Response{},
					nil,
				)
			},
			want: &Epic{
				ID:          123,
				IID:         5,
				GroupID:     456,
				Title:       "Test Epic",
				Description: "Test Description",
				State:       "opened",
				WebURL:      "https://gitlab.com/groups/test/group/-/epics/5",
				Author: map[string]any{
					"id":       int64(1),
					"username": "testuser",
					"name":     "Test User",
				},
				StartDate: "2024-03-01",
				DueDate:   "2024-06-30",
				Labels:    []string{"epic", "high-priority"},
			},
			wantErr: false,
		},
		{
			name: "successful create with minimal options",
			opts: &CreateEpicOptions{
				Title: "Minimal Epic",
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateEpicOptions{
					Title:       gitlab.Ptr("Minimal Epic"),
					Description: gitlab.Ptr(""),
				}

				epics.On("CreateEpic", int64(456), expectedOpts).Return(
					&gitlab.Epic{
						ID:      123,
						IID:     5,
						GroupID: 456,
						Title:   "Minimal Epic",
						State:   "opened",
						Author: &gitlab.EpicAuthor{
							ID:       1,
							Username: "testuser",
						},
						Labels: gitlab.LabelOptions{},
					},
					&gitlab.Response{},
					nil,
				)
			},
			want: &Epic{
				ID:      123,
				IID:     5,
				GroupID: 456,
				Title:   "Minimal Epic",
				State:   "opened",
				Author: map[string]any{
					"id":       int64(1),
					"username": "testuser",
					"name":     "",
				},
				Labels: []string{},
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			setup:   func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {},
			want:    nil,
			wantErr: true,
			errType: ErrCreateOptionsRequired,
		},
		{
			name: "empty title",
			opts: &CreateEpicOptions{
				Title: "",
			},
			setup:   func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {},
			want:    nil,
			wantErr: true,
			errType: ErrEpicTitleRequired,
		},
		{
			name: "invalid start date",
			opts: &CreateEpicOptions{
				Title:     "Test Epic",
				StartDate: "2024-3-5",
			},
			setup:   func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {},
			want:    nil,
			wantErr: true,
			errType: ErrInvalidDateFormat,
		},
		{
			name: "invalid due date",
			opts: &CreateEpicOptions{
				Title:   "Test Epic",
				DueDate: "03/15/2024",
			},
			setup:   func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {},
			want:    nil,
			wantErr: true,
			errType: ErrInvalidDateFormat,
		},
		{
			name: "tier required",
			opts: &CreateEpicOptions{
				Title: "Test Epic",
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateEpicOptions{
					Title:       gitlab.Ptr("Test Epic"),
					Description: gitlab.Ptr(""),
				}

				epics.On("CreateEpic", int64(456), expectedOpts).Return(
					(*gitlab.Epic)(nil),
					&gitlab.Response{},
					errors.New("403 Forbidden"),
				)
			},
			want:    nil,
			wantErr: true,
			errType: ErrEpicsTierRequired,
		},
		{
			name: "group not found",
			opts: &CreateEpicOptions{
				Title: "Test Epic",
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					(*gitlab.Group)(nil),
					&gitlab.Response{},
					errors.New("group not found"),
				)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "create epic API error",
			opts: &CreateEpicOptions{
				Title: "Test Epic",
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.CreateEpicOptions{
					Title:       gitlab.Ptr("Test Epic"),
					Description: gitlab.Ptr(""),
				}

				epics.On("CreateEpic", int64(456), expectedOpts).Return(
					(*gitlab.Epic)(nil),
					&gitlab.Response{},
					errors.New("API error"),
				)
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockGroups := &MockGroupsService{}
			mockEpics := &MockEpicsService{}

			tt.setup(mockClient, mockGroups, mockEpics)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))

			got, err := app.CreateGroupEpic("test/group", tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.IID, got.IID)
				assert.Equal(t, tt.want.GroupID, got.GroupID)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Description, got.Description)
				assert.Equal(t, tt.want.State, got.State)
			}

			mockClient.AssertExpectations(t)
			mockGroups.AssertExpectations(t)
			mockEpics.AssertExpectations(t)
		})
	}
}

func TestApp_parseDate(t *testing.T) {
	app := NewWithClient("token", "https://gitlab.com/", &MockGitLabClient{})

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid date YYYY-MM-DD",
			input:   "2024-03-15",
			wantErr: false,
		},
		{
			name:    "invalid format MM/DD/YYYY",
			input:   "03/15/2024",
			wantErr: true,
		},
		{
			name:    "invalid format YYYY-M-D",
			input:   "2024-3-5",
			wantErr: true,
		},
		{
			name:    "completely invalid",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.parseDate(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidDateFormat)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				// Verify the date was parsed correctly
				parsedTime := time.Time(*got)
				assert.Equal(t, "2024-03-15", parsedTime.Format("2006-01-02"))
			}
		})
	}
}

func TestApp_GetLatestPipeline(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name    string
		opts    *GetLatestPipelineOptions
		setup   func(*MockGitLabClient, *MockProjectsService, *MockPipelinesService)
		want    *Pipeline
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful get latest pipeline",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, pipelines *MockPipelinesService) {
				client.On("Projects").Return(projects)
				client.On("Pipelines").Return(pipelines)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				orderBy := "updated_at"
				sort := "desc"
				expectedOpts := &gitlab.ListProjectPipelinesOptions{
					OrderBy:     &orderBy,
					Sort:        &sort,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}

				pipelines.On("ListProjectPipelines", int64(123), expectedOpts).Return(
					[]*gitlab.PipelineInfo{
						{
							ID:        42,
							IID:       10,
							ProjectID: 123,
							Status:    "success",
							Source:    "push",
							Ref:       "main",
							SHA:       "abc123def456",
							WebURL:    "https://gitlab.com/test/project/-/pipelines/42",
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Pipeline{
				ID:        42,
				IID:       10,
				ProjectID: 123,
				Status:    "success",
				Source:    "push",
				Ref:       "main",
				SHA:       "abc123def456",
				WebURL:    "https://gitlab.com/test/project/-/pipelines/42",
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name: "successful get with ref filter",
			opts: &GetLatestPipelineOptions{Ref: "develop"},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, pipelines *MockPipelinesService) {
				client.On("Projects").Return(projects)
				client.On("Pipelines").Return(pipelines)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				orderBy := "updated_at"
				sort := "desc"
				ref := "develop"
				expectedOpts := &gitlab.ListProjectPipelinesOptions{
					OrderBy:     &orderBy,
					Sort:        &sort,
					Ref:         &ref,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}

				pipelines.On("ListProjectPipelines", int64(123), expectedOpts).Return(
					[]*gitlab.PipelineInfo{
						{
							ID:        43,
							IID:       11,
							ProjectID: 123,
							Status:    "running",
							Source:    "push",
							Ref:       "develop",
							SHA:       "xyz789abc123",
							WebURL:    "https://gitlab.com/test/project/-/pipelines/43",
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &Pipeline{
				ID:        43,
				IID:       11,
				ProjectID: 123,
				Status:    "running",
				Source:    "push",
				Ref:       "develop",
				SHA:       "xyz789abc123",
				WebURL:    "https://gitlab.com/test/project/-/pipelines/43",
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name: "no pipelines found",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, pipelines *MockPipelinesService) {
				client.On("Projects").Return(projects)
				client.On("Pipelines").Return(pipelines)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				orderBy := "updated_at"
				sort := "desc"
				expectedOpts := &gitlab.ListProjectPipelinesOptions{
					OrderBy:     &orderBy,
					Sort:        &sort,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}

				pipelines.On("ListProjectPipelines", int64(123), expectedOpts).Return(
					[]*gitlab.PipelineInfo{},
					&gitlab.Response{}, nil,
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "no pipelines found",
		},
		{
			name: "project not found",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, pipelines *MockPipelinesService) {
				client.On("Projects").Return(projects)

				projects.On("GetProject", "test/nonexistent", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("404 Project Not Found"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to get project",
		},
		{
			name: "api error when listing",
			opts: nil,
			setup: func(client *MockGitLabClient, projects *MockProjectsService, pipelines *MockPipelinesService) {
				client.On("Projects").Return(projects)
				client.On("Pipelines").Return(pipelines)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				orderBy := "updated_at"
				sort := "desc"
				expectedOpts := &gitlab.ListProjectPipelinesOptions{
					OrderBy:     &orderBy,
					Sort:        &sort,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}

				pipelines.On("ListProjectPipelines", int64(123), expectedOpts).Return(
					([]*gitlab.PipelineInfo)(nil), (*gitlab.Response)(nil), errors.New("API error"),
				)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "failed to list project pipelines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockPipelines := &MockPipelinesService{}

			tt.setup(mockClient, mockProjects, mockPipelines)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)

			projectPath := "test/project"
			if tt.name == "project not found" {
				projectPath = "test/nonexistent"
			}

			result, err := app.GetLatestPipeline(projectPath, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.want.ID, result.ID)
				assert.Equal(t, tt.want.Status, result.Status)
				assert.Equal(t, tt.want.Ref, result.Ref)
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockPipelines.AssertExpectations(t)
		})
	}
}
