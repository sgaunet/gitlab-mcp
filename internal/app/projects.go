package app

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	maxProjectsPerPage = 100  // GitLab API page-size ceiling.
	maxProjectsTotal   = 1000 // Hard cap on the number of projects returned in one call.

	// Ordering used for group project listings, for deterministic pagination.
	projectsOrderBy = "path"
	projectsSort    = "asc"
)

// ListGroupProjectsOptions holds the filters accepted by ListGroupProjects.
type ListGroupProjectsOptions struct {
	IncludeSubgroups bool   // Recurse into every descendant subgroup.
	IncludeArchived  bool   // Include archived projects in the results.
	Search           string // Filter by name/path substring.
	Topic            string // Filter by topic/tag.
	Limit            int64  // Maximum number of projects to return.
}

// GroupProject represents a project discovered under a group.
type GroupProject struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
	WebURL      string   `json:"web_url"`
	Archived    bool     `json:"archived"`
}

// normalizeListGroupProjectsOptions returns a copy of opts with defaults applied.
// The caller's struct is never modified.
func normalizeListGroupProjectsOptions(opts *ListGroupProjectsOptions) ListGroupProjectsOptions {
	if opts == nil {
		return ListGroupProjectsOptions{
			IncludeSubgroups: true,
			Limit:            maxProjectsPerPage,
		}
	}

	normalized := *opts
	if normalized.Limit <= 0 {
		normalized.Limit = maxProjectsPerPage
	}
	if normalized.Limit > maxProjectsTotal {
		normalized.Limit = maxProjectsTotal
	}

	return normalized
}

// buildListGroupProjectsOptions converts our options to GitLab API options.
func buildListGroupProjectsOptions(opts ListGroupProjectsOptions) *gitlab.ListGroupProjectsOptions {
	listOpts := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: min(opts.Limit, maxProjectsPerPage),
			Page:    1,
		},
		IncludeSubGroups: new(opts.IncludeSubgroups),
		OrderBy:          new(projectsOrderBy),
		Sort:             new(projectsSort),
	}

	// Leaving Archived unset returns both archived and active projects, which is
	// what include_archived=true means. Setting it to false excludes archived ones.
	if !opts.IncludeArchived {
		listOpts.Archived = new(false)
	}

	if opts.Search != "" {
		listOpts.Search = new(opts.Search)
	}
	if opts.Topic != "" {
		listOpts.Topic = new(opts.Topic)
	}

	return listOpts
}

// convertGitLabProject converts a GitLab project to our GroupProject representation.
func convertGitLabProject(project *gitlab.Project) GroupProject {
	return GroupProject{
		ID:          project.ID,
		Name:        project.Name,
		Path:        project.PathWithNamespace,
		Description: project.Description,
		Topics:      project.Topics,
		WebURL:      project.WebURL,
		Archived:    project.Archived,
	}
}

// hasMoreProjectPages reports whether another page should be requested.
func hasMoreProjectPages(batchSize int, resp *gitlab.Response, collected, limit int64) bool {
	if batchSize == 0 || collected >= limit {
		return false
	}
	// resp is nil when the underlying client wrapper drops it; stop rather than panic.
	return resp != nil && resp.NextPage > 0
}

// ListGroupProjects lists all projects of a GitLab group, recursing through subgroups by default.
// The group path is passed directly to the API, so no group ID resolution is required.
func (a *App) ListGroupProjects(groupPath string, opts *ListGroupProjectsOptions) ([]GroupProject, error) {
	a.logger.Debug("Listing projects for group", "group_path", groupPath, "options", opts)

	if groupPath == "" {
		return nil, ErrGroupPathRequired
	}

	normalized := normalizeListGroupProjectsOptions(opts)
	listOpts := buildListGroupProjectsOptions(normalized)

	result := make([]GroupProject, 0, normalized.Limit)
	for {
		projects, resp, err := a.client.Groups().ListGroupProjects(groupPath, listOpts)
		if err != nil {
			a.logger.Error("Failed to list group projects", "error", err, "group_path", groupPath)
			return nil, fmt.Errorf("failed to list projects for group %s: %w", groupPath, err)
		}

		for _, project := range projects {
			result = append(result, convertGitLabProject(project))
		}

		if !hasMoreProjectPages(len(projects), resp, int64(len(result)), normalized.Limit) {
			break
		}
		listOpts.Page = resp.NextPage
	}

	if int64(len(result)) > normalized.Limit {
		result = result[:normalized.Limit]
	}

	a.logger.Info("Successfully retrieved group projects",
		"count", len(result),
		"group_path", groupPath,
		"include_subgroups", normalized.IncludeSubgroups)

	return result, nil
}
