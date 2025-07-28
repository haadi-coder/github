package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// IssuesService provides access to issue-related API methods.
type IssuesService struct {
	client *Client
}

// Label represents a GitHub label.
// GitHub API docs: https://docs.github.com/en/rest/issues/labels
type Label struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
}

// Issue represents a GitHub issue.
// GitHub API docs: https://docs.github.com/en/rest/issues/issues
type Issue struct {
	ID            int64      `json:"id"`
	URL           string     `json:"url"`
	RepositoryURL string     `json:"repository_url"`
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

// Get fetches an issue by its number in a repository.
// This method retrieves detailed information about a specific issue,
// including its title, body, labels, assignees, and other metadata.
// The issue number is the unique identifier within the repository.
func (s *IssuesService) Get(ctx context.Context, owner string, repo string, issueNum int) (*Issue, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, issueNum)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

// IssueCreateRequest represents the request body for creating an issue.
// GitHub API docs: https://docs.github.com/en/rest/issues/issues#create-an-issue
type IssueCreateRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Milestone string   `json:"milestone,omitempty"`
	Labels    []*Label `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Type      string   `json:"type,omitempty"`
}

// Create creates a new issue in a repository.
// This method allows you to create a new issue with specified title, body,
// assignees, labels, and other optional parameters. The created issue
// will be owned by the specified repository owner and repository name.
func (s *IssuesService) Create(ctx context.Context, owner string, repo string, body *IssueCreateRequest) (*Issue, error) {
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)

	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

// IssueUpdateRequest represents the request body for updating an issue.
// GitHub API docs: https://docs.github.com/en/rest/issues/issues#update-an-issue
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

// Update updates an existing issue in a repository.
// This method allows you to modify an existing issue by its number.
// You can update the title, body, assignees, labels, state, and other
// properties of the issue. Only provided fields will be updated.
func (s *IssuesService) Update(ctx context.Context, owner string, repo string, issueNum int, body *IssueUpdateRequest) (*Issue, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, issueNum)

	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	issue := new(Issue)
	if _, err = s.client.Do(ctx, req, issue); err != nil {
		return nil, err
	}

	return issue, nil
}

// IssueLockRequest represents the request body for locking an issue.
// GitHub API docs: https://docs.github.com/en/rest/issues/issues#lock-an-issue
type IssueLockRequest struct {
	LockReason string `json:"lock_reason"`
}

// Lock locks an issue, limiting comments to collaborators only.
// This method prevents non-collaborators from commenting on the issue.
// You can optionally specify a lock reason such as "off-topic", "too heated",
// "resolved", or "spam" to provide context for why the issue was locked.
func (s *IssuesService) Lock(ctx context.Context, owner string, repo string, issueNum int, body *IssueLockRequest) error {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/lock", owner, repo, issueNum)

	req, err := s.client.NewRequest(http.MethodPut, path, body)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

// Unlock unlocks a previously locked issue.
// This method removes the lock from an issue, allowing all users
// (including non-collaborators) to comment on it again.
func (s *IssuesService) Unlock(ctx context.Context, owner string, repo string, issueNum int) error {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/lock", owner, repo, issueNum)
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

// IssueListOptions specifies the optional parameters to various List methods that support pagination.
// GitHub API docs: https://docs.github.com/en/rest/issues/issues#list-repository-issues
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

// ListByRepo lists issues in a repository.
// This method retrieves a list of issues for the specified repository.
// You can filter and sort the results using various options such as
// issue state, assignee, creator, labels, and creation date.
// The results are returned in pages according to the pagination options.
func (s *IssuesService) ListByRepo(ctx context.Context, owner string, repo string, opts *IssueListOptions) ([]*Issue, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/issues", owner, repo)

	if opts != nil {
		v := url.Values{}

		if opts.ListOptions != nil {
			opts.Apply(v)
		}
		if opts.Assignee != nil {
			v.Set("assignee", *opts.Assignee)
		}
		if opts.Creator != nil {
			v.Set("creator", *opts.Creator)
		}
		if opts.Mentioned != nil {
			v.Set("mentioned", *opts.Mentioned)
		}
		if opts.State != nil {
			v.Set("state", *opts.State)
		}
		if opts.Type != nil {
			v.Set("type", *opts.Type)
		}
		if len(opts.Labels) != 0 {
			v.Set("labels", strings.Join(opts.Labels, ","))
		}
		if opts.Since != nil {
			t, _ := opts.Since.MarshalJSON()
			v.Set("since", string(t))
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

	issues := new([]*Issue)
	res, err := s.client.Do(ctx, req, issues)
	if err != nil {
		return nil, res, err
	}

	return *issues, res, nil
}

// IssueCommentRequest represents the request body for creating or updating an issue comment.
// GitHub API docs: https://docs.github.com/en/rest/issues/comments
type IssueCommentRequest struct {
	Body string `json:"body"`
}

// IssueComment represents a comment on an issue.
// GitHub API docs: https://docs.github.com/en/rest/issues/comments
type IssueComment struct {
	ID        int        `json:"id"`
	URL       string     `json:"url"`
	Body      string     `json:"body"`
	User      *User      `json:"user"`
	CreatedAt *Timestamp `json:"created_at"`
	UpdatedAt *Timestamp `json:"updated_at"`
	IssueURL  string     `json:"issue_url"`
}

// CreateComment creates a comment on an issue.
// This method adds a new comment to the specified issue. The comment
// will be authored by the authenticated user and will appear in the
// issue's comment thread.
func (s *IssuesService) CreateComment(ctx context.Context, owner string, repo string, issueNum int, body IssueCommentRequest) (*IssueComment, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, issueNum)

	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	comment := new(IssueComment)
	if _, err = s.client.Do(ctx, req, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// IssueCommentListOptions specifies the optional parameters to list issue comments.
// GitHub API docs: https://docs.github.com/en/rest/issues/comments#list-issue-comments
type IssueCommentListOptions struct {
	*ListOptions
	Since     *Timestamp
	Sort      *string
	Direction *string
}

// ListCommentsByRepo lists comments in a repository.
// This method retrieves all comments across all issues in the specified
// repository. You can filter the results by creation date and sort them
// according to your preferences. The results are returned in pages.
func (s *IssuesService) ListCommentsByRepo(ctx context.Context, owner string, repo string, opts *IssueCommentListOptions) ([]*IssueComment, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/issues/comments", owner, repo)

	if opts != nil {
		v := url.Values{}

		if opts.ListOptions != nil {
			opts.Apply(v)
		}
		if opts.Since != nil {
			t, _ := opts.Since.MarshalJSON()
			v.Set("since", string(t))
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

	comments := new([]*IssueComment)
	res, err := s.client.Do(ctx, req, comments)
	if err != nil {
		return nil, res, err
	}

	return *comments, res, nil
}
