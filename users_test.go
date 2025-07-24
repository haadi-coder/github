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

func TestUsersService_Get(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		responseStatus int
		responseBody   string
		expectedUser   *User
	}{
		{
			name:           "Get user",
			path:           "/users/testuser",
			method:         "GET",
			responseStatus: http.StatusOK,
			responseBody:   `{"id":1,"login":"testuser","name":"Test User"}`,
			expectedUser: &User{
				Id:    1,
				Login: "testuser",
				Name:  "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, tt.method, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			user, err := client.User.Get(context.Background(), "testuser")
			if tt.expectedUser != nil {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUser.Id, user.Id)
				assert.Equal(t, tt.expectedUser.Login, user.Login)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUsersService_GetAuthenticated(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		token          string
		responseStatus int
		responseBody   string
		expectedUser   *User
	}{
		{
			name:           "Get Authenticated user",
			path:           "/user",
			method:         "GET",
			token:          "test-token",
			responseStatus: http.StatusOK,
			responseBody:   `{"id":2,"login":"authuser","name":"Auth User"}`,
			expectedUser: &User{
				Id:    2,
				Login: "authuser",
				Name:  "Auth User",
			},
		},
		{
			name:           "Without token",
			path:           "/user",
			method:         "GET",
			token:          "",
			responseStatus: http.StatusUnauthorized,
			responseBody:   `{"message":"Unauthorized"}`,
			expectedUser:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, tt.method, r.Method)
				if tt.token != "" {
					assert.Equal(t, "Bearer "+tt.token, r.Header.Get("Authorization"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL), WithToken(tt.token))

			user, err := client.User.GetAuthenticated(context.Background())
			if tt.expectedUser != nil {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUser.Id, user.Id)
				assert.Equal(t, tt.expectedUser.Login, user.Login)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUsersService_List(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		opts           *UsersListOptions
		responseStatus int
		responseBody   string
		expectedUsers  []*User
		expectedQuery  string
	}{
		{
			name:           "Users list with pagination",
			path:           "/users",
			method:         "GET",
			opts:           &UsersListOptions{Since: 1, ListOptions: &ListOptions{Page: 2, PerPage: 50}},
			responseStatus: http.StatusOK,
			responseBody:   `[{"id":1,"login":"user1"},{"id":2,"login":"user2"}]`,
			expectedUsers: []*User{
				{Id: 1, Login: "user1"},
				{Id: 2, Login: "user2"},
			},
			expectedQuery: "page=2&per_page=50&since=1",
		},
		{
			name:           "Empty Users list",
			path:           "/users",
			method:         "GET",
			opts:           &UsersListOptions{ListOptions: &ListOptions{Page: 1, PerPage: 30}},
			responseStatus: http.StatusOK,
			responseBody:   `[]`,
			expectedUsers:  []*User{},
			expectedQuery:  "page=1&per_page=30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.expectedQuery, r.URL.RawQuery)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL))

			userList, resp, err := client.User.List(context.Background(), tt.opts)
			require.NoError(t, err)
			require.NotNil(t, resp)
			if len(tt.expectedUsers) > 0 {
				assert.Len(t, userList, len(tt.expectedUsers))
				for i := range userList {
					assert.Equal(t, tt.expectedUsers[i].Id, userList[i].Id)
					assert.Equal(t, tt.expectedUsers[i].Login, userList[i].Login)
				}
			} else {
				assert.Empty(t, userList)
			}
		})
	}
}

func TestUsersService_UpdateAuthenticated(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		token          string
		requestBody    string
		responseStatus int
		responseBody   string
		expectedUser   *User
	}{
		{
			name:           "Update profile",
			path:           "/user",
			method:         "PATCH",
			token:          "test-token",
			requestBody:    `{"name":"New Name","email":"new@example.com"}`,
			responseStatus: http.StatusOK,
			responseBody:   `{"id":3,"name":"New Name","email":"new@example.com"}`,
			expectedUser: &User{
				Id:    3,
				Name:  "New Name",
				Email: "new@example.com",
			},
		},
		{
			name:           "Permission denied",
			path:           "/user",
			method:         "PATCH",
			token:          "invalid-token",
			requestBody:    `{"name":"New Name"}`,
			responseStatus: http.StatusForbidden,
			responseBody:   `{"message":"Forbidden"}`,
			expectedUser:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				if tt.token != "" {
					assert.Equal(t, "Bearer "+tt.token, r.Header.Get("Authorization"))
				}

				body, _ := io.ReadAll(r.Body)
				assert.JSONEq(t, tt.requestBody, string(body))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL), WithToken(tt.token))

			var body UserUpdateRequest
			_ = json.Unmarshal([]byte(tt.requestBody), &body)

			user, err := client.User.UpdateAuthenticated(context.Background(), body)
			if tt.expectedUser != nil {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUser.Id, user.Id)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUsersService_ListAuthenticatedUserFollowers(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		token          string
		opts           *ListOptions
		responseStatus int
		responseBody   string
		expectedUsers  []*User
		expectedQuery  string
	}{
		{
			name:           "Followers list",
			path:           "/user/followers",
			method:         "GET",
			token:          "test-token",
			opts:           &ListOptions{Page: 2, PerPage: 50},
			responseStatus: http.StatusOK,
			responseBody:   `[{"id":1,"login":"follower1"},{"id":2,"login":"follower2"}]`,
			expectedUsers: []*User{
				{Id: 1, Login: "follower1"},
				{Id: 2, Login: "follower2"},
			},
			expectedQuery: "page=2&per_page=50",
		},
		{
			name:           "Empty followers list",
			path:           "/user/followers",
			method:         "GET",
			token:          "test-token",
			opts:           &ListOptions{Page: 1, PerPage: 30},
			responseStatus: http.StatusOK,
			responseBody:   `[]`,
			expectedUsers:  []*User{},
			expectedQuery:  "page=1&per_page=30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.expectedQuery, r.URL.RawQuery)
				if tt.token != "" {
					assert.Equal(t, "Bearer "+tt.token, r.Header.Get("Authorization"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL), WithToken(tt.token))

			followers, resp, err := client.User.ListAuthenticatedUserFollowers(context.Background(), tt.opts)
			require.NoError(t, err)
			require.NotNil(t, resp)
			if len(tt.expectedUsers) > 0 {
				assert.Len(t, followers, len(tt.expectedUsers))
				for i := range followers {
					assert.Equal(t, tt.expectedUsers[i].Id, followers[i].Id)
					assert.Equal(t, tt.expectedUsers[i].Login, followers[i].Login)
				}
			} else {
				assert.Empty(t, followers)
			}
		})
	}
}

func TestUsersService_FollowUnfollow(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		token          string
		responseStatus int
		isFollow       bool
	}{
		{
			name:           "Подписка на пользователя",
			path:           "/user/following/testuser",
			method:         "PUT",
			token:          "test-token",
			responseStatus: http.StatusNoContent,
			isFollow:       true,
		},
		{
			name:           "Отписка от пользователя",
			path:           "/user/following/testuser",
			method:         "DELETE",
			token:          "test-token",
			responseStatus: http.StatusNoContent,
			isFollow:       false,
		},
		{
			name:           "Ошибка при подписке (токен отсутствует)",
			path:           "/user/following/testuser",
			method:         "PUT",
			token:          "",
			responseStatus: http.StatusUnauthorized,
			isFollow:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.path, r.URL.Path)
				assert.Equal(t, tt.method, r.Method)
				if tt.token != "" {
					assert.Equal(t, "Bearer "+tt.token, r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseStatus != http.StatusNoContent {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
				}
			}))
			defer ts.Close()

			client := NewClient(WithBaseURl(ts.URL), WithToken(tt.token))

			var err error
			if tt.isFollow {
				err = client.User.Follow(context.Background(), "testuser")
			} else {
				err = client.User.Unfollow(context.Background(), "testuser")
			}

			if tt.responseStatus == http.StatusNoContent {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
