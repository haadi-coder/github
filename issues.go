package github

import (
	"context"
	"net/http"
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

type IssueCreateRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone string   `json:"milestone,omitempty"`
	Labels    []*Label `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Type      string   `json:"type,omitempty"`
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

type Label struct {
	Id          int64  `json:"id"`
	Url         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
}

func (s *IssuesService) Get(ctx context.Context, owner string, repoName string, issueNum int) (*Issue, error) {
	url := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum))
	req, err := s.client.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func (s *IssuesService) Create(ctx context.Context, owner string, repoName string, body *IssueCreateRequest) (*Issue, error) {
	url := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues")
	req, err := s.client.NewRequest(http.MethodPost, url.String(), body)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

func (s *IssuesService) Edit(ctx context.Context, owner string, repoName string, issueNum int, body *IssueUpdateRequest) (*Issue, error) {
	url := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum))
	req, err := s.client.NewRequest(http.MethodPatch, url.String(), body)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

type IssueLockRequest struct {
	LockReason string `json:"lock_reason"`
}

func (s *IssuesService) Lock(ctx context.Context, owner string, repoName string, issueNum int, body *IssueLockRequest) error {
	url := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum), "lock")
	req, err := s.client.NewRequest(http.MethodPut, url.String(), body)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (s *IssuesService) Unlock(ctx context.Context, owner string, repoName string, issueNum int) error {
	url := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum), "lock")
	req, err := s.client.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (s *IssuesService) ListByRepo(ctx context.Context, owner string, repoName string, opts *IssueListOptions) ([]*Issue, *Response, error) {
	rawUrl := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues")

	if opts != nil {
		q := rawUrl.Query()

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

		rawUrl.RawQuery = q.Encode()
	}

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}

	issues := new([]*Issue)
	res, err := s.client.Do(ctx, req, issues)
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
	Sort      *string
	Direction *string
}

func (s *IssuesService) CreateComment(ctx context.Context, owner string, repoName string, issueNum int, body IssueCommentRequest) (*IssueComment, error) {
	url := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues", strconv.Itoa(issueNum), "comments")
	req, err := s.client.NewRequest(http.MethodPost, url.String(), body)
	if err != nil {
		return nil, err
	}

	comment := new(IssueComment)
	if _, err = s.client.Do(ctx, req, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *IssuesService) ListCommentsByRepo(ctx context.Context, owner string, repoName string, opts *IssueCommentListOptions) ([]*IssueComment, *Response, error) {
	rawUrl := s.client.baseUrl.JoinPath("repos", owner, repoName, "issues/comments")

	if opts != nil {
		q := rawUrl.Query()

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

		rawUrl.RawQuery = q.Encode()
	}

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}

	comments := new([]*IssueComment)
	res, err := s.client.Do(ctx, req, comments)
	if err != nil {
		return nil, res, err
	}

	return *comments, res, nil
}
