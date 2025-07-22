package main

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type IssuesService struct {
	client *Client
}

type Issue struct {
	Id            int64      `json:"id"`
	Url           string     `json:"url"`
	RepositoryUrl string     `json:"repository_url"`
	Number        int        `json:"number"`
	State         string     `json:"state"`
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	User          *User      `json:"user"`
	Assignee      *User      `json:"assignee"`
	Assignees     []*User    `json:"assignees"`
	Locked        bool       `json:"locked"`
	Comments      int        `json:"comments"`
	ClosedAt      *Timestamp `json:"closed_at"`
	CreatedAt     *Timestamp `json:"created_at"`
	UpdatedAt     *Timestamp `json:"updated_at"`
	ClosedBy      *User      `json:"closed_by"`
}

type IssueRequest struct {
	Title       string   `json:"title"`
	Body        string   `json:"body,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	State       string   `json:"state"`
	StateReason string   `json:"state_reason"`
	Labels      []string `json:"labels,omitempty"`
	Assignees   []string `json:"assignees,omitempty"`
}

type IssueListOptions struct {
	*ListOptions
	State     string
	Assignee  string
	Type      string
	Creator   string
	Mentioned string
	Labels    []string
	Since     *Timestamp
	Sort      string
	Direction string
}

func (s *IssuesService) Get(ctx context.Context, owner string, repoName string, issueNum int) (*Issue, error) {
	path, _ := url.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum))
	return fetchIssue[Issue](ctx, s.client, http.MethodGet, path, nil)
}

func (s *IssuesService) Create(ctx context.Context, owner string, repoName string, body *IssueRequest) (*Issue, error) {
	path, _ := url.JoinPath("repos", owner, repoName, "issues")
	return fetchIssue[Issue](ctx, s.client, http.MethodPost, path, body)
}

func (s *IssuesService) Edit(ctx context.Context, owner string, repoName string, issueNum int, body *IssueRequest) (*Issue, error) {
	path, _ := url.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum))
	return fetchIssue[Issue](ctx, s.client, http.MethodPatch, path, body)
}

type IssueLockRequest struct {
	LockReason string `json:"lock_reason"`
}

func (s *IssuesService) Lock(ctx context.Context, owner string, repoName string, issueNum int, body *IssueLockRequest) error {
	path, _ := url.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum), "lock")
	req, err := s.client.NewRequest(http.MethodPut, path, body)
	if err != nil {
		return err
	}

	_, err = s.client.fetch(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *IssuesService) Unlock(ctx context.Context, owner string, repoName string, issueNum int) error {
	path, _ := url.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum), "lock")
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

func (s *IssuesService) ListByRepo(ctx context.Context, owner string, repoName string, opts *IssueListOptions) ([]*Issue, *Response, error) {
	path, _ := url.JoinPath("repos", owner, repoName, "issues")
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		q := req.URL.Query()
		opts.paginateQuery(q)

		if opts.Assignee != "" {
			q.Set("assignee", opts.Assignee)
		}
		if opts.Creator != "" {
			q.Set("creator", opts.Creator)
		}
		if opts.Mentioned != "" {
			q.Set("mentioned", opts.Mentioned)
		}
		if opts.State != "" {
			q.Set("state", opts.State)
		}
		if opts.Type != "" {
			q.Set("type", opts.Type)
		}
		if len(opts.Labels) != 0 {
			q.Set("labels", strings.Join(opts.Labels, ","))
		}
		if opts.Since != nil {
			t, _ := opts.Since.MarshalJSON()
			q.Set("since", string(t))
		}
		if opts.Sort != "" {
			q.Set("sort", opts.Sort)
		}
		if opts.Direction != "" {
			q.Set("direction", opts.Direction)
		}

		req.URL.RawQuery = q.Encode()
	}

	issues := new([]*Issue)
	res, err := s.client.fetch(ctx, req, issues)
	if err != nil {
		return nil, res, err
	}

	return *issues, res, nil
}

type IssueCommentRequest struct {
	Body string `json:"body"`
}

type IssueComment struct {
	Id        int        `json:"id"`
	Url       string     `json:"url"`
	Body      string     `json:"body"`
	User      *User      `json:"user"`
	CreatedAt *Timestamp `json:"created_at"`
	UpdatedAt *Timestamp `json:"updated_at"`
	IssueUrl  string     `json:"issue_url"`
}

type IssueCommentListOptions struct {
	*ListOptions
	Since     *Timestamp
	Sort      string
	Direction string
}

func (s *IssuesService) CreateComment(ctx context.Context, owner string, repoName string, issueNum int, body IssueCommentRequest) (*IssueComment, error) {
	path, _ := url.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum), "comments")
	return fetchIssue[IssueComment](ctx, s.client, http.MethodPost, path, body)
}

func (s *IssuesService) ListCommentsByRepo(ctx context.Context, owner string, repoName string, opts *IssueCommentListOptions) ([]*IssueComment, *Response, error) {
	path, _ := url.JoinPath("repos", owner, repoName, "issues/comments")
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		q := req.URL.Query()
		opts.paginateQuery(q)
		if opts.Since != nil {
			t, _ := opts.Since.MarshalJSON()
			q.Set("since", string(t))
		}
		if opts.Sort != "" {
			q.Set("sort", opts.Sort)
		}
		if opts.Direction != "" {
			q.Set("direction", opts.Direction)
		}

		req.URL.RawQuery = q.Encode()
	}

	comments := new([]*IssueComment)
	res, err := s.client.fetch(ctx, req, comments)
	if err != nil {
		return nil, res, err
	}

	return *comments, res, nil
}

func fetchIssue[Res Issue | IssueComment](ctx context.Context, client *Client, method string, path string, body any) (*Res, error) {
	req, err := client.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	result := new(Res)
	_, err = client.fetch(ctx, req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
