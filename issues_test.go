package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuesService_Get(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		issueNum       int
		expectedUrl    string
		responseBody   string
		expectedResult *Issue
	}{
		{
			name:        "Issue get",
			owner:       "octocat",
			repoName:    "Hello-World",
			issueNum:    1,
			expectedUrl: "/repos/octocat/Hello-World/issues/1",
			responseBody: `{
                "id": 1,
                "url": "https://api.github.com/repos/octocat/Hello-World/issues/1",
                "repository_url": "https://api.github.com/repos/octocat/Hello-World",
                "number": 1,
                "state": "open",
                "title": "Test Issue",
                "body": "Test Body",
                "labels": [{"id": 1, "name": "bug"}],
                "user": {"id": 583231, "login": "octocat"},
                "assignee": {"id": 583231, "login": "octocat"},
                "assignees": [{"id": 583231, "login": "octocat"}],
                "locked": false,
                "comments": 1,
                "created_at": "2023-10-10T12:00:00Z",
                "updated_at": "2023-10-11T14:30:00Z"
            }`,
			expectedResult: &Issue{
				Id:            1,
				Url:           "https://api.github.com/repos/octocat/Hello-World/issues/1",
				RepositoryUrl: "https://api.github.com/repos/octocat/Hello-World",
				Number:        1,
				State:         "open",
				Title:         "Test Issue",
				Body:          "Test Body",
				Labels:        []*Label{{Id: 1, Name: "bug"}},
				User:          &User{Id: 583231, Login: "octocat"},
				Assignee:      &User{Id: 583231, Login: "octocat"},
				Assignees:     []*User{{Id: 583231, Login: "octocat"}},
				Locked:        false,
				Comments:      1,
				CreatedAt:     &Timestamp{time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)},
				UpdatedAt:     &Timestamp{time.Date(2023, 10, 11, 14, 30, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL), WithRequestHook(func(r *http.Request) {
				fmt.Print(r.URL)
			}))

			issue, err := client.Issues.Get(context.Background(), tt.owner, tt.repoName, tt.issueNum)
			require.NoError(t, err)
			require.NotNil(t, issue)

			assert.Equal(t, tt.expectedResult.Id, issue.Id)
			assert.Equal(t, tt.expectedResult.Title, issue.Title)
			assert.Equal(t, tt.expectedResult.User.Login, issue.User.Login)
			assert.Equal(t, tt.expectedResult.Assignee.Login, issue.Assignee.Login)
			assert.Equal(t, tt.expectedResult.Assignees[0].Login, issue.Assignees[0].Login)
			assert.Equal(t, tt.expectedResult.Locked, issue.Locked)
			assert.Equal(t, tt.expectedResult.Comments, issue.Comments)

			if tt.expectedResult.CreatedAt != nil {
				assert.Equal(t, tt.expectedResult.CreatedAt.Format(time.RFC3339), issue.CreatedAt.Format(time.RFC3339))
			}
			if tt.expectedResult.UpdatedAt != nil {
				assert.Equal(t, tt.expectedResult.UpdatedAt.Format(time.RFC3339), issue.UpdatedAt.Format(time.RFC3339))
			}
		})
	}
}

func TestIssuesService_Create(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		body           *IssueCreateRequest
		expectedUrl    string
		responseBody   string
		expectedResult *Issue
	}{
		{
			name:     "Create Issue",
			owner:    "octocat",
			repoName: "Hello-World",
			body: &IssueCreateRequest{
				Title:  "New Issue",
				Body:   "New Body",
				Labels: []*Label{{Name: "bug"}},
			},
			expectedUrl: "/repos/octocat/Hello-World/issues",
			responseBody: `{
                "id": 1,
                "title": "New Issue",
                "body": "New Body",
                "labels": [{"id": 1, "name": "bug"}],
                "user": {"id": 583231, "login": "octocat"},
                "created_at": "2023-10-10T12:00:00Z"
            }`,
			expectedResult: &Issue{
				Id:        1,
				Title:     "New Issue",
				Body:      "New Body",
				Labels:    []*Label{{Id: 1, Name: "bug"}},
				User:      &User{Id: 583231, Login: "octocat"},
				CreatedAt: &Timestamp{time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, _ := io.ReadAll(r.Body)
				var reqBody IssueCreateRequest
				_ = json.Unmarshal(body, &reqBody)
				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			issue, err := client.Issues.Create(context.Background(), tt.owner, tt.repoName, tt.body)
			require.NoError(t, err)
			require.NotNil(t, issue)

			assert.Equal(t, tt.expectedResult.Id, issue.Id)
			assert.Equal(t, tt.expectedResult.Title, issue.Title)
			assert.Equal(t, tt.expectedResult.Body, issue.Body)
			assert.Equal(t, tt.expectedResult.Labels[0].Name, issue.Labels[0].Name)
			assert.Equal(t, tt.expectedResult.User.Login, issue.User.Login)
		})
	}
}

