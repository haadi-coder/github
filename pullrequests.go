package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type PullRequestsService struct {
	client *Client
}

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

func (s *PullRequestsService) Get(ctx context.Context, owner string, repo string, pull int) (*PullRequest, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, pull)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	pr := new(PullRequest)
	if _, err = s.client.Do(ctx, req, pr); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return pr, nil
}

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

func (s *PullRequestsService) Create(ctx context.Context, owner string, repo string, body *PullRequestCreateRequest) (*PullRequest, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls", owner, repo)
	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	pr := new(PullRequest)
	if _, err = s.client.Do(ctx, req, pr); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return pr, nil
}

type PullRequestUpdateRequest struct {
	Title              string `json:"title,omitempty"`
	Base               string `json:"base,omitempty"`
	Body               string `json:"body,omitempty"`
	State              string `json:"state,omitempty"`
	MantainerCanModify bool   `json:"maintainer_can_modify,omitempty"`
}

func (s *PullRequestsService) Update(ctx context.Context, owner string, repo string, pull int, body *PullRequestUpdateRequest) (*PullRequest, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, pull)
	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	pr := new(PullRequest)
	if _, err = s.client.Do(ctx, req, pr); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return pr, nil
}

type MergeRequest struct {
	CommitTitle   string `json:"commit_title,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	Sha           string `json:"sha,omitempty"`
	MergeMethod   string `json:"merge_method,omitempty"`
}

type Merge struct {
	Sha     string `json:"sha"`
	Merged  bool   `json:"merged"`
	Message string `json:"message"`
}

func (s *PullRequestsService) Merge(ctx context.Context, owner string, repo string, pull int, body *MergeRequest) (*Merge, error) {
	path := fmt.Sprintf("repos/%s/%s/pulls/%d/merge", owner, repo, pull)
	req, err := s.client.NewRequest(http.MethodPut, path, body)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	merge := new(Merge)
	if _, err = s.client.Do(ctx, req, merge); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return merge, nil
}

type PullRequestListOptions struct {
	*ListOptions
	State     *string
	Head      *string
	Base      *string
	Sort      *string
	Direction *string
}

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
		return nil, nil, fmt.Errorf("request creating error: %w", err)
	}

	prs := new([]*PullRequest)
	res, err := s.client.Do(ctx, req, prs)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *prs, res, nil
}
