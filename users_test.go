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

func TestUsersService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/testuser", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"login":"testuser","name":"Test User"}`))
	}))
	defer ts.Close()

	client := NewClient(WithBaseURl(ts.URL))

	user, err := client.User.Get(context.Background(), "testuser")
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.Id)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "Test User", user.Name)
}

func TestUsersService_GetAuthenticated(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":2,"login":"authuser","name":"Auth User"}`))
	}))
	defer ts.Close()

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	user, err := client.User.GetAuthenticated(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), user.Id)
	assert.Equal(t, "authuser", user.Login)
	assert.Equal(t, "Auth User", user.Name)
}

func TestUsersService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "page=2&per_page=50&since=1", r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"login":"user1"},{"id":2,"login":"user2"}]`))
	}))
	defer ts.Close()

	opts := &UsersListOptions{
		Since: 1,
		ListOptions: &ListOptions{
			Page:    2,
			PerPage: 50,
		},
	}

	client := NewClient(WithBaseURl(ts.URL))

	userList, resp, err := client.User.List(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, userList, 2)
	assert.Equal(t, "user1", userList[0].Login)
	assert.Equal(t, "user2", userList[1].Login)
}

func TestUsersService_UpdateAuthenticated(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		body, _ := io.ReadAll(r.Body)
		assert.JSONEq(t, `{"name":"New Name","email":"new@example.com"}`, string(body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":3,"name":"New Name","email":"new@example.com"}`))
	}))
	defer ts.Close()

	body := UserUpdateRequest{
		Name:  "New Name",
		Email: "new@example.com",
	}

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	user, err := client.User.UpdateAuthenticated(context.Background(), body)
	require.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
	assert.Equal(t, "new@example.com", user.Email)
}

func TestUsersService_ListAuthenticatedUserFollowers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/followers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "page=1&per_page=30", r.URL.RawQuery)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"login":"follower1"},{"id":2,"login":"follower2"}]`))
	}))
	defer ts.Close()

	opts := &ListOptions{
		Page:    1,
		PerPage: 30,
	}

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	followers, resp, err := client.User.ListAuthenticatedUserFollowers(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, followers, 2)
	assert.Equal(t, "follower1", followers[0].Login)
	assert.Equal(t, "follower2", followers[1].Login)
}

func TestUsersService_Follow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/following/testuser", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	err := client.User.Follow(context.Background(), "testuser")
	require.NoError(t, err)
}

func TestUsersService_Unfollow(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user/following/testuser", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(WithBaseURl(ts.URL), WithToken("test-token"))

	err := client.User.Unfollow(context.Background(), "testuser")
	require.NoError(t, err)
}
