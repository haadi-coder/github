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
		name           string
		owner          string
		repoName       string
		pullNum        int
		expectedUrl    string
		responseBody   string
		expectedResult *PullRequest
	}{
		{
			name:        "Get PullRequest",
			owner:       "octocat",
			repoName:    "Hello-World",
			pullNum:     1,
			expectedUrl: "/repos/octocat/Hello-World/pulls/1",
			responseBody: fmt.Sprintf(`{
                "id": 1,
                "title": "Test PR",
                "number": 1,
                "state": "open",
                "created_at": "%s",
                "updated_at": "%s"
            }`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expectedResult: &PullRequest{
				Id:        1,
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
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			pr, err := client.PullRequests.Get(context.Background(), tt.owner, tt.repoName, tt.pullNum)
			require.NoError(t, err)
			require.NotNil(t, pr)

			assert.Equal(t, tt.expectedResult.Id, pr.Id)
			assert.Equal(t, tt.expectedResult.Title, pr.Title)

			if tt.expectedResult.CreatedAt != nil {
				require.NotNil(t, pr.CreatedAt)
				assert.Equal(t, tt.expectedResult.CreatedAt.Format(time.RFC3339), pr.CreatedAt.Format(time.RFC3339))
			} else {
				assert.Nil(t, pr.CreatedAt)
			}

			if tt.expectedResult.UpdatedAt != nil {
				require.NotNil(t, pr.UpdatedAt)
				assert.Equal(t, tt.expectedResult.UpdatedAt.Format(time.RFC3339), pr.UpdatedAt.Format(time.RFC3339))
			} else {
				assert.Nil(t, pr.UpdatedAt)
			}

		})
	}
}

func TestPullRequestsService_Create(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 11, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		owner          string
		repoName       string
		body           *PullRequestCreateRequest
		expectedUrl    string
		responseBody   string
		expectedResult *PullRequest
	}{
		{
			name:     "Creating PR",
			owner:    "octocat",
			repoName: "Hello-World",
			body: &PullRequestCreateRequest{
				Head:  "main",
				Base:  "dev",
				Title: "New feature",
				Body:  "Adds new feature",
			},
			expectedUrl: "/repos/octocat/Hello-World/pulls",
			responseBody: fmt.Sprintf(`{
                "id": 1,
                "title": "New feature",
                "number": 1,
                "state": "open",
                "created_at": "%s",
                "updated_at": "%s"
            }`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expectedResult: &PullRequest{
				Id:        1,
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
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, _ := io.ReadAll(r.Body)
				var reqBody PullRequestCreateRequest
				_ = json.Unmarshal(body, &reqBody)
				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			pr, err := client.PullRequests.Create(context.Background(), tt.owner, tt.repoName, tt.body)
			require.NoError(t, err)
			require.NotNil(t, pr)

			assert.Equal(t, tt.expectedResult.Id, pr.Id)
			assert.Equal(t, tt.expectedResult.Title, pr.Title)

			if tt.expectedResult.CreatedAt != nil {
				require.NotNil(t, pr.CreatedAt)
				assert.Equal(t, tt.expectedResult.CreatedAt.Format(time.RFC3339), pr.CreatedAt.Format(time.RFC3339))
			} else {
				assert.Nil(t, pr.CreatedAt)
			}

			if tt.expectedResult.UpdatedAt != nil {
				require.NotNil(t, pr.UpdatedAt)
				assert.Equal(t, tt.expectedResult.UpdatedAt.Format(time.RFC3339), pr.UpdatedAt.Format(time.RFC3339))
			} else {
				assert.Nil(t, pr.UpdatedAt)
			}
		})
	}
}

