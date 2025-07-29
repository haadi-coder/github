package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// RepositoriesService provides access to repository-related API methods.
type RepositoriesService struct {
	client *Client
}

// Repository represents a GitHub repository.
// GitHub API docs: https://docs.github.com/en/rest/repos/repos
type Repository struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Fullname        string     `json:"full_name"`
	Owner           *User      `json:"owner"`
	Private         bool       `json:"private"`
	HtmlURL         string     `json:"html_url"`
	Description     string     `json:"description"`
	Fork            bool       `json:"fork"`
	URL             string     `json:"url"`
	CloneURL        string     `json:"clone_url"`
	MirrorURL       string     `json:"mirror_url"`
	Language        string     `json:"language"`
	ForksCount      int        `json:"forks_count"`
	StargazersCount int        `json:"stargazers_count"`
	WatchersCount   int        `json:"watchers_count"`
	Size            int        `json:"size"`
	DefaultBranch   string     `json:"default_branch"`
	OpenIssuesCount int        `json:"open_issues_count"`
	IsTemplate      bool       `json:"is_template"`
	Topics          []string   `json:"topics"`
	HasIssues       bool       `json:"has_issues"`
	HasProjects     bool       `json:"has_projects"`
	HasWiki         bool       `json:"has_wiki"`
	HasPages        bool       `json:"has_pages"`
	HasDownloads    bool       `json:"has_downloads"`
	Archived        bool       `json:"archived"`
	Disabled        bool       `json:"disabled"`
	Visibility      string     `json:"visibility"`
	PushedAt        *Timestamp `json:"pushed_at"`
	CreatedAt       *Timestamp `json:"created_at"`
	UpdatedAt       *Timestamp `json:"updated_at"`
	Permissions     struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	}
}

