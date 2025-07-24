package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type RepositoriesService struct {
	client *Client
}

type Repository struct {
	Id              int64      `json:"id"`
	Name            string     `json:"name"`
	Fullname        string     `json:"full_name"`
	Owner           *User      `json:"owner"`
	Private         bool       `json:"private"`
	HtmlUrl         string     `json:"html_url"`
	Description     string     `json:"description"`
	
	Fork            bool       `json:"fork"`
	Url             string     `json:"url"`
	CloneUrl        string     `json:"clone_url"`
	MirrorUrl       string     `json:"mirror_url"`
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

func (s *RepositoriesService) Get(ctx context.Context, owner string, repo string) (*Repository, error) {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	r := new(Repository)
	if _, err = s.client.Do(ctx, req, r); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return r, nil
}

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

func (s *RepositoriesService) Update(ctx context.Context, owner string, repo string, body RepositoryUpdateRequest) (*Repository, error) {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	r := new(Repository)
	if _, err = s.client.Do(ctx, req, r); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return r, nil
}

func (s *RepositoriesService) Delete(ctx context.Context, owner string, repo string) error {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("repsponse parsing error: %w", err)
	}

	return nil
}

type RepositoryCreateRequest struct {
	Name                     string `json:"name"`
	Description              string `json:"description,omitempty"`
	Homepage                 string `json:"homepage,omitempty"`
	Private                  bool   `json:"private,omitempty"`
	HasIssues                bool   `json:"has_issues,omitempty"`
	HasProjects              bool   `json:"has_projects,omitempty"`
	HasWiki                  bool   `json:"has_wiki,omitempty"`
	HasDiscussions           bool   `json:"has_discussions,omitempty"`
	TeamId                   int    `json:"team_id,omitempty"`
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

func (s *RepositoriesService) Create(ctx context.Context, body RepositoryCreateRequest) (*Repository, error) {
	path := "user/repos"
	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	repo := new(Repository)
	if _, err = s.client.Do(ctx, req, repo); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return repo, nil
}

type RepositoryListOptions struct {
	*ListOptions
	Type      *string
	Sort      *string
	Direction *string
	Anon      *string
}

func (s *RepositoriesService) List(ctx context.Context, owner string, opts *RepositoryListOptions) ([]*Repository, *Response, error) {
	path := fmt.Sprintf("users/%s/repos", owner)

	if opts != nil {
		q := url.Values{}

		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Type != nil {
			q.Set("type", *opts.Type)
		}
		if opts.Sort != nil {
			q.Set("sort", *opts.Sort)
		}
		if opts.Direction != nil {
			q.Set("direction", *opts.Direction)
		}

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	repos := new([]*Repository)
	res, err := s.client.Do(ctx, req, repos)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *repos, res, nil
}

func (s *RepositoriesService) ListContributors(ctx context.Context, owner string, repo string, opts *RepositoryListOptions) ([]*User, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/contributors", owner, repo)

	if opts != nil {
		q := url.Values{}

		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Anon != nil {
			q.Set("anon", *opts.Anon)
		}

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	contributors := new([]*User)
	res, err := s.client.Do(ctx, req, contributors)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *contributors, res, nil
}