func TestPullRequestsService_Update(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 12, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		owner          string
		repoName       string
		pullNum        int
		body           *PullRequestUpdateRequest
		expectedUrl    string
		responseBody   string
		expectedResult *PullRequest
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
			expectedUrl: "/repos/octocat/Hello-World/pulls/1",
			responseBody: fmt.Sprintf(`{
                "id": 1,
                "title": "Updated title",
                "number": 1,
                "state": "closed",
                "created_at": "%s",
                "updated_at": "%s"
            }`, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339)),
			expectedResult: &PullRequest{
				Id:        1,
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
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "PATCH", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, _ := io.ReadAll(r.Body)
				var reqBody PullRequestUpdateRequest
				_ = json.Unmarshal(body, &reqBody)
				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			pr, err := client.PullRequests.Update(context.Background(), tt.owner, tt.repoName, tt.pullNum, tt.body)
			require.NoError(t, err)
			require.NotNil(t, pr)

			assert.Equal(t, tt.expectedResult.Id, pr.Id)
			assert.Equal(t, tt.expectedResult.Title, pr.Title)

			if tt.expectedResult.CreatedAt != nil {
				require.NotNil(t, pr.CreatedAt)
				assert.Equal(t, tt.expectedResult.CreatedAt.Format(time.RFC3339), pr.CreatedAt.Format(time.RFC3339))
			} else {
				assert.Nil(t, pr.CreatedAt)
			}

			if tt.expectedResult.UpdatedAt != nil {
				require.NotNil(t, pr.UpdatedAt)
				assert.Equal(t, tt.expectedResult.UpdatedAt.Format(time.RFC3339), pr.UpdatedAt.Format(time.RFC3339))
			} else {
				assert.Nil(t, pr.UpdatedAt)
			}
		})
	}
}

func TestPullRequestsService_Merge(t *testing.T) {
	mergedAt := time.Date(2023, 10, 13, 16, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		owner          string
		repoName       string
		pullNum        int
		body           *MergeRequest
		expectedUrl    string
		responseBody   string
		expectedResult *Merge
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
			expectedUrl: "/repos/octocat/Hello-World/pulls/1/merge",
			responseBody: fmt.Sprintf(`{
                "sha": "abc123",
                "merged": true,
                "message": "PR merged",
                "updated_at": "%s"
            }`, mergedAt.Format(time.RFC3339)),
			expectedResult: &Merge{
				Sha:     "abc123",
				Merged:  true,
				Message: "PR merged",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.Path)
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, _ := io.ReadAll(r.Body)
				var reqBody MergeRequest
				_ = json.Unmarshal(body, &reqBody)
				assert.Equal(t, *tt.body, reqBody)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			merge, err := client.PullRequests.Merge(context.Background(), tt.owner, tt.repoName, tt.pullNum, tt.body)
			require.NoError(t, err)
			require.NotNil(t, merge)

			assert.Equal(t, tt.expectedResult.Sha, merge.Sha)
			assert.Equal(t, tt.expectedResult.Merged, merge.Merged)
		})
	}
}

func TestPullRequestsService_List(t *testing.T) {
	createdAt := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 10, 11, 14, 30, 0, 0, time.UTC)

	state := "open"
	tests := []struct {
		name           string
		owner          string
		repoName       string
		opts           *PullRequestListOptions
		expectedUrl    string
		responseBody   string
		expectedResult []*PullRequest
	}{
		{
			name:     "List of PRs",
			owner:    "octocat",
			repoName: "Hello-World",
			opts: &PullRequestListOptions{
				ListOptions: &ListOptions{Page: 1, PerPage: 30},
				State:       &state,
			},
			expectedUrl: "/repos/octocat/Hello-World/pulls?page=1&per_page=30&state=open",
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
			expectedResult: []*PullRequest{
				{
					Id:        1,
					Title:     "PR 1",
					Number:    1,
					State:     "open",
					CreatedAt: &Timestamp{createdAt},
					UpdatedAt: &Timestamp{updatedAt},
				},
				{
					Id:        2,
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
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedUrl, r.URL.String())
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			prs, _, err := client.PullRequests.List(context.Background(), tt.owner, tt.repoName, tt.opts)
			require.NoError(t, err)
			require.NotNil(t, prs)
			assert.Len(t, prs, len(tt.expectedResult))

			for i := range prs {
				assert.Equal(t, tt.expectedResult[i].Id, prs[i].Id)
				assert.Equal(t, tt.expectedResult[i].Title, prs[i].Title)

				if tt.expectedResult[i].CreatedAt != nil {
					require.NotNil(t, prs[i].CreatedAt)
					assert.Equal(t, tt.expectedResult[i].CreatedAt.Format(time.RFC3339), prs[i].CreatedAt.Format(time.RFC3339))
				} else {
					assert.Nil(t, prs[i].CreatedAt)
				}

				if tt.expectedResult[i].UpdatedAt != nil {
					require.NotNil(t, prs[i].UpdatedAt)
					assert.Equal(t, tt.expectedResult[i].UpdatedAt.Format(time.RFC3339), prs[i].UpdatedAt.Format(time.RFC3339))
				} else {
					assert.Nil(t, prs[i].UpdatedAt)
				}
			}
		})
	}
}
