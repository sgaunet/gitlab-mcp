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

// TestConvertGitLabEpicIssueAssignment tests the convertGitLabEpicIssueAssignment function.
func TestConvertGitLabEpicIssueAssignment(t *testing.T) {
	tests := []struct {
		name  string
		input *gitlab.EpicIssueAssignment
		want  EpicIssueAssignment
	}{
		{
			name: "full conversion with author",
			input: &gitlab.EpicIssueAssignment{
				Issue: &gitlab.Issue{
					ID:          123,
					IID:         10,
					Title:       "Test Issue",
					Description: "Test Description",
					State:       "opened",
					WebURL:      "https://gitlab.com/test/project/-/issues/10",
					Labels:      []string{"bug", "high-priority"},
					Author: &gitlab.IssueAuthor{
						ID:       456,
						Username: "testuser",
						Name:     "Test User",
					},
				},
				Epic: &gitlab.Epic{
					ID:  789,
					IID: 5,
				},
			},
			want: EpicIssueAssignment{
				ID:          123,
				IID:         10,
				EpicID:      789,
				EpicIID:     5,
				Title:       "Test Issue",
				Description: "Test Description",
				State:       "opened",
				WebURL:      "https://gitlab.com/test/project/-/issues/10",
				Labels:      []string{"bug", "high-priority"},
				Author: map[string]any{
					"id":       int64(456),
					"username": "testuser",
					"name":     "Test User",
				},
			},
		},
		{
			name: "conversion with nil author",
			input: &gitlab.EpicIssueAssignment{
				Issue: &gitlab.Issue{
					ID:          123,
					IID:         10,
					Title:       "Test Issue",
					Description: "Test Description",
					State:       "opened",
					WebURL:      "https://gitlab.com/test/project/-/issues/10",
					Labels:      []string{},
					Author:      nil,
				},
				Epic: &gitlab.Epic{
					ID:  789,
					IID: 5,
				},
			},
			want: EpicIssueAssignment{
				ID:          123,
				IID:         10,
				EpicID:      789,
				EpicIID:     5,
				Title:       "Test Issue",
				Description: "Test Description",
				State:       "opened",
				WebURL:      "https://gitlab.com/test/project/-/issues/10",
				Labels:      []string{},
				Author:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertGitLabEpicIssueAssignment(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestApp_SetDefaultEpicOptions tests the setDefaultEpicOptions helper function.
func TestApp_SetDefaultEpicOptions(t *testing.T) {
	tests := []struct {
		name  string
		input *ListEpicsOptions
		want  *ListEpicsOptions
	}{
		{
			name:  "nil options returns defaults",
			input: nil,
			want: &ListEpicsOptions{
				State: "opened",
				Limit: 100,
			},
		},
		{
			name: "empty state sets default",
			input: &ListEpicsOptions{
				State: "",
				Limit: 50,
			},
			want: &ListEpicsOptions{
				State: "opened",
				Limit: 50,
			},
		},
		{
			name: "zero limit sets default",
			input: &ListEpicsOptions{
				State: "closed",
				Limit: 0,
			},
			want: &ListEpicsOptions{
				State: "closed",
				Limit: 100,
			},
		},
		{
			name: "limit exceeds max gets capped",
			input: &ListEpicsOptions{
				State: "opened",
				Limit: 200,
			},
			want: &ListEpicsOptions{
				State: "opened",
				Limit: 100,
			},
		},
		{
			name: "valid options unchanged",
			input: &ListEpicsOptions{
				State: "all",
				Limit: 50,
			},
			want: &ListEpicsOptions{
				State: "all",
				Limit: 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			app := NewWithClient("token", "https://gitlab.com/", mockClient)

			got := app.setDefaultEpicOptions(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestApp_ValidateAddIssueToEpicOptions tests the validateAddIssueToEpicOptions function.
func TestApp_ValidateAddIssueToEpicOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    *AddIssueToEpicOptions
		wantErr bool
		errType error
	}{
		{
			name: "valid options",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
			errType: ErrCreateOptionsRequired,
		},
		{
			name: "empty group path",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			wantErr: true,
			errType: ErrGroupPathRequired,
		},
		{
			name: "zero epic IID",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     0,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			wantErr: true,
			errType: ErrEpicIIDRequired,
		},
		{
			name: "negative epic IID",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     -1,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			wantErr: true,
			errType: ErrEpicIIDRequired,
		},
		{
			name: "empty project path",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "",
				IssueIID:    10,
			},
			wantErr: true,
			errType: ErrProjectPathRequired,
		},
		{
			name: "zero issue IID",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    0,
			},
			wantErr: true,
			errType: ErrInvalidIssueIID,
		},
		{
			name: "negative issue IID",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    -1,
			},
			wantErr: true,
			errType: ErrInvalidIssueIID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			app := NewWithClient("token", "https://gitlab.com/", mockClient)

			err := app.validateAddIssueToEpicOptions(tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestApp_ListGroupEpics tests the ListGroupEpics function.
func TestApp_ListGroupEpics(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name      string
		groupPath string
		opts      *ListEpicsOptions
		setup     func(*MockGitLabClient, *MockGroupsService, *MockEpicsService)
		want      []Epic
		wantErr   bool
		errType   error
	}{
		{
			name:      "success with defaults",
			groupPath: "test/group",
			opts:      nil,
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListGroupEpicsOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				epics.On("ListGroupEpics", int64(456), expectedOpts).Return(
					[]*gitlab.Epic{
						{
							ID:          1,
							IID:         10,
							GroupID:     456,
							Title:       "Q1 2024 Launch",
							Description: "Epic description",
							State:       "opened",
							WebURL:      "https://gitlab.com/groups/test/group/-/epics/10",
							Labels:      []string{"roadmap", "high-priority"},
							CreatedAt:   &testTime,
							UpdatedAt:   &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []Epic{
				{
					ID:          1,
					IID:         10,
					GroupID:     456,
					Title:       "Q1 2024 Launch",
					Description: "Epic description",
					State:       "opened",
					WebURL:      "https://gitlab.com/groups/test/group/-/epics/10",
					Labels:      []string{"roadmap", "high-priority"},
					CreatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
					UpdatedAt:   testTime.Format("2006-01-02T15:04:05Z"),
				},
			},
			wantErr: false,
		},
		{
			name:      "success with custom options",
			groupPath: "test/group",
			opts: &ListEpicsOptions{
				State: "closed",
				Limit: 50,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListGroupEpicsOptions{
					State:       gitlab.Ptr("closed"),
					ListOptions: gitlab.ListOptions{PerPage: 50, Page: 1},
				}

				epics.On("ListGroupEpics", int64(456), expectedOpts).Return(
					[]*gitlab.Epic{
						{
							ID:        2,
							IID:       20,
							GroupID:   456,
							Title:     "Completed Epic",
							State:     "closed",
							CreatedAt: &testTime,
							UpdatedAt: &testTime,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []Epic{
				{
					ID:        2,
					IID:       20,
					GroupID:   456,
					Title:     "Completed Epic",
					State:     "closed",
					CreatedAt: testTime.Format("2006-01-02T15:04:05Z"),
					UpdatedAt: testTime.Format("2006-01-02T15:04:05Z"),
				},
			},
			wantErr: false,
		},
		{
			name:      "empty results",
			groupPath: "test/group",
			opts:      nil,
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListGroupEpicsOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				epics.On("ListGroupEpics", int64(456), expectedOpts).Return(
					[]*gitlab.Epic{}, &gitlab.Response{}, nil,
				)
			},
			want:    []Epic{},
			wantErr: false,
		},
		{
			name:      "group not found",
			groupPath: "nonexistent/group",
			opts:      nil,
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)

				groups.On("GetGroup", "nonexistent/group", (*gitlab.GetGroupOptions)(nil)).Return(
					(*gitlab.Group)(nil), &gitlab.Response{}, errors.New("404 Not Found"),
				)
			},
			wantErr: true,
		},
		{
			name:      "tier required (403)",
			groupPath: "test/group",
			opts:      nil,
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListGroupEpicsOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				epics.On("ListGroupEpics", int64(456), expectedOpts).Return(
					([]*gitlab.Epic)(nil), &gitlab.Response{}, errors.New("403 Forbidden"),
				)
			},
			wantErr: true,
			errType: ErrEpicsTierRequired,
		},
		{
			name:      "API error",
			groupPath: "test/group",
			opts:      nil,
			setup: func(client *MockGitLabClient, groups *MockGroupsService, epics *MockEpicsService) {
				client.On("Groups").Return(groups)
				client.On("Epics").Return(epics)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListGroupEpicsOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				epics.On("ListGroupEpics", int64(456), expectedOpts).Return(
					([]*gitlab.Epic)(nil), &gitlab.Response{}, errors.New("500 Internal Server Error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockGroups := &MockGroupsService{}
			mockEpics := &MockEpicsService{}

			if tt.setup != nil {
				tt.setup(mockClient, mockGroups, mockEpics)
			}

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			got, err := app.ListGroupEpics(tt.groupPath, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockClient.AssertExpectations(t)
			mockGroups.AssertExpectations(t)
			mockEpics.AssertExpectations(t)
		})
	}
}

// TestApp_AddIssueToEpic tests the AddIssueToEpic function.
func TestApp_AddIssueToEpic(t *testing.T) {
	tests := []struct {
		name    string
		opts    *AddIssueToEpicOptions
		setup   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService)
		want    *EpicIssueAssignment
		wantErr bool
		errType error
	}{
		{
			name: "success",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)
				client.On("Projects").Return(projects).Times(1)
				client.On("Issues").Return(issues).Times(1)
				client.On("EpicIssues").Return(epicIssues).Times(1)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				).Times(1)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(1)

				issues.On("GetIssue", int64(123), 10).Return(
					&gitlab.Issue{ID: 999, IID: 10}, &gitlab.Response{}, nil,
				).Times(1)

				epicIssues.On("AssignEpicIssue", int64(456), int64(5), int64(999)).Return(
					&gitlab.EpicIssueAssignment{
						Issue: &gitlab.Issue{
							ID:          999,
							IID:         10,
							Title:       "Test Issue",
							Description: "Test Description",
							State:       "opened",
							WebURL:      "https://gitlab.com/test/project/-/issues/10",
							Labels:      []string{"bug"},
							Author: &gitlab.IssueAuthor{
								ID:       777,
								Username: "testuser",
								Name:     "Test User",
							},
						},
						Epic: &gitlab.Epic{
							ID:  888,
							IID: 5,
						},
					},
					&gitlab.Response{}, nil,
				).Times(1)
			},
			want: &EpicIssueAssignment{
				ID:          999,
				IID:         10,
				EpicID:      888,
				EpicIID:     5,
				Title:       "Test Issue",
				Description: "Test Description",
				State:       "opened",
				WebURL:      "https://gitlab.com/test/project/-/issues/10",
				Labels:      []string{"bug"},
				Author: map[string]any{
					"id":       int64(777),
					"username": "testuser",
					"name":     "Test User",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil options",
			opts:    nil,
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrCreateOptionsRequired,
		},
		{
			name: "empty group path",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrGroupPathRequired,
		},
		{
			name: "invalid epic IID (zero)",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     0,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrEpicIIDRequired,
		},
		{
			name: "invalid epic IID (negative)",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     -1,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrEpicIIDRequired,
		},
		{
			name: "empty project path",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "",
				IssueIID:    10,
			},
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrProjectPathRequired,
		},
		{
			name: "invalid issue IID (zero)",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    0,
			},
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrInvalidIssueIID,
		},
		{
			name: "invalid issue IID (negative)",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    -1,
			},
			setup:   func(*MockGitLabClient, *MockGroupsService, *MockProjectsService, *MockIssuesService, *MockEpicIssuesService) {},
			wantErr: true,
			errType: ErrInvalidIssueIID,
		},
		{
			name: "group not found",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "nonexistent/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)

				groups.On("GetGroup", "nonexistent/group", (*gitlab.GetGroupOptions)(nil)).Return(
					(*gitlab.Group)(nil), &gitlab.Response{}, errors.New("404 Not Found"),
				).Times(1)
			},
			wantErr: true,
		},
		{
			name: "project not found",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "nonexistent/project",
				IssueIID:    10,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)
				client.On("Projects").Return(projects).Times(1)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				).Times(1)

				projects.On("GetProject", "nonexistent/project", (*gitlab.GetProjectOptions)(nil)).Return(
					(*gitlab.Project)(nil), &gitlab.Response{}, errors.New("404 Not Found"),
				).Times(1)
			},
			wantErr: true,
		},
		{
			name: "issue not found",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    999,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)
				client.On("Projects").Return(projects).Times(1)
				client.On("Issues").Return(issues).Times(1)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				).Times(1)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(1)

				issues.On("GetIssue", int64(123), 999).Return(
					(*gitlab.Issue)(nil), &gitlab.Response{}, errors.New("404 Not Found"),
				).Times(1)
			},
			wantErr: true,
		},
		{
			name: "issue ID is zero",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)
				client.On("Projects").Return(projects).Times(1)
				client.On("Issues").Return(issues).Times(1)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				).Times(1)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(1)

				issues.On("GetIssue", int64(123), 10).Return(
					&gitlab.Issue{ID: 0, IID: 10}, &gitlab.Response{}, nil,
				).Times(1)
			},
			wantErr: true,
			errType: ErrIssueNotFound,
		},
		{
			name: "AssignEpicIssue fails",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)
				client.On("Projects").Return(projects).Times(1)
				client.On("Issues").Return(issues).Times(1)
				client.On("EpicIssues").Return(epicIssues).Times(1)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				).Times(1)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(1)

				issues.On("GetIssue", int64(123), 10).Return(
					&gitlab.Issue{ID: 999, IID: 10}, &gitlab.Response{}, nil,
				).Times(1)

				epicIssues.On("AssignEpicIssue", int64(456), int64(5), int64(999)).Return(
					(*gitlab.EpicIssueAssignment)(nil), &gitlab.Response{}, errors.New("500 Internal Server Error"),
				).Times(1)
			},
			wantErr: true,
		},
		{
			name: "tier required (403)",
			opts: &AddIssueToEpicOptions{
				GroupPath:   "test/group",
				EpicIID:     5,
				ProjectPath: "test/project",
				IssueIID:    10,
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService, projects *MockProjectsService, issues *MockIssuesService, epicIssues *MockEpicIssuesService) {
				client.On("Groups").Return(groups).Times(1)
				client.On("Projects").Return(projects).Times(1)
				client.On("Issues").Return(issues).Times(1)
				client.On("EpicIssues").Return(epicIssues).Times(1)

				groups.On("GetGroup", "test/group", (*gitlab.GetGroupOptions)(nil)).Return(
					&gitlab.Group{ID: 456}, &gitlab.Response{}, nil,
				).Times(1)

				projects.On("GetProject", "test/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				).Times(1)

				issues.On("GetIssue", int64(123), 10).Return(
					&gitlab.Issue{ID: 999, IID: 10}, &gitlab.Response{}, nil,
				).Times(1)

				epicIssues.On("AssignEpicIssue", int64(456), int64(5), int64(999)).Return(
					(*gitlab.EpicIssueAssignment)(nil), &gitlab.Response{}, errors.New("403 Forbidden"),
				).Times(1)
			},
			wantErr: true,
			errType: ErrEpicsTierRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockGroups := &MockGroupsService{}
			mockProjects := &MockProjectsService{}
			mockIssues := &MockIssuesService{}
			mockEpicIssues := &MockEpicIssuesService{}

			if tt.setup != nil {
				tt.setup(mockClient, mockGroups, mockProjects, mockIssues, mockEpicIssues)
			}

			app := NewWithClient("token", "https://gitlab.com/", mockClient)
			app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

			got, err := app.AddIssueToEpic(tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockClient.AssertExpectations(t)
			mockGroups.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockIssues.AssertExpectations(t)
			mockEpicIssues.AssertExpectations(t)
		})
	}
}
