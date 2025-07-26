package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PullRequestsService provides access to pull request-related API methods.
type PullRequestsService struct {
	client *Client
}

// PullRequest represents a GitHub pull request.
// GitHub API docs: https://docs.github.com/en/rest/pulls/pulls
type PullRequest struct {
	Id                 int         `json:"id"`
	Title              string      `json:"title"`
	Body               string      `json:"body"`
	Url                string      `json:"url"`
	Number             int         `json:"number"`
	State              string      `json:"state"`
	Locked             bool        `json:"locked"`
	ActiveLockReason   string      `json:"active_lock_reason"`
	Labels             []*Label    `json:"labels"`
	CreatedAt          *Timestamp  `json:"created_at"`
	UpdatedAt          *Timestamp  `json:"updated_at"`
	ClosedAt           *Timestamp  `json:"closed_at"`
	Assignee           *User       `json:"assignee"`
	Assignees          []*User     `json:"assignees"`
	RequestedReviewers []*User     `json:"requested_reviewers"`
	Repository         *Repository `json:"repository"`
	User               *User       `json:"user"`
	HtmlUrl            string      `json:"html_url"`
	DiffUrl            string      `json:"diff_url"`
	PatchUrl           string      `json:"patch_url"`
	IssueUrl           string      `json:"issue_url"`
	CommitsUrl         string      `json:"commits_url"`
	CommentsUrl        string      `json:"comments_url"`
	StatusesUrl        string      `json:"statuses_url"`
}

// Get fetches a pull request by its number in a repository.
// This method retrieves detailed information about a specific pull request,
// including its title, description, status, participants, and related metadata.
// The pull request number is the unique identifier within the repository.
func (s *PullRequestsService) Get(ctx context.Context, owner string, repo string, pull int) (*PullRequest, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, pull)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	pr := new(PullRequest)
	if _, err = s.client.Do(ctx, req, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

// PullRequestCreateRequest represents the request body for creating a pull request.
// GitHub API docs: https://docs.github.com/en/rest/pulls/pulls#create-a-pull-request
type PullRequestCreateRequest struct {
	Head               string `json:"head"`
	Base               string `json:"base"`
	Title              string `json:"title,omitempty"`
	HeadRepo           string `json:"head_repo,omitempty"`
	Body               string `json:"body,omitempty"`
	MantainerCanModify bool   `json:"maintainer_can_modify,omitempty"`
	Draft              bool   `json:"draft,omitempty"`
	Issue              int    `json:"issue,omitempty"`
}

// Create creates a new pull request in a repository.
// This method allows you to create a new pull request by specifying the source
// branch (head) and target branch (base). You can also provide a title, description,
// and other optional parameters. The created pull request will be owned by the
// specified repository owner and repository name.
func (s *PullRequestsService) Create(ctx context.Context, owner string, repo string, body *PullRequestCreateRequest) (*PullRequest, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls", owner, repo)

	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	pr := new(PullRequest)
	if _, err = s.client.Do(ctx, req, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

// PullRequestUpdateRequest represents the request body for updating a pull request.
// GitHub API docs: https://docs.github.com/en/rest/pulls/pulls#update-a-pull-request
type PullRequestUpdateRequest struct {
	Title              string `json:"title,omitempty"`
	Base               string `json:"base,omitempty"`
	Body               string `json:"body,omitempty"`
	State              string `json:"state,omitempty"`
	MantainerCanModify bool   `json:"maintainer_can_modify,omitempty"`
}

// Update updates an existing pull request in a repository.
// This method allows you to modify an existing pull request by its number.
// You can update the title, description, target branch, state, and other
// properties of the pull request. Only provided fields will be updated.
func (s *PullRequestsService) Update(ctx context.Context, owner string, repo string, pull int, body *PullRequestUpdateRequest) (*PullRequest, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, pull)

	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	pr := new(PullRequest)
	if _, err = s.client.Do(ctx, req, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

// MergeRequest represents the request body for merging a pull request.
// GitHub API docs: https://docs.github.com/en/rest/pulls/pulls#merge-a-pull-request
type MergeRequest struct {
	CommitTitle   string `json:"commit_title,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	Sha           string `json:"sha,omitempty"`
	MergeMethod   string `json:"merge_method,omitempty"`
}

// Merge represents the response from merging a pull request.
// GitHub API docs: https://docs.github.com/en/rest/pulls/pulls#merge-a-pull-request
type Merge struct {
	Sha     string `json:"sha"`
	Merged  bool   `json:"merged"`
	Message string `json:"message"`
}

// Merge merges a pull request.
// This method allows you to merge a pull request into its target branch.
// You can specify the merge method (merge, squash, or rebase), commit title,
// commit message, and the specific SHA to merge. The method returns information
// about the merge operation including the resulting commit SHA.
func (s *PullRequestsService) Merge(ctx context.Context, owner string, repo string, pull int, body *MergeRequest) (*Merge, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls/%d/merge", owner, repo, pull)

	req, err := s.client.NewRequest(http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	merge := new(Merge)
	if _, err = s.client.Do(ctx, req, merge); err != nil {
		return nil, err
	}

	return merge, nil
}

// PullRequestListOptions specifies the optional parameters to list pull requests.
// GitHub API docs: https://docs.github.com/en/rest/pulls/pulls#list-pull-requests
type PullRequestListOptions struct {
	*ListOptions
	State     *string
	Head      *string
	Base      *string
	Sort      *string
	Direction *string
}

// List retrieves a list of pull requests for a repository.
// This method allows you to list pull requests with various filtering options
// such as state (open, closed, all), source branch, target branch, and sorting.
// The results are returned in pages according to the pagination options.
func (s *PullRequestsService) List(ctx context.Context, owner string, repo string, opts *PullRequestListOptions) ([]*PullRequest, *Response, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls", owner, repo)

	if opts != nil {
		q := url.Values{}

		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Base != nil {
			q.Set("base", *opts.Base)
		}
		if opts.Direction != nil {
			q.Set("direction", *opts.Direction)
		}
		if opts.Head != nil {
			q.Set("head", *opts.Head)
		}
		if opts.Sort != nil {
			q.Set("sort", *opts.Sort)
		}
		if opts.State != nil {
			q.Set("state", *opts.State)
		}

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	prs := new([]*PullRequest)
	res, err := s.client.Do(ctx, req, prs)
	if err != nil {
		return nil, res, err
	}

	return *prs, res, nil
}
