package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoriesService_Get(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		responseStatus int
		responseBody   string
		expected       *Repository
		expectError    bool
	}{
		{
			name:           "Repository Create",
			owner:          "octocat",
			repoName:       "Hello-World",
			responseStatus: http.StatusOK,
			responseBody: `{
                "id": 1296269,
                "name": "Hello-World",
                "full_name": "octocat/Hello-World",
                "private": false,
                "description": "My first repository on GitHub!",
                "owner": {"id": 583231, "login": "octocat"},
				"html_url": "",
				"fork": false,
				"url": "",
				"clone_url": "",
				"mirror_url": "",
				"language": "",
				"forks_count": 0,
				"stargazers_count": 0,
				"watchers_count": 0,
				"size": 0,
				"default_branch": "",
				"open_issues_count": 0,
				"is_template": false,
				"topics": null,
				"has_issues": false,
				"has_projects": false,
				"has_wiki": false,
				"has_pages": false,
				"has_downloads":false,
				"archived": false,
				"disabled": false,
				"visibility": "",
				"pushed_at": null,
				"created_at": null,
				"updated_at": null
            }`,
			expected: &Repository{
				Id:          1296269,
				Name:        "Hello-World",
				Description: "My first repository on GitHub!",
				Fullname:    "octocat/Hello-World",
				Private:     false,
				IsTemplate:  false,
				Owner:       &User{Id: 583231, Login: "octocat"},
			},
			expectError: false,
		},
		{
			name:           "Repository Not Found",
			owner:          "octocat",
			repoName:       "Unknown",
			responseStatus: http.StatusNotFound,
			responseBody: `{
                "message": "Not Found"
            }`,
			expected:    nil,
			expectError: true,
		},
		{
			name:           "Auth Error",
			owner:          "octocat",
			repoName:       "Hello-World",
			responseStatus: http.StatusUnauthorized,
			responseBody: `{
                "message": "Unauthorized"
            }`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/"+tt.owner+"/"+tt.repoName, r.URL.Path)
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
			require.NoError(t, err)

			repo, err := client.Repositories.Get(context.Background(), tt.owner, tt.repoName)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)

			assert.Equal(t, tt.expected, repo)
		})
	}
}

func TestRepositoriesService_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           RepositoryCreateRequest
		responseStatus int
		responseBody   string
		expected       *Repository
		expectError    bool
	}{
		{
			name: "Repository create",
			body: RepositoryCreateRequest{
				Name:        "Hello-World",
				Description: "My first repo",
				Private:     false,
			},
			responseStatus: http.StatusCreated,
			responseBody: `{
                "id": 1296269,
                "name": "Hello-World",
                "description": "My first repo"
            }`,
			expected: &Repository{
				Id:          1296269,
				Name:        "Hello-World",
				Description: "My first repo",
			},
			expectError: false,
		},
		{
			name: "Validation failed",
			body: RepositoryCreateRequest{
				Name: "",
			},
			responseStatus: http.StatusBadRequest,
			responseBody: `{
                "message": "Validation Failed"
            }`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/user/repos", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				reqBody, _ := io.ReadAll(r.Body)
				var request RepositoryCreateRequest
				_ = json.Unmarshal(reqBody, &request)
				assert.Equal(t, tt.body, request)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
			require.NoError(t, err)

			repo, err := client.Repositories.Create(context.Background(), tt.body)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)

			assert.Equal(t, tt.expected, repo)
		})
	}
}

func TestRepositoriesService_Update(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		body           RepositoryUpdateRequest
		responseStatus int
		responseBody   string
		expected       *Repository
		expectError    bool
	}{
		{
			name:     "Description update",
			owner:    "octocat",
			repoName: "Hello-World",
			body: RepositoryUpdateRequest{
				Description: "Updated description",
			},
			responseStatus: http.StatusOK,
			responseBody: `{
                "id": 1296269,
                "description": "Updated description"
            }`,
			expected: &Repository{
				Id:          1296269,
				Description: "Updated description",
			},
			expectError: false,
		},
		{
			name:     "Permission denied",
			owner:    "octocat",
			repoName: "Hello-World",
			body: RepositoryUpdateRequest{
				Description: "Restricted",
			},
			responseStatus: http.StatusForbidden,
			responseBody: `{
                "message": "Forbidden"
            }`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/"+tt.owner+"/"+tt.repoName, r.URL.Path)
				assert.Equal(t, "PATCH", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				reqBody, _ := io.ReadAll(r.Body)
				var request RepositoryUpdateRequest
				_ = json.Unmarshal(reqBody, &request)
				assert.Equal(t, tt.body, request)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
			require.NoError(t, err)

			repo, err := client.Repositories.Update(context.Background(), tt.owner, tt.repoName, tt.body)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, repo)

			assert.Equal(t, tt.expected, repo)
		})
	}
}

func TestRepositoriesService_Delete(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		repoName       string
		responseStatus int
		expectError    bool
	}{
		{
			name:           "Delete",
			owner:          "octocat",
			repoName:       "Hello-World",
			responseStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "Repository not found",
			owner:          "octocat",
			repoName:       "Unknown",
			responseStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/repos/"+tt.owner+"/"+tt.repoName, r.URL.Path)
				assert.Equal(t, "DELETE", r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				w.WriteHeader(tt.responseStatus)
				if tt.responseStatus != http.StatusNoContent {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"message":"Not Found"}`))
				}
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
			require.NoError(t, err)

			err = client.Repositories.Delete(context.Background(), tt.owner, tt.repoName)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRepositoriesService_List(t *testing.T) {
	typ := "all"
	tests := []struct {
		name           string
		owner          string
		opts           *RepositoryListOptions
		expectedURL    string
		responseStatus int
		responseBody   string
		expectedResult []*Repository
		expectError    bool
	}{
		{
			name:  "List with pagination",
			owner: "octocat",
			opts: &RepositoryListOptions{
				ListOptions: &ListOptions{
					Page:    2,
					PerPage: 50,
				},
				Type: &typ,
			},
			expectedURL:    "/users/octocat/repos?page=2&per_page=50&type=all",
			responseStatus: http.StatusOK,
			responseBody:   `[{"id":1,"name":"repo1"},{"id":2,"name":"repo2"}]`,
			expectedResult: []*Repository{
				{Id: 1, Name: "repo1"},
				{Id: 2, Name: "repo2"},
			},
			expectError: false,
		},
		{
			name:           "Empty list",
			owner:          "octocat",
			opts:           nil,
			expectedURL:    "/users/octocat/repos",
			responseStatus: http.StatusOK,
			responseBody:   "[]",
			expectedResult: []*Repository{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.String())
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
			require.NoError(t, err)

			repos, _, err := client.Repositories.List(context.Background(), tt.owner, tt.opts)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, repos, len(tt.expectedResult))
			assert.Equal(t, tt.expectedResult, repos)
		})
	}
}

func TestRepositoriesService_ListContributors(t *testing.T) {
	anon := "true"
	tests := []struct {
		name           string
		owner          string
		repoName       string
		opts           *RepositoryListOptions
		expectedURL    string
		responseStatus int
		responseBody   string
		expectedResult []*User
		expectError    bool
	}{
		{
			name:     "Contributors list",
			owner:    "octocat",
			repoName: "Hello-World",
			opts: &RepositoryListOptions{
				ListOptions: &ListOptions{
					Page:    1,
					PerPage: 30,
				},
				Anon: &anon,
			},
			expectedURL:    "/repos/octocat/Hello-World/contributors?anon=true&page=1&per_page=30",
			responseStatus: http.StatusOK,
			responseBody:   `[{"id":1,"login":"contrib1"},{"id":2,"login":"contrib2"}]`,
			expectedResult: []*User{
				{Id: 1, Login: "contrib1"},
				{Id: 2, Login: "contrib2"},
			},
			expectError: false,
		},
		{
			name:           "Empty list",
			owner:          "octocat",
			repoName:       "Hello-World",
			opts:           nil,
			expectedURL:    "/repos/octocat/Hello-World/contributors",
			responseStatus: http.StatusOK,
			responseBody:   "[]",
			expectedResult: []*User{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedURL, r.URL.String())
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
			require.NoError(t, err)

			users, _, err := client.Repositories.ListContributors(context.Background(), tt.owner, tt.repoName, tt.opts)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, users, len(tt.expectedResult))
			assert.Equal(t, tt.expectedResult, users)
		})
	}
}