// Get fetches a repository by its owner and name.
// This method retrieves detailed information about a specific repository,
// including its metadata, statistics, and permissions for the authenticated user.
func (s *RepositoriesService) Get(ctx context.Context, owner string, repo string) (*Repository, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	r := new(Repository)
	resp, err := s.client.Do(ctx, req, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// RepositoryUpdateRequest represents the request body for updating a repository.
// GitHub API docs: https://docs.github.com/en/rest/repos/repos#update-a-repository
type RepositoryUpdateRequest struct {
	Name                      string `json:"name,omitempty"`
	Description               string `json:"description,omitempty"`
	Homepage                  string `json:"homepage,omitempty"`
	Private                   bool   `json:"private,omitempty"`
	Visibility                string `json:"visibility,omitempty"`
	HasIssues                 bool   `json:"has_issues,omitempty"`
	HasProjects               bool   `json:"has_projects,omitempty"`
	HasWiki                   bool   `json:"has_wiki,omitempty"`
	IsTemplate                bool   `json:"is_template,omitempty"`
	DefaultBranch             string `json:"default_branch,omitempty"`
	AllowSquashMerge          bool   `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit          bool   `json:"allow_merge_commit,omitempty"`
	AllowRebaseMerge          bool   `json:"allow_rebase_merge,omitempty"`
	AllowAutoMerge            bool   `json:"allow_auto_merge,omitempty"`
	DeleteBranchOnMerge       bool   `json:"delete_branch_on_merge,omitempty"`
	AllowUpdateBranch         bool   `json:"allow_update_branch,omitempty"`
	UseSquashPrTitleAsDefault bool   `json:"use_squash_pr_title_as_default,omitempty"`
	SquashMergeCommitTitle    string `json:"squash_merge_commit_title,omitempty"`
	SquashMergeCommitMessage  string `json:"squash_merge_commit_message,omitempty"`
	MergeCommitTitle          string `json:"merge_commit_title,omitempty"`
	MergeCommitMessage        string `json:"merge_commit_message,omitempty"`
	Archived                  bool   `json:"archived,omitempty"`
	AllowForking              bool   `json:"allow_forking,omitempty"`
}

// Update modifies an existing repository's properties.
// This method allows you to update various settings and metadata of a repository
// such as its name, description, visibility, merge settings, and other configuration
// options. Only the provided fields will be updated.
func (s *RepositoriesService) Update(ctx context.Context, owner string, repo string, body RepositoryUpdateRequest) (*Repository, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, nil, err
	}

	r := new(Repository)
	resp, err := s.client.Do(ctx, req, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// Delete removes a repository permanently.
// This method deletes the specified repository. This action cannot be undone,
// and all data including issues, pull requests, and wiki pages will be lost.
// Note that this requires admin permissions on the repository.
func (s *RepositoriesService) Delete(ctx context.Context, owner string, repo string) (*Response, error) {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// RepositoryCreateRequest represents the request body for creating a repository.
// GitHub API docs: https://docs.github.com/en/rest/repos/repos#create-a-repository-for-the-authenticated-user
type RepositoryCreateRequest struct {
	Name                     string `json:"name"`
	Description              string `json:"description,omitempty"`
	Homepage                 string `json:"homepage,omitempty"`
	Private                  bool   `json:"private,omitempty"`
	HasIssues                bool   `json:"has_issues,omitempty"`
	HasProjects              bool   `json:"has_projects,omitempty"`
	HasWiki                  bool   `json:"has_wiki,omitempty"`
	HasDiscussions           bool   `json:"has_discussions,omitempty"`
	TeamID                   int    `json:"team_id,omitempty"`
	AutoInit                 bool   `json:"auto_init,omitempty"`
	GitignoreTemplate        string `json:"gitignore_template,omitempty"`
	LicenseTemplate          string `json:"license_template,omitempty"`
	AllowSquashMerge         bool   `json:"allow_squash_merge,omitempty"`
	AllowMergeCommit         bool   `json:"allow_merge_commit,omitempty"`
	AllowRebaseMerge         bool   `json:"allow_rebase_merge,omitempty"`
	AllowAutoMerge           bool   `json:"allow_auto_merge,omitempty"`
	DeleteBranchOnMerge      bool   `json:"delete_branch_on_merge,omitempty"`
	SquashMergeCommitTitle   string `json:"squash_merge_commit_title,omitempty"`
	SquashMergeCommitMessage string `json:"squash_merge_commit_message,omitempty"`
	MergeCommitTitle         string `json:"merge_commit_title,omitempty"`
	MergeCommitMessage       string `json:"merge_commit_message,omitempty"`
	HasDownloads             bool   `json:"has_downloads,omitempty"`
	IsTemplate               bool   `json:"is_template,omitempty"`
}

// Create creates a new repository for the authenticated user.
// This method allows you to create a new repository with various initial
// configuration options such as description, visibility, initialization
// settings, and merge preferences. The repository will be owned by
// the authenticated user.
func (s *RepositoriesService) Create(ctx context.Context, body RepositoryCreateRequest) (*Repository, *Response, error) {
	path := "user/repos"

	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, nil, err
	}

	repo := new(Repository)
	resp, err := s.client.Do(ctx, req, repo)
	if err != nil {
		return nil, resp, err
	}

	return repo, resp, nil
}

// RepositoryListOptions specifies the optional parameters to list repositories.
// GitHub API docs: https://docs.github.com/en/rest/repos/repos#list-repositories-for-a-user
type RepositoryListOptions struct {
	*ListOptions
	Type      *string
	Sort      *string
	Direction *string
	Anon      *string
}

// List retrieves repositories for a specific user.
// This method allows you to list repositories owned by a particular user
// with various filtering and sorting options. You can filter by repository
// type (all, owner, member) and sort by creation date, update date, or
// other criteria. The results are returned in pages.
func (s *RepositoriesService) List(ctx context.Context, owner string, opts *RepositoryListOptions) ([]*Repository, *Response, error) {
	path := fmt.Sprintf("users/%s/repos", owner)

	if opts != nil {
		v := url.Values{}

		if opts.ListOptions != nil {
			opts.Apply(v)
		}
		if opts.Type != nil {
			v.Set("type", *opts.Type)
		}
		if opts.Sort != nil {
			v.Set("sort", *opts.Sort)
		}
		if opts.Direction != nil {
			v.Set("direction", *opts.Direction)
		}

		if len(v) != 0 {
			path += "?" + v.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	repos := new([]*Repository)
	res, err := s.client.Do(ctx, req, repos)
	if err != nil {
		return nil, res, err
	}

	return *repos, res, nil
}

// ListContributors retrieves the list of contributors for a repository.
// This method returns a list of users who have contributed to the specified
// repository. You can include anonymous contributors in the results and
// the list is sorted by the number of contributions. The results are
// returned in pages according to the pagination options.
func (s *RepositoriesService) ListContributors(ctx context.Context, owner string, repo string, opts *RepositoryListOptions) ([]*User, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/contributors", owner, repo)

	if opts != nil {
		v := url.Values{}

		if opts.ListOptions != nil {
			opts.Apply(v)
		}
		if opts.Anon != nil {
			v.Set("anon", *opts.Anon)
		}

		if len(v) != 0 {
			path += "?" + v.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	contributors := new([]*User)
	res, err := s.client.Do(ctx, req, contributors)
	if err != nil {
		return nil, res, err
	}

	return *contributors, res, nil
}
