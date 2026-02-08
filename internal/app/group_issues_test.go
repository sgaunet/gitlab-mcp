package app

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/api/client-go"
)

func TestExtractGroupPath(t *testing.T) {
	tests := []struct {
		name        string
		projectPath string
		want        string
		wantErr     bool
	}{
		{
			name:        "valid two-level path",
			projectPath: "myorg/project",
			want:        "myorg",
			wantErr:     false,
		},
		{
			name:        "valid three-level path",
			projectPath: "myorg/team/project",
			want:        "myorg/team",
			wantErr:     false,
		},
		{
			name:        "valid four-level path",
			projectPath: "a/b/c/d",
			want:        "a/b/c",
			wantErr:     false,
		},
		{
			name:        "invalid single-level path",
			projectPath: "standalone",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "invalid empty path",
			projectPath: "",
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractGroupPath(tt.projectPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMergeIssues(t *testing.T) {
	testTime := time.Now()
	currentProjectID := int64(100)

	tests := []struct {
		name            string
		projectIssues   []*gitlab.Issue
		groupIssues     []*gitlab.Issue
		currentProjectID int64
		wantCount       int
		description     string
	}{
		{
			name:            "empty sets",
			projectIssues:   []*gitlab.Issue{},
			groupIssues:     []*gitlab.Issue{},
			currentProjectID: currentProjectID,
			wantCount:       0,
			description:     "both empty should return empty",
		},
		{
			name: "project only",
			projectIssues: []*gitlab.Issue{
				{ID: 1, IID: 1, ProjectID: 100, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
				{ID: 2, IID: 2, ProjectID: 100, Title: "P2", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			groupIssues:      []*gitlab.Issue{},
			currentProjectID: currentProjectID,
			wantCount:        2,
			description:      "only project issues should return all project issues",
		},
		{
			name:          "group only",
			projectIssues: []*gitlab.Issue{},
			groupIssues: []*gitlab.Issue{
				{ID: 10, IID: 10, ProjectID: 200, Title: "G1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
				{ID: 11, IID: 11, ProjectID: 201, Title: "G2", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			currentProjectID: currentProjectID,
			wantCount:        2,
			description:      "only group issues should return all group issues",
		},
		{
			name: "no overlap",
			projectIssues: []*gitlab.Issue{
				{ID: 1, IID: 1, ProjectID: 100, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			groupIssues: []*gitlab.Issue{
				{ID: 10, IID: 10, ProjectID: 200, Title: "G1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			currentProjectID: currentProjectID,
			wantCount:        2,
			description:      "no overlap should return all issues",
		},
		{
			name: "full overlap - same project",
			projectIssues: []*gitlab.Issue{
				{ID: 1, IID: 1, ProjectID: 100, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			groupIssues: []*gitlab.Issue{
				{ID: 1, IID: 1, ProjectID: 100, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			currentProjectID: currentProjectID,
			wantCount:        1,
			description:      "full overlap should deduplicate",
		},
		{
			name: "partial overlap",
			projectIssues: []*gitlab.Issue{
				{ID: 1, IID: 1, ProjectID: 100, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
				{ID: 2, IID: 2, ProjectID: 100, Title: "P2", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
			},
			groupIssues: []*gitlab.Issue{
				{ID: 1, IID: 1, ProjectID: 100, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime}, // duplicate
				{ID: 10, IID: 10, ProjectID: 200, Title: "G1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime}, // unique
			},
			currentProjectID: currentProjectID,
			wantCount:        3,
			description:      "partial overlap should deduplicate and merge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeIssues(tt.projectIssues, tt.groupIssues, tt.currentProjectID)
			assert.Equal(t, tt.wantCount, len(result), tt.description)
		})
	}
}

func TestListProjectIssuesWithGroupIssues(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name        string
		projectPath string
		opts        *ListIssuesOptions
		setup       func(*MockGitLabClient, *MockProjectsService, *MockIssuesService)
		wantCount   int
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "include_group_issues false - only project issues",
			projectPath: "myorg/project",
			opts: &ListIssuesOptions{
				State:              "opened",
				Limit:              100,
				IncludeGroupIssues: false,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "myorg/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				expectedOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), expectedOpts).Return(
					[]*gitlab.Issue{
						{ID: 1, IID: 1, ProjectID: 123, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
						{ID: 2, IID: 2, ProjectID: 123, Title: "P2", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:        "include_group_issues true - merged results",
			projectPath: "myorg/team/project",
			opts: &ListIssuesOptions{
				State:              "opened",
				Limit:              100,
				IncludeGroupIssues: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues).Times(2)

				projects.On("GetProject", "myorg/team/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				projectOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), projectOpts).Return(
					[]*gitlab.Issue{
						{ID: 1, IID: 1, ProjectID: 123, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)

				groupOpts := &gitlab.ListGroupIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListGroupIssues", "myorg/team", groupOpts, []gitlab.RequestOptionFunc(nil)).Return(
					[]*gitlab.Issue{
						{ID: 10, IID: 10, ProjectID: 456, Title: "G1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:        "group path extraction fails - graceful fallback",
			projectPath: "standalone",
			opts: &ListIssuesOptions{
				State:              "opened",
				Limit:              100,
				IncludeGroupIssues: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues)

				projects.On("GetProject", "standalone", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				projectOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), projectOpts).Return(
					[]*gitlab.Issue{
						{ID: 1, IID: 1, ProjectID: 123, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)
				// No group issues call expected - should fallback
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "list group issues fails - graceful fallback",
			projectPath: "myorg/project",
			opts: &ListIssuesOptions{
				State:              "opened",
				Limit:              100,
				IncludeGroupIssues: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues).Times(2)

				projects.On("GetProject", "myorg/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				projectOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), projectOpts).Return(
					[]*gitlab.Issue{
						{ID: 1, IID: 1, ProjectID: 123, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)

				groupOpts := &gitlab.ListGroupIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListGroupIssues", "myorg", groupOpts, []gitlab.RequestOptionFunc(nil)).Return(
					[]*gitlab.Issue(nil),
					(*gitlab.Response)(nil),
					errors.New("permission denied"),
				)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "empty project issues - group issues returned",
			projectPath: "myorg/project",
			opts: &ListIssuesOptions{
				State:              "opened",
				Limit:              100,
				IncludeGroupIssues: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues).Times(2)

				projects.On("GetProject", "myorg/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				projectOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), projectOpts).Return(
					[]*gitlab.Issue{},
					&gitlab.Response{}, nil,
				)

				groupOpts := &gitlab.ListGroupIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListGroupIssues", "myorg", groupOpts, []gitlab.RequestOptionFunc(nil)).Return(
					[]*gitlab.Issue{
						{ID: 10, IID: 10, ProjectID: 456, Title: "G1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "deduplication - project issue in group results",
			projectPath: "myorg/project",
			opts: &ListIssuesOptions{
				State:              "opened",
				Limit:              100,
				IncludeGroupIssues: true,
			},
			setup: func(client *MockGitLabClient, projects *MockProjectsService, issues *MockIssuesService) {
				client.On("Projects").Return(projects)
				client.On("Issues").Return(issues).Times(2)

				projects.On("GetProject", "myorg/project", (*gitlab.GetProjectOptions)(nil)).Return(
					&gitlab.Project{ID: 123}, &gitlab.Response{}, nil,
				)

				projectOpts := &gitlab.ListProjectIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListProjectIssues", int64(123), projectOpts).Return(
					[]*gitlab.Issue{
						{ID: 1, IID: 1, ProjectID: 123, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
						{ID: 2, IID: 2, ProjectID: 123, Title: "P2", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime},
					},
					&gitlab.Response{}, nil,
				)

				groupOpts := &gitlab.ListGroupIssuesOptions{
					State:       gitlab.Ptr("opened"),
					ListOptions: gitlab.ListOptions{PerPage: 100, Page: 1},
				}

				issues.On("ListGroupIssues", "myorg", groupOpts, []gitlab.RequestOptionFunc(nil)).Return(
					[]*gitlab.Issue{
						{ID: 1, IID: 1, ProjectID: 123, Title: "P1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime}, // duplicate
						{ID: 10, IID: 10, ProjectID: 456, Title: "G1", State: "opened", CreatedAt: &testTime, UpdatedAt: &testTime}, // unique
					},
					&gitlab.Response{}, nil,
				)
			},
			wantCount: 3, // P1, P2, G1 (duplicate P1 from group excluded)
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockProjects := &MockProjectsService{}
			mockIssues := &MockIssuesService{}

			tt.setup(mockClient, mockProjects, mockIssues)

			app := NewWithClient("token", "https://gitlab.com/", mockClient)

			result, err := app.ListProjectIssues(tt.projectPath, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(result))
			}

			mockClient.AssertExpectations(t)
			mockProjects.AssertExpectations(t)
			mockIssues.AssertExpectations(t)
		})
	}
}