func TestIssuesService_Edit(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		issueNum       int
		body           *IssueUpdateRequest
		expectedUrl    string
		responseBody   string
		expectedResult *Issue
	}{
		{
			name:     "Update Issue",
			owner:    "octocat",
			repoName: "Hello-World",
			issueNum: 1,
			body: &IssueUpdateRequest{
				Title:  "Updated Title",
				State:  "closed",
				Labels: []*Label{{Name: "enhancement"}},
			},
			expectedUrl: "/repos/octocat/Hello-World/issues/1",
			responseBody: `{
                "id": 1,
                "title": "Updated Title",
                "state": "closed",
                "labels": [{"id": 2, "name": "enhancement"}],
                "updated_at": "2023-10-12T15:00:00Z"
            }`,
			expectedResult: &Issue{
				Id:        1,
				Title:     "Updated Title",
				State:     "closed",
				Labels:    []*Label{{Id: 2, Name: "enhancement"}},
				UpdatedAt: &Timestamp{time.Date(2023, 10, 12, 15, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "PATCH", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, _ := io.ReadAll(r.Body)
				var reqBody IssueUpdateRequest
				_ = json.Unmarshal(body, &reqBody)
				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			issue, err := client.Issues.Edit(context.Background(), tt.owner, tt.repoName, tt.issueNum, tt.body)
			require.NoError(t, err)
			require.NotNil(t, issue)

			assert.Equal(t, tt.expectedResult.Id, issue.Id)
			assert.Equal(t, tt.expectedResult.Title, issue.Title)
			assert.Equal(t, tt.expectedResult.State, issue.State)
			assert.Equal(t, tt.expectedResult.Labels[0].Name, issue.Labels[0].Name)
		})
	}
}

func TestIssuesService_LockUnlock(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		issueNum       int
		body           *IssueLockRequest
		isLock         bool
		expectedUrl    string
		responseStatus int
	}{
		{
			name:           "Lock issue",
			owner:          "octocat",
			repoName:       "Hello-World",
			issueNum:       1,
			body:           &IssueLockRequest{LockReason: "spam"},
			isLock:         true,
			expectedUrl:    "/repos/octocat/Hello-World/issues/1/lock",
			responseStatus: http.StatusNoContent,
		},
		{
			name:           "Unlock issue",
			owner:          "octocat",
			repoName:       "Hello-World",
			issueNum:       1,
			body:           nil,
			isLock:         false,
			expectedUrl:    "/repos/octocat/Hello-World/issues/1/lock",
			responseStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				if tt.isLock {
					assert.Equal(t, "PUT", r.Method)
				} else {
					assert.Equal(t, "DELETE", r.Method)
				}
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				if tt.isLock {
					body, _ := io.ReadAll(r.Body)
					var reqBody IssueLockRequest
					_ = json.Unmarshal(body, &reqBody)
					assert.Equal(t, *tt.body, reqBody)
				}

				w.WriteHeader(tt.responseStatus)
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

			var err error
			if tt.isLock {
				err = client.Issues.Lock(context.Background(), tt.owner, tt.repoName, tt.issueNum, tt.body)
			} else {
				err = client.Issues.Unlock(context.Background(), tt.owner, tt.repoName, tt.issueNum)
			}
			require.NoError(t, err)
		})
	}
}

