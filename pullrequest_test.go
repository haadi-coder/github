package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestsService_Get(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 11, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		owner        string
		repoName     string
		pullNum      int
		expectedURL  string
		responseBody string
		expected     *PullRequest
	}{
		{
			name:        "Get PullRequest",
			owner:       "octocat",
			repoName:    "Hello-World",
			pullNum:     1,
			expectedURL: "/repos/octocat/Hello-World/pulls/1",
			responseBody: fmt.Sprintf(`{
                "id": 1,
                "title": "Test PR",
                "number": 1,
                "state": "open",
                "created_at": "%s",
                "updated_at": "%s"
            }`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expected: &PullRequest{
				ID:        1,
				Title:     "Test PR",
				Number:    1,
				State:     "open",
				CreatedAt: &Timestamp{createdAt},
				UpdatedAt: &Timestamp{updatedAt},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				_, _ = w.Write([]byte(tt.responseBody))
			}))

			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			pr, resp, err := client.PullRequests.Get(context.Background(), tt.owner, tt.repoName, tt.pullNum)
			require.NoError(t, err)

			require.NotNil(t, pr)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expected, pr)
		})
	}
}

func TestPullRequestsService_Create(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 11, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		owner        string
		repoName     string
		body         *PullRequestCreateRequest
		expectedURL  string
		responseBody string
		expected     *PullRequest
	}{
		{
			name:     "Creating PR",
			owner:    "octocat",
			repoName: "Hello-World",
			body: &PullRequestCreateRequest{
				Head:  "github",
				Base:  "dev",
				Title: "New feature",
				Body:  "Adds new feature",
			},
			expectedURL: "/repos/octocat/Hello-World/pulls",
			responseBody: fmt.Sprintf(`{
                "id": 1,
                "title": "New feature",
                "number": 1,
                "state": "open",
                "created_at": "%s",
                "updated_at": "%s"
            }`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expected: &PullRequest{
				ID:        1,
				Title:     "New feature",
				Number:    1,
				State:     "open",
				CreatedAt: &Timestamp{createdAt},
				UpdatedAt: &Timestamp{updatedAt},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody PullRequestCreateRequest

				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &reqBody)

				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)

				_, _ = w.Write([]byte(tt.responseBody))
			}))

			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			pr, resp, err := client.PullRequests.Create(context.Background(), tt.owner, tt.repoName, tt.body)
			require.NoError(t, err)

			require.NotNil(t, pr)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			assert.Equal(t, tt.expected, pr)
		})
	}
}

func TestPullRequestsService_Update(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 12, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		owner        string
		repoName     string
		pullNum      int
		body         *PullRequestUpdateRequest
		expectedURL  string
		responseBody string
		expected     *PullRequest
	}{
		{
			name:     "Update PR",
			owner:    "octocat",
			repoName: "Hello-World",
			pullNum:  1,
			body: &PullRequestUpdateRequest{
				Title: "Updated title",
				State: "closed",
			},
			expectedURL: "/repos/octocat/Hello-World/pulls/1",
			responseBody: fmt.Sprintf(`{
                "id": 1,
                "title": "Updated title",
                "number": 1,
                "state": "closed",
                "created_at": "%s",
                "updated_at": "%s"
            }`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expected: &PullRequest{
				ID:        1,
				Title:     "Updated title",
				Number:    1,
				State:     "closed",
				CreatedAt: &Timestamp{createdAt},
				UpdatedAt: &Timestamp{updatedAt},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.Path)
				assert.Equal(t, "PATCH", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody PullRequestUpdateRequest

				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &reqBody)

				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				_, _ = w.Write([]byte(tt.responseBody))
			}))

			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			pr, resp, err := client.PullRequests.Update(context.Background(), tt.owner, tt.repoName, tt.pullNum, tt.body)
			require.NoError(t, err)

			require.NotNil(t, pr)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expected, pr)
		})
	}
}

func TestPullRequestsService_Merge(t *testing.T) {
	mergedAt := time.Date(2023, 10, 13, 16, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		owner        string
		repoName     string
		pullNum      int
		body         *MergeRequest
		expectedURL  string
		responseBody string
		expected     *Merge
	}{
		{
			name:     "Merge PR",
			owner:    "octocat",
			repoName: "Hello-World",
			pullNum:  1,
			body: &MergeRequest{
				Sha:         "abc123",
				MergeMethod: "merge",
			},
			expectedURL: "/repos/octocat/Hello-World/pulls/1/merge",
			responseBody: fmt.Sprintf(`{
                "sha": "abc123",
                "merged": true,
                "message": "PR merged",
                "updated_at": "%s"
            }`, mergedAt.Format(time.RFC3339)),
			expected: &Merge{
				Sha:     "abc123",
				Merged:  true,
				Message: "PR merged",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.Path)
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody MergeRequest

				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &reqBody)

				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				_, _ = w.Write([]byte(tt.responseBody))
			}))

			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			merge, resp, err := client.PullRequests.Merge(context.Background(), tt.owner, tt.repoName, tt.pullNum, tt.body)
			require.NoError(t, err)

			require.NotNil(t, merge)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expected, merge)
		})
	}
}

func TestPullRequestsService_List(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 11, 14, 30, 0, 0, time.UTC)

	state := "open"
	tests := []struct {
		name         string
		owner        string
		repoName     string
		opts         *PullRequestListOptions
		expectedURL  string
		responseBody string
		expected     []*PullRequest
	}{
		{
			name:     "List of PRs",
			owner:    "octocat",
			repoName: "Hello-World",
			opts: &PullRequestListOptions{
				ListOptions: &ListOptions{Page: 1, PerPage: 30},
				State:       &state,
			},
			expectedURL: "/repos/octocat/Hello-World/pulls?page=1&per_page=30&state=open",
			responseBody: fmt.Sprintf(`[
                {
                    "id": 1,
                    "title": "PR 1",
                    "number": 1,
                    "state": "open",
                    "created_at": "%s",
                    "updated_at": "%s"
                },
                {
                    "id": 2,
                    "title": "PR 2",
                    "number": 2,
                    "state": "closed",
                    "created_at": "%s",
                    "updated_at": "%s"
                }
            ]`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339),
				createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expected: []*PullRequest{
				{
					ID:        1,
					Title:     "PR 1",
					Number:    1,
					State:     "open",
					CreatedAt: &Timestamp{createdAt},
					UpdatedAt: &Timestamp{updatedAt},
				},
				{
					ID:        2,
					Title:     "PR 2",
					Number:    2,
					State:     "closed",
					CreatedAt: &Timestamp{createdAt},
					UpdatedAt: &Timestamp{updatedAt},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				_, _ = w.Write([]byte(tt.responseBody))
			}))

			defer ts.Close()

			client, err := NewClient(WithBaseURL(ts.URL))
			require.NoError(t, err)

			prs, _, err := client.PullRequests.List(context.Background(), tt.owner, tt.repoName, tt.opts)
			require.NoError(t, err)

			require.NotNil(t, prs)
			assert.Len(t, prs, len(tt.expected))
			assert.Equal(t, tt.expected, prs)
		})
	}
}
