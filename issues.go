package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type IssuesService struct {
	client *Client
}

type Label struct {
	Id          int64  `json:"id"`
	Url         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
}

type Issue struct {
	Id            int64      `json:"id"`
	Url           string     `json:"url"`
	RepositoryUrl string     `json:"repository_url"`
	Number        int        `json:"number"`
	State         string     `json:"state"`
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	Labels        []*Label   `json:"labels"`
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

func (s *IssuesService) Get(ctx context.Context, owner string, repo string, issueNum int) (*Issue, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, issueNum)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return issue, nil
}

type IssueCreateRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone string   `json:"milestone,omitempty"`
	Labels    []*Label `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Type      string   `json:"type,omitempty"`
}

func (s *IssuesService) Create(ctx context.Context, owner string, repo string, body *IssueCreateRequest) (*Issue, error) {
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)
	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return issue, nil
}

type IssueUpdateRequest struct {
	Title       string   `json:"title"`
	Body        string   `json:"body,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	State       string   `json:"state"`
	StateReason string   `json:"state_reason"`
	Milestone   string   `json:"milestone,omitempty"`
	Labels      []*Label `json:"labels,omitempty"`
	Assignees   []string `json:"assignees,omitempty"`
	Type        string   `json:"type,omitempty"`
}

func (s *IssuesService) Update(ctx context.Context, owner string, repo string, issueNum int, body *IssueUpdateRequest) (*Issue, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, issueNum)
	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return issue, nil
}

type IssueLockRequest struct {
	LockReason string `json:"lock_reason"`
}

func (s *IssuesService) Lock(ctx context.Context, owner string, repo string, issueNum int, body *IssueLockRequest) error {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/lock", owner, repo, issueNum)
	req, err := s.client.NewRequest(http.MethodPut, path, body)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("repsponse parsing error: %w", err)
	}

	return nil
}

func (s *IssuesService) Unlock(ctx context.Context, owner string, repo string, issueNum int) error {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/lock", owner, repo, issueNum)
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("repsponse parsing error: %w", err)
	}

	return nil
}

type IssueListOptions struct {
	*ListOptions
	State     *string
	Assignee  *string
	Type      *string
	Creator   *string
	Mentioned *string
	Labels    []string
	Since     *Timestamp
	Sort      *string
	Direction *string
}

func (s *IssuesService) ListByRepo(ctx context.Context, owner string, repo string, opts *IssueListOptions) ([]*Issue, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)

	if opts != nil {
		q := url.Values{}

		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Assignee != nil {
			q.Set("assignee", *opts.Assignee)
		}
		if opts.Creator != nil {
			q.Set("creator", *opts.Creator)
		}
		if opts.Mentioned != nil {
			q.Set("mentioned", *opts.Mentioned)
		}
		if opts.State != nil {
			q.Set("state", *opts.State)
		}
		if opts.Type != nil {
			q.Set("type", *opts.Type)
		}
		if len(opts.Labels) != 0 {
			q.Set("labels", strings.Join(opts.Labels, ","))
		}
		if opts.Since != nil {
			t, _ := opts.Since.MarshalJSON()
			q.Set("since", string(t))
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

	issues := new([]*Issue)
	res, err := s.client.Do(ctx, req, issues)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
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

func (s *IssuesService) CreateComment(ctx context.Context, owner string, repo string, issueNum int, body IssueCommentRequest) (*IssueComment, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, issueNum)
	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	comment := new(IssueComment)
	if _, err = s.client.Do(ctx, req, comment); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return comment, nil
}

type IssueCommentListOptions struct {
	*ListOptions
	Since     *Timestamp
	Sort      *string
	Direction *string
}

func (s *IssuesService) ListCommentsByRepo(ctx context.Context, owner string, repo string, opts *IssueCommentListOptions) ([]*IssueComment, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/comments", owner, repo)

	if opts != nil {
		q := url.Values{}

		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Since != nil {
			t, _ := opts.Since.MarshalJSON()
			q.Set("since", string(t))
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

	comments := new([]*IssueComment)
	res, err := s.client.Do(ctx, req, comments)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *comments, res, nil
}