func TestIssuesService_ListByRepo(t *testing.T) {
	state := "open"
	assignee := "octocat"
	tests := []struct {
		name           string
		owner          string
		repoName       string
		opts           *IssueListOptions
		expectedUrl    string
		responseBody   string
		expectedResult []*Issue
	}{
		{
			name:     "Issue list with filtering",
			owner:    "octocat",
			repoName: "Hello-World",
			opts: &IssueListOptions{
				ListOptions: &ListOptions{Page: 1, PerPage: 30},
				State:       &state,
				Assignee:    &assignee,
				Labels:      []string{"enhancement"},
			},
			expectedUrl:  "/repos/octocat/Hello-World/issues?assignee=octocat&labels=enhancement&page=1&per_page=30&state=open",
			responseBody: `[{"id":1,"title":"Issue 1"},{"id":2,"title":"Issue 2"}]`,
			expectedResult: []*Issue{
				{Id: 1, Title: "Issue 1"},
				{Id: 2, Title: "Issue 2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			issues, _, err := client.Issues.ListByRepo(context.Background(), tt.owner, tt.repoName, tt.opts)
			require.NoError(t, err)
			require.NotNil(t, issues)
			assert.Len(t, issues, len(tt.expectedResult))
			for i := range issues {
				assert.Equal(t, tt.expectedResult[i].Id, issues[i].Id)
				assert.Equal(t, tt.expectedResult[i].Title, issues[i].Title)
			}
		})
	}
}

func TestIssuesService_CreateComment(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		issueNum       int
		body           IssueCommentRequest
		expectedUrl    string
		responseBody   string
		expectedResult *IssueComment
	}{
		{
			name:        "Comment creating",
			owner:       "octocat",
			repoName:    "Hello-World",
			issueNum:    1,
			body:        IssueCommentRequest{Body: "New comment"},
			expectedUrl: "/repos/octocat/Hello-World/issues/1/comments",
			responseBody: `{
                "id": 1,
                "body": "New comment",
                "user": {"id": 583231, "login": "octocat"},
                "created_at": "2023-10-10T12:00:00Z"
            }`,
			expectedResult: &IssueComment{
				Id:        1,
				Body:      "New comment",
				User:      &User{Id: 583231, Login: "octocat"},
				CreatedAt: &Timestamp{time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, _ := io.ReadAll(r.Body)
				var reqBody IssueCommentRequest
				_ = json.Unmarshal(body, &reqBody)
				assert.Equal(t, tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			comment, err := client.Issues.CreateComment(context.Background(), tt.owner, tt.repoName, tt.issueNum, tt.body)
			require.NoError(t, err)
			require.NotNil(t, comment)

			assert.Equal(t, tt.expectedResult.Id, comment.Id)
			assert.Equal(t, tt.expectedResult.Body, comment.Body)
			assert.Equal(t, tt.expectedResult.User.Login, comment.User.Login)
		})
	}
}

func TestIssuesService_ListCommentsByRepo(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		opts           *IssueCommentListOptions
		expectedUrl    string
		responseBody   string
		expectedResult []*IssueComment
	}{
		{
			name:     "Comments list",
			owner:    "octocat",
			repoName: "Hello-World",
			opts: &IssueCommentListOptions{
				ListOptions: &ListOptions{Page: 1, PerPage: 30},
				Since:       &Timestamp{time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)},
			},
			expectedUrl:  "/repos/octocat/Hello-World/issues/comments?page=1&per_page=30&since=2023-10-10T12:00:00Z",
			responseBody: `[{"id":1,"body":"Comment 1"},{"id":2,"body":"Comment 2"}]`,
			expectedResult: []*IssueComment{
				{Id: 1, Body: "Comment 1"},
				{Id: 2, Body: "Comment 2"},
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				url, _ := url.QueryUnescape(r.URL.String())
				assert.Equal(t, tt.expectedUrl, url)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			comments, _, err := client.Issues.ListCommentsByRepo(context.Background(), tt.owner, tt.repoName, tt.opts)
			require.NoError(t, err)
			require.NotNil(t, comments)
			assert.Len(t, comments, len(tt.expectedResult))
			for i := range comments {
				assert.Equal(t, tt.expectedResult[i].Id, comments[i].Id)
				assert.Equal(t, tt.expectedResult[i].Body, comments[i].Body)
			}
		})
	}
}
