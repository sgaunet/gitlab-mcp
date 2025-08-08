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
				
				issues.On("ListProjectIssues", 123, expectedOpts).Return(
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
					Assignees:   []map[string]interface{}{{"id": 1, "username": "user1", "name": "User One"}},
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
				
				issues.On("ListProjectIssues", 123, expectedOpts).Return(
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
				
				issues.On("ListProjectIssues", 123, expectedOpts).Return(
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
				
				issues.On("ListProjectIssues", 123, expectedOpts).Return(
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
				
				issues.On("ListProjectIssues", 123, expectedOpts).Return(
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
				
				issues.On("CreateIssue", 123, expectedOpts).Return(
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
				Assignees:   []int{1, 2},
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)
				
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)
				
				expectedLabels := gitlab.LabelOptions([]string{"bug", "priority-high"})
				expectedOpts := &gitlab.CreateIssueOptions{
					Title:       gitlab.Ptr("Full Issue"),
					Description: gitlab.Ptr("Full description"),
					Labels:      &expectedLabels,
					AssigneeIDs: &[]int{1, 2},
				}
				
				issues.On("CreateIssue", 123, expectedOpts).Return(
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
					{"id": 1, "username": "user1", "name": "User One"},
					{"id": 2, "username": "user2", "name": "User Two"},
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
				
				labels.On("ListLabels", 123, expectedOpts).Return(
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
				
				labels.On("ListLabels", 123, expectedOpts).Return(
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
		name   string
		input  string
		want   []string
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
		name      string
		issueIID  int
		opts      *UpdateIssueOptions
		setup     func(*MockGitLabClient, *MockProjectsService, *MockIssuesService)
		want      *Issue
		wantErr   bool
	}{
		{
			name:     "successful update with all options",
			issueIID: 10,
			opts: &UpdateIssueOptions{
				Title:       "Updated Title",
				Description: "Updated description",
				State:       "closed",
				Labels:      []string{"bug", "fixed"},
				Assignees:   []int{1, 2},
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
					AssigneeIDs: &[]int{1, 2},
				}
				
				issues.On("UpdateIssue", 123, 10, expectedOpts).Return(
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
					{"id": 1, "username": "user1", "name": "User One"},
					{"id": 2, "username": "user2", "name": "User Two"},
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
				
				issues.On("UpdateIssue", 123, 5, expectedOpts).Return(
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
				
				issues.On("UpdateIssue", 123, 1, expectedOpts).Return(
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
				
				notes.On("CreateIssueNote", 123, 10, expectedOpts).Return(
					&gitlab.Note{
						ID:     1,
						Body:   "This is a test note",
						System: false,
						Author: gitlab.NoteAuthor{ID: 1, Username: "testuser", Name: "Test User"},
						CreatedAt: &testTime,
						UpdatedAt: &testTime,
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
				Author:    map[string]interface{}{"id": 1, "username": "testuser", "name": "Test User"},
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				Noteable:  map[string]interface{}{"id": 50, "iid": 10, "type": "Issue"},
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
				
				notes.On("CreateIssueNote", 123, 10, expectedOpts).Return(
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
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Title:        "Test MR",
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)
				
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)
				
				expectedOpts := &gitlab.CreateMergeRequestOptions{
					Title:              gitlab.Ptr("Test MR"),
					SourceBranch:       gitlab.Ptr("feature-branch"),
					TargetBranch:       gitlab.Ptr("main"),
					RemoveSourceBranch: gitlab.Ptr(false),
				}
				
				mrs.On("CreateMergeRequest", 123, expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           1,
							IID:          100,
							Title:        "Test MR",
							Description:  "",
							State:        "opened",
							SourceBranch: "feature-branch",
							TargetBranch: "main",
							Author:       &gitlab.BasicUser{ID: 1, Username: "testuser", Name: "Test User"},
							Assignees:    []*gitlab.BasicUser{},
							Reviewers:    []*gitlab.BasicUser{},
							Labels:       gitlab.Labels{},
							Milestone:    nil,
							WebURL:       "https://gitlab.com/test/project/-/merge_requests/100",
							Draft:        false,
							CreatedAt:    &testTime,
							UpdatedAt:    &testTime,
						},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           1,
				IID:          100,
				Title:        "Test MR",
				Description:  "",
				State:        "opened",
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Author:       map[string]interface{}{"id": 1, "username": "testuser", "name": "Test User"},
				Assignees:    []map[string]interface{}{},
				Reviewers:    []map[string]interface{}{},
				Labels:       []string{},
				Milestone:    nil,
				WebURL:       "https://gitlab.com/test/project/-/merge_requests/100",
				Draft:        false,
				CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name: "successful create with all options",
			opts: &CreateMergeRequestOptions{
				SourceBranch:       "feature-branch",
				TargetBranch:       "main",
				Title:              "Test MR with options",
				Description:        "This is a test merge request",
				Assignees:          []interface{}{1, 2},
				Reviewers:          []interface{}{3, 4},
				Labels:             []string{"enhancement", "feature"},
				Milestone:          10,
				RemoveSourceBranch: true,
				Draft:              true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)
				
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)
				
				expectedOpts := &gitlab.CreateMergeRequestOptions{
					Title:              gitlab.Ptr("Test MR with options"),
					SourceBranch:       gitlab.Ptr("feature-branch"),
					TargetBranch:       gitlab.Ptr("main"),
					Description:        gitlab.Ptr("This is a test merge request"),
					AssigneeIDs:        &[]int{1, 2},
					ReviewerIDs:        &[]int{3, 4},
					Labels:             (*gitlab.LabelOptions)(&[]string{"enhancement", "feature"}),
					MilestoneID:        gitlab.Ptr(10),
					RemoveSourceBranch: gitlab.Ptr(true),
				}
				
				mrs.On("CreateMergeRequest", 123, expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           2,
							IID:          101,
							Title:        "Test MR with options",
							Description:  "This is a test merge request",
							State:        "opened",
							SourceBranch: "feature-branch",
							TargetBranch: "main",
							Author:       &gitlab.BasicUser{ID: 1, Username: "testuser", Name: "Test User"},
							Assignees:    []*gitlab.BasicUser{{ID: 1, Username: "assignee1", Name: "Assignee One"}, {ID: 2, Username: "assignee2", Name: "Assignee Two"}},
							Reviewers:    []*gitlab.BasicUser{{ID: 3, Username: "reviewer1", Name: "Reviewer One"}, {ID: 4, Username: "reviewer2", Name: "Reviewer Two"}},
							Labels:       gitlab.Labels{"enhancement", "feature"},
							Milestone:    &gitlab.Milestone{ID: 10, Title: "v1.0"},
							WebURL:       "https://gitlab.com/test/project/-/merge_requests/101",
							Draft:        true,
							CreatedAt:    &testTime,
							UpdatedAt:    &testTime,
						},
					}, &gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           2,
				IID:          101,
				Title:        "Test MR with options",
				Description:  "This is a test merge request",
				State:        "opened",
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Author:       map[string]interface{}{"id": 1, "username": "testuser", "name": "Test User"},
				Assignees:    []map[string]interface{}{{"id": 1, "username": "assignee1", "name": "Assignee One"}, {"id": 2, "username": "assignee2", "name": "Assignee Two"}},
				Reviewers:    []map[string]interface{}{{"id": 3, "username": "reviewer1", "name": "Reviewer One"}, {"id": 4, "username": "reviewer2", "name": "Reviewer Two"}},
				Labels:       []string{"enhancement", "feature"},
				Milestone:    map[string]interface{}{"id": 10, "title": "v1.0"},
				WebURL:       "https://gitlab.com/test/project/-/merge_requests/101",
				Draft:        true,
				CreatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty title",
			opts: &CreateMergeRequestOptions{
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Title:        "",
			},
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty source branch",
			opts: &CreateMergeRequestOptions{
				SourceBranch: "",
				TargetBranch: "main",
				Title:        "Test MR",
			},
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty target branch",
			opts: &CreateMergeRequestOptions{
				SourceBranch: "feature-branch",
				TargetBranch: "",
				Title:        "Test MR",
			},
			setup:   func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService) {},
			want:    nil,
			wantErr: true,
		},
		{
			name: "project not found",
			opts: &CreateMergeRequestOptions{
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Title:        "Test MR",
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				
				projects.On("GetProject", "invalid/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), (*gitlab.Response)(nil), errors.New("project not found"),
				)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "create merge request API error",
			opts: &CreateMergeRequestOptions{
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Title:        "Test MR",
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)
				
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)
				
				expectedOpts := &gitlab.CreateMergeRequestOptions{
					Title:              gitlab.Ptr("Test MR"),
					SourceBranch:       gitlab.Ptr("feature-branch"),
					TargetBranch:       gitlab.Ptr("main"),
					RemoveSourceBranch: gitlab.Ptr(false),
				}
				
				mrs.On("CreateMergeRequest", 123, expectedOpts).Return(
					(*gitlab.MergeRequest)(nil), (*gitlab.Response)(nil), errors.New("API error"),
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
			mockMRs := &MockMergeRequestsService{}
			
			tt.setup(mockClient, mockProjects, mockMRs)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			
			projectPath := "test/project"
			if tt.name == "project not found" {
				projectPath = "invalid/project"
			}
			got, err := app.CreateProjectMergeRequest(projectPath, tt.opts)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			
			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockMRs.AssertExpectations(t)
		})
	}
}

func TestApp_CreateProjectMergeRequest_WithUsernameResolution(t *testing.T) {
	t.Parallel()
	
	testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	
	testCases := []struct {
		name     string
		opts     *CreateMergeRequestOptions
		setup    func(*MockGitLabClient, *MockProjectsService, *MockMergeRequestsService, *MockUsersService, *MockMilestonesService)
		want     *MergeRequest
		wantErr  bool
		errMsg   string
	}{
		{
			name: "successful create with username resolution",
			opts: &CreateMergeRequestOptions{
				SourceBranch:       "feature-branch",
				TargetBranch:       "main",
				Title:              "Test MR with usernames",
				Description:        "This is a test merge request",
				Assignees:          []interface{}{"alice", "bob"},
				Reviewers:          []interface{}{"charlie"},
				Labels:             []string{"enhancement"},
				Milestone:          "v1.0",
				RemoveSourceBranch: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, mrs *MockMergeRequestsService, users *MockUsersService, milestones *MockMilestonesService) {
				client.On("Projects").Return(projects)
				client.On("MergeRequests").Return(mrs)
				client.On("Users").Return(users)
				client.On("Milestones").Return(milestones)
				
				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)
				
				// Mock user lookups
				aliceUsername := "alice"
				users.On("ListUsers", &gitlab.ListUsersOptions{
					Username: &aliceUsername,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}).Return([]*gitlab.User{{ID: 10, Username: "alice"}}, &gitlab.Response{}, nil)
				
				bobUsername := "bob"
				users.On("ListUsers", &gitlab.ListUsersOptions{
					Username: &bobUsername,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}).Return([]*gitlab.User{{ID: 20, Username: "bob"}}, &gitlab.Response{}, nil)
				
				charlieUsername := "charlie"
				users.On("ListUsers", &gitlab.ListUsersOptions{
					Username: &charlieUsername,
					ListOptions: gitlab.ListOptions{PerPage: 1, Page: 1},
				}).Return([]*gitlab.User{{ID: 30, Username: "charlie"}}, &gitlab.Response{}, nil)
				
				// Mock milestone lookup
				state := "active"
				milestones.On("ListMilestones", 123, &gitlab.ListMilestonesOptions{
					State:       &state,
					ListOptions: gitlab.ListOptions{PerPage: maxMilestonesPerPage, Page: 1},
				}).Return([]*gitlab.Milestone{
					{ID: 100, Title: "v1.0"},
					{ID: 101, Title: "v2.0"},
				}, &gitlab.Response{}, nil)
				
				expectedOpts := &gitlab.CreateMergeRequestOptions{
					Title:              gitlab.Ptr("Test MR with usernames"),
					SourceBranch:       gitlab.Ptr("feature-branch"),
					TargetBranch:       gitlab.Ptr("main"),
					Description:        gitlab.Ptr("This is a test merge request"),
					AssigneeIDs:        &[]int{10, 20},
					ReviewerIDs:        &[]int{30},
					Labels:             (*gitlab.LabelOptions)(&[]string{"enhancement"}),
					MilestoneID:        gitlab.Ptr(100),
					RemoveSourceBranch: gitlab.Ptr(true),
				}
				
				mrs.On("CreateMergeRequest", 123, expectedOpts).Return(
					&gitlab.MergeRequest{
						BasicMergeRequest: gitlab.BasicMergeRequest{
							ID:           3,
							IID:          102,
							Title:        "Test MR with usernames",
							Description:  "This is a test merge request",
							State:        "opened",
							SourceBranch: "feature-branch",
							TargetBranch: "main",
							Author: &gitlab.BasicUser{
								ID:       1,
								Username: "testuser",
								Name:     "Test User",
							},
							Assignees: []*gitlab.BasicUser{
								{ID: 10, Username: "alice", Name: "Alice"},
								{ID: 20, Username: "bob", Name: "Bob"},
							},
							Reviewers: []*gitlab.BasicUser{
								{ID: 30, Username: "charlie", Name: "Charlie"},
							},
							Labels: []string{"enhancement"},
							Milestone: &gitlab.Milestone{
								ID:    100,
								Title: "v1.0",
							},
							WebURL:    "https://gitlab.com/test/project/-/merge_requests/102",
							Draft:     false,
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: &MergeRequest{
				ID:           3,
				IID:          102,
				Title:        "Test MR with usernames",
				Description:  "This is a test merge request",
				State:        "opened",
				SourceBranch: "feature-branch",
				TargetBranch: "main",
				Author: map[string]interface{}{
					"id":       1,
					"username": "testuser",
					"name":     "Test User",
				},
				Assignees: []map[string]interface{}{
					{"id": 10, "username": "alice", "name": "Alice"},
					{"id": 20, "username": "bob", "name": "Bob"},
				},
				Reviewers: []map[string]interface{}{
					{"id": 30, "username": "charlie", "name": "Charlie"},
				},
				Labels: []string{"enhancement"},
				Milestone: map[string]interface{}{
					"id":    100,
					"title": "v1.0",
				},
				WebURL:    "https://gitlab.com/test/project/-/merge_requests/102",
				Draft:     false,
				CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
			},
			wantErr: false,
		},
	}
	
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			
			mockClient := new(MockGitLabClient)
			mockProjects := new(MockProjectsService)
			mockMRs := new(MockMergeRequestsService)
			mockUsers := new(MockUsersService)
			mockMilestones := new(MockMilestonesService)
			
			if tc.setup != nil {
				tc.setup(mockClient, mockProjects, mockMRs, mockUsers, mockMilestones)
			}
			
			app := NewWithClient("test-token", "https://gitlab.com/", mockClient)
			
			got, err := app.CreateProjectMergeRequest("test/project", tc.opts)
			
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
			mockMRs.AssertExpectations(t)
			mockUsers.AssertExpectations(t)
			mockMilestones.AssertExpectations(t)
		})
	}
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
				projects.On("EditProject", 123, expectedOpts).Return(
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
				projects.On("EditProject", 123, expectedOpts).Return(
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
				projects.On("EditProject", 123, expectedOpts).Return(
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
				projects.On("EditProject", 123, expectedOpts).Return(
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
				projects.On("EditProject", 123, expectedOpts).Return(
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
