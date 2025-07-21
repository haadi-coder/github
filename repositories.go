package main

import (
	"context"
	"net/http"
	"net/url"
)

type RepositoriesService struct {
	client *Client
}

type Repository struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Fullname string `json:"full_name"`
	Owner    *User  `json:"owner"`
}

type RepositoryRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Private             bool   `json:"private"`
	Visibility          string `json:"visibility"`
	HasIssues           bool   `json:"has_issues"`
	HasWiki             bool   `json:"has_wiki"`
	IsTemplate          bool   `json:"is_template"`
	DefaultBranch       string `json:"default_branch"`
	AllowSquashMerge    bool   `json:"allow_squash_merge"`
	AllowMergeCommit    bool   `json:"allow_merge_commit"`
	AllowRebaseMerge    bool   `json:"allow_rebase_merge"`
	AllowAutoMerge      bool   `json:"allow_auto_merge"`
	DeleteBranchOnMerge bool   `json:"delete_branch_on_merge"`
	MergeCommitTitle    string `json:"merge_commit_title"`
	MergeCommitMessage  string `json:"merge_commit_message"`
	Archived            bool   `json:"archived"`
	AllowForking        bool   `json:"allow_forking"`
}

type RepositoryListOptions struct {
	*ListOptions
	Type      string
	Sort      string
	Direction string
}

func (s *RepositoriesService) Get(ctx context.Context, owner string, repoName string) (*Repository, error) {
	path, err := url.JoinPath("repos", owner, repoName)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	repo := new(Repository)
	_, err = s.client.fetch(ctx, req, repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *RepositoriesService) Update(ctx context.Context, owner string, repoName string, body RepositoryRequest) (*Repository, error) {
	path, err := url.JoinPath("repos", owner, repoName)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	repo := new(Repository)
	_, err = s.client.fetch(ctx, req, repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *RepositoriesService) Delete(ctx context.Context, owner string, repoName string) error {
	path, err := url.JoinPath("repos", owner, repoName)
	if err != nil {
		return err
	}

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.fetch(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *RepositoriesService) List(ctx context.Context, owner string, opts *RepositoryListOptions) ([]*Repository, *Response, error) {
	path, err := url.JoinPath("users", owner, "repos")
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		q := req.URL.Query()
		opts.paginateQuery(q)

		if opts.Type != "" {
			q.Set("type", opts.Type)
		}
		if opts.Sort != "" {
			q.Set("sort", opts.Sort)
		}
		if opts.Direction != "" {
			q.Set("direction", opts.Direction)
		}

		req.URL.RawQuery = q.Encode()
	}

	repos := new([]*Repository)
	res, err := s.client.fetch(ctx, req, repos)
	if err != nil {
		return nil, res, err
	}

	return *repos, res, nil
}
