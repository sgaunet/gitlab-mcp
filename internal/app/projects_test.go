package app

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// newTestAppWithGroups builds an App wired to a mock Groups service.
func newTestAppWithGroups(mockClient *MockGitLabClient) *App {
	app := NewWithClient("token", "https://gitlab.com/", mockClient)
	app.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	return app
}

// expectedProjectOpts builds the GitLab options the app is expected to send.
func expectedProjectOpts(page, perPage int64, subgroups bool) *gitlab.ListGroupProjectsOptions {
	opts := &gitlab.ListGroupProjectsOptions{
		ListOptions:      gitlab.ListOptions{PerPage: perPage, Page: page},
		IncludeSubGroups: new(subgroups),
		Archived:         new(false),
		OrderBy:          new(projectsOrderBy),
		Sort:             new(projectsSort),
	}
	return opts
}

// TestNormalizeListGroupProjectsOptions verifies defaults and clamping, and that the
// caller's options struct is never mutated.
func TestNormalizeListGroupProjectsOptions(t *testing.T) {
	t.Run("nil options get defaults", func(t *testing.T) {
		got := normalizeListGroupProjectsOptions(nil)
		assert.True(t, got.IncludeSubgroups)
		assert.Equal(t, int64(maxProjectsPerPage), got.Limit)
		assert.False(t, got.IncludeArchived)
	})

	t.Run("zero limit gets default", func(t *testing.T) {
		got := normalizeListGroupProjectsOptions(&ListGroupProjectsOptions{IncludeSubgroups: true})
		assert.Equal(t, int64(maxProjectsPerPage), got.Limit)
	})

	t.Run("excessive limit is clamped", func(t *testing.T) {
		got := normalizeListGroupProjectsOptions(&ListGroupProjectsOptions{Limit: 99999})
		assert.Equal(t, int64(maxProjectsTotal), got.Limit)
	})

	t.Run("caller options are not mutated", func(t *testing.T) {
		opts := &ListGroupProjectsOptions{Limit: 0, IncludeSubgroups: false}
		_ = normalizeListGroupProjectsOptions(opts)
		assert.Equal(t, int64(0), opts.Limit, "input Limit must stay untouched")
		assert.False(t, opts.IncludeSubgroups)
	})
}

// TestBuildListGroupProjectsOptions verifies the mapping onto GitLab API options.
func TestBuildListGroupProjectsOptions(t *testing.T) {
	t.Run("archived excluded by default", func(t *testing.T) {
		got := buildListGroupProjectsOptions(ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 100})
		require.NotNil(t, got.Archived)
		assert.False(t, *got.Archived)
		require.NotNil(t, got.IncludeSubGroups)
		assert.True(t, *got.IncludeSubGroups)
		assert.Equal(t, int64(100), got.PerPage)
		assert.Equal(t, int64(1), got.Page)
		assert.Nil(t, got.Search)
		assert.Nil(t, got.Topic)
	})

	t.Run("include_archived omits the archived filter", func(t *testing.T) {
		got := buildListGroupProjectsOptions(ListGroupProjectsOptions{IncludeArchived: true, Limit: 100})
		assert.Nil(t, got.Archived, "Archived must be omitted so GitLab returns both kinds")
	})

	t.Run("per_page never exceeds the API ceiling", func(t *testing.T) {
		got := buildListGroupProjectsOptions(ListGroupProjectsOptions{Limit: maxProjectsTotal})
		assert.Equal(t, int64(maxProjectsPerPage), got.PerPage)
	})

	t.Run("per_page shrinks to a small limit", func(t *testing.T) {
		got := buildListGroupProjectsOptions(ListGroupProjectsOptions{Limit: 5})
		assert.Equal(t, int64(5), got.PerPage)
	})

	t.Run("search and topic are passed through", func(t *testing.T) {
		got := buildListGroupProjectsOptions(ListGroupProjectsOptions{
			Limit: 100, Search: "api", Topic: "golang",
		})
		require.NotNil(t, got.Search)
		require.NotNil(t, got.Topic)
		assert.Equal(t, "api", *got.Search)
		assert.Equal(t, "golang", *got.Topic)
	})

	t.Run("include_subgroups false is sent explicitly", func(t *testing.T) {
		got := buildListGroupProjectsOptions(ListGroupProjectsOptions{Limit: 100})
		require.NotNil(t, got.IncludeSubGroups)
		assert.False(t, *got.IncludeSubGroups)
	})
}

