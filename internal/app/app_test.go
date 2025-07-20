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