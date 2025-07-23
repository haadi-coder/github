package github

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositories_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/octocat/Hello-World", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "id": 1296269,
            "name": "Hello-World",
            "full_name": "octocat/Hello-World",
            "private": false,
            "html_url": "https://github.com/octocat/Hello-World ",
            "description": "My first repository on GitHub!",
            "fork": false,
            "owner": {"id": 583231, "login": "octocat"}
        }`))
	}))
	defer ts.Close()

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
	repo, err := client.Repositories.Get(context.Background(), "octocat", "Hello-World")
	require.NoError(t, err)
	assert.Equal(t, int64(1296269), repo.Id)
	assert.Equal(t, "octocat", repo.Owner.Login)
	assert.Equal(t, "octocat/Hello-World", repo.Fullname)
}

func TestRepositoriesService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/repos", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{"name":"Hello-World","description":"My first repo"}`, string(body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1296269,"name":"Hello-World"}`))
	}))
	defer ts.Close()

	req := RepositoryCreateRequest{
		Name:        "Hello-World",
		Description: "My first repo",
	}

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))
	repo, err := client.Repositories.Create(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(1296269), repo.Id)
	assert.Equal(t, "Hello-World", repo.Name)
}

func TestRepositoriesService_Edit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/octocat/Hello-World", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{"description":"Updated description"}`, string(body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1296269,"description":"Updated description"}`))
	}))
	defer ts.Close()

	req := RepositoryUpdateRequest{
		Description: "Updated description",
	}

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	repo, err := client.Repositories.Edit(context.Background(), "octocat", "Hello-World", req)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", repo.Description)
}

func TestRepositoriesService_Delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/octocat/Hello-World", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	err := client.Repositories.Delete(context.Background(), "octocat", "Hello-World")
	require.NoError(t, err)
}

func TestRepositoriesService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/octocat/repos", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.RawQuery, "type=all")
		assert.Contains(t, r.URL.RawQuery, "page=2")
		assert.Contains(t, r.URL.RawQuery, "per_page=50")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"name":"repo1"},{"id":2,"name":"repo2"}]`))
	}))
	defer ts.Close()

	all := "all"
	opts := &RepositoryListOptions{
		ListOptions: &ListOptions{
			Page:    2,
			PerPage: 50,
		},
		Type: &all,
	}

	client := NewClient(WithBaseURl(ts.URL))
	repoList, resp, err := client.Repositories.List(context.Background(), "octocat", opts)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, repoList, 2)
	assert.Equal(t, "repo1", repoList[0].Name)
	assert.Equal(t, "repo2", repoList[1].Name)
}

func TestRepositoriesService_ListContributors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/octocat/Hello-World/contributors", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.RawQuery, "anon=true")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"login":"contrib1"},{"id":2,"login":"contrib2"}]`))
	}))
	defer ts.Close()

	anon := "true"
	opts := &RepositoryListOptions{
		ListOptions: nil,

		Anon: &anon,
	}

	client := NewClient(WithBaseURl(ts.URL))

	contributors, resp, err := client.Repositories.ListContributors(context.Background(), "octocat", "Hello-World", opts)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, contributors, 2)
	assert.Equal(t, "contrib1", contributors[0].Login)
	assert.Equal(t, "contrib2", contributors[1].Login)
}