// TestConvertGitLabProject verifies the GitLab -> GroupProject mapping.
func TestConvertGitLabProject(t *testing.T) {
	got := convertGitLabProject(&gitlab.Project{
		ID:                42,
		Name:              "my-service",
		Path:              "my-service",
		PathWithNamespace: "myorg/team/my-service",
		Description:       "Payment gateway",
		Topics:            []string{"go", "api"},
		WebURL:            "https://gitlab.com/myorg/team/my-service",
		Archived:          true,
	})

	assert.Equal(t, GroupProject{
		ID:          42,
		Name:        "my-service",
		Path:        "myorg/team/my-service",
		Description: "Payment gateway",
		Topics:      []string{"go", "api"},
		WebURL:      "https://gitlab.com/myorg/team/my-service",
		Archived:    true,
	}, got, "Path must carry the namespaced path so it can feed the other tools")
}

// TestApp_ListGroupProjects tests the ListGroupProjects method.
func TestApp_ListGroupProjects(t *testing.T) {
	apiError := errors.New("gitlab client: 404 group not found")

	tests := []struct {
		name      string
		groupPath string
		opts      *ListGroupProjectsOptions
		setup     func(*MockGitLabClient, *MockGroupsService)
		want      []GroupProject
		wantErr   bool
		errType   error
	}{
		{
			name:      "successful listing with defaults",
			groupPath: "myorg",
			opts:      nil,
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 100, true)).Return(
					[]*gitlab.Project{
						{
							ID:                1,
							Name:              "alpha",
							PathWithNamespace: "myorg/alpha",
							Description:       "First project",
							Topics:            []string{"go"},
							WebURL:            "https://gitlab.com/myorg/alpha",
						},
						{
							ID:                2,
							Name:              "beta",
							PathWithNamespace: "myorg/sub/beta",
							Description:       "Nested project",
							Topics:            []string{"python", "api"},
							WebURL:            "https://gitlab.com/myorg/sub/beta",
							Archived:          true,
						},
					},
					&gitlab.Response{}, nil,
				)
			},
			want: []GroupProject{
				{
					ID: 1, Name: "alpha", Path: "myorg/alpha", Description: "First project",
					Topics: []string{"go"}, WebURL: "https://gitlab.com/myorg/alpha",
				},
				{
					ID: 2, Name: "beta", Path: "myorg/sub/beta", Description: "Nested project",
					Topics: []string{"python", "api"}, WebURL: "https://gitlab.com/myorg/sub/beta", Archived: true,
				},
			},
		},
		{
			name:      "paginates across pages",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 250},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				// Page 1 -> NextPage 2. PerPage must stay constant across pages.
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 100, true)).Return(
					[]*gitlab.Project{{ID: 1, Name: "a", PathWithNamespace: "myorg/a"}},
					&gitlab.Response{NextPage: 2}, nil,
				).Once()
				// Page 2 -> NextPage 0, loop terminates.
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(2, 100, true)).Return(
					[]*gitlab.Project{{ID: 2, Name: "b", PathWithNamespace: "myorg/b"}},
					&gitlab.Response{NextPage: 0}, nil,
				).Once()
			},
			want: []GroupProject{
				{ID: 1, Name: "a", Path: "myorg/a"},
				{ID: 2, Name: "b", Path: "myorg/b"},
			},
		},
		{
			name:      "stops paginating once the limit is reached",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 1},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				// Only one call must happen even though NextPage says more exist.
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 1, true)).Return(
					[]*gitlab.Project{{ID: 1, Name: "a", PathWithNamespace: "myorg/a"}},
					&gitlab.Response{NextPage: 2}, nil,
				).Once()
			},
			want: []GroupProject{{ID: 1, Name: "a", Path: "myorg/a"}},
		},
		{
			name:      "truncates a page larger than the limit",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 2},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 2, true)).Return(
					[]*gitlab.Project{
						{ID: 1, Name: "a", PathWithNamespace: "myorg/a"},
						{ID: 2, Name: "b", PathWithNamespace: "myorg/b"},
						{ID: 3, Name: "c", PathWithNamespace: "myorg/c"},
					},
					&gitlab.Response{}, nil,
				).Once()
			},
			want: []GroupProject{
				{ID: 1, Name: "a", Path: "myorg/a"},
				{ID: 2, Name: "b", Path: "myorg/b"},
			},
		},
		{
			name:      "include_subgroups false",
			groupPath: "myorg/team",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: false, Limit: 100},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				groups.On("ListGroupProjects", "myorg/team", expectedProjectOpts(1, 100, false)).Return(
					[]*gitlab.Project{{ID: 7, Name: "direct", PathWithNamespace: "myorg/team/direct"}},
					&gitlab.Response{}, nil,
				)
			},
			want: []GroupProject{{ID: 7, Name: "direct", Path: "myorg/team/direct"}},
		},
		{
			name:      "include_archived sends no archived filter",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, IncludeArchived: true, Limit: 100},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				opts := expectedProjectOpts(1, 100, true)
				opts.Archived = nil
				groups.On("ListGroupProjects", "myorg", opts).Return(
					[]*gitlab.Project{{ID: 9, Name: "old", PathWithNamespace: "myorg/old", Archived: true}},
					&gitlab.Response{}, nil,
				)
			},
			want: []GroupProject{{ID: 9, Name: "old", Path: "myorg/old", Archived: true}},
		},
		{
			name:      "search and topic filters",
			groupPath: "myorg",
			opts: &ListGroupProjectsOptions{
				IncludeSubgroups: true, Limit: 100, Search: "api", Topic: "golang",
			},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				opts := expectedProjectOpts(1, 100, true)
				opts.Search = new("api")
				opts.Topic = new("golang")
				groups.On("ListGroupProjects", "myorg", opts).Return(
					[]*gitlab.Project{{ID: 3, Name: "api-gw", PathWithNamespace: "myorg/api-gw"}},
					&gitlab.Response{}, nil,
				)
			},
			want: []GroupProject{{ID: 3, Name: "api-gw", Path: "myorg/api-gw"}},
		},
		{
			name:      "empty group returns an empty slice",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 100},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 100, true)).Return(
					[]*gitlab.Project{}, &gitlab.Response{}, nil,
				)
			},
			want: []GroupProject{},
		},
		{
			name:      "nil response does not loop or panic",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 100},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 100, true)).Return(
					[]*gitlab.Project{{ID: 1, Name: "a", PathWithNamespace: "myorg/a"}},
					(*gitlab.Response)(nil), nil,
				).Once()
			},
			want: []GroupProject{{ID: 1, Name: "a", Path: "myorg/a"}},
		},
		{
			name:      "empty group path",
			groupPath: "",
			opts:      nil,
			setup:     nil,
			wantErr:   true,
			errType:   ErrGroupPathRequired,
		},
		{
			name:      "API error is wrapped",
			groupPath: "myorg",
			opts:      &ListGroupProjectsOptions{IncludeSubgroups: true, Limit: 100},
			setup: func(client *MockGitLabClient, groups *MockGroupsService) {
				client.On("Groups").Return(groups)
				groups.On("ListGroupProjects", "myorg", expectedProjectOpts(1, 100, true)).Return(
					([]*gitlab.Project)(nil), (*gitlab.Response)(nil), apiError,
				)
			},
			wantErr: true,
			errType: apiError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitLabClient{}
			mockGroups := &MockGroupsService{}

			if tt.setup != nil {
				tt.setup(mockClient, mockGroups)
			}

			app := newTestAppWithGroups(mockClient)

			got, err := app.ListGroupProjects(tt.groupPath, tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					require.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockClient.AssertExpectations(t)
			mockGroups.AssertExpectations(t)
		})
	}
}
