package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// UsersService provides access to user-related API methods.
type UsersService struct {
	client *Client
}

// User represents a GitHub user.
// GitHub API docs: https://docs.github.com/en/rest/users/users
type User struct {
	Id          int64      `json:"id"`
	Login       string     `json:"login"`
	NodeID      string     `json:"node_id"`
	AvatarURL   string     `json:"avatar_url"`
	URL         string     `json:"url"`
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Company     string     `json:"company"`
	Blog        string     `json:"blog"`
	Location    string     `json:"location"`
	Email       string     `json:"email"`
	Hireable    bool       `json:"hireable"`
	Bio         string     `json:"bio"`
	PublicRepos int        `json:"public_repos"`
	Followers   int        `json:"followers"`
	Following   int        `json:"following"`
	CreatedAt   *Timestamp `json:"created_at"`
	UpdatedAt   *Timestamp `json:"updated_at"`
}

// Get retrieves information about a specific GitHub user by username.
// This method returns public profile information for any GitHub user,
// including their name, company, location, bio, and various statistics
// such as follower count and public repository count.
func (s *UsersService) Get(ctx context.Context, username string) (*User, error) {
	path := fmt.Sprintf("users/%s", username)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetAuthenticated retrieves information about the currently authenticated user.
// This method returns detailed profile information for the authenticated user,
// including private information that is only available when authenticated.
// It requires proper authentication credentials to be configured in the client.
func (s *UsersService) GetAuthenticated(ctx context.Context) (*User, error) {
	path := "user"

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UsersListOptions specifies the optional parameters to list users.
// GitHub API docs: https://docs.github.com/en/rest/users/users#list-users
type UsersListOptions struct {
	Since int
	*ListOptions
}

// List retrieves a list of GitHub users.
// This method returns a paginated list of GitHub users. You can use
// the Since parameter to specify the user ID to start listing from,
// which is useful for pagination through large sets of users.
func (s *UsersService) List(ctx context.Context, opts *UsersListOptions) ([]*User, *Response, error) {
	path := "users"

	if opts != nil {
		q := url.Values{}
		if opts.ListOptions != nil {
			opts.paginateQuery(q)
		}
		if opts.Since != 0 {
			q.Set("since", fmt.Sprintf("%d", opts.Since))
		}

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	users := new([]*User)
	res, err := s.client.Do(ctx, req, users)
	if err != nil {
		return nil, res, err
	}

	return *users, res, nil
}

// UserUpdateRequest represents the request body for updating user profile.
// GitHub API docs: https://docs.github.com/en/rest/users/users#update-the-authenticated-user
type UserUpdateRequest struct {
	Name            string `json:"name,omitempty"`
	Email           string `json:"email,omitempty"`
	Blog            string `json:"blog,omitempty"`
	TwitterUsername string `json:"twitter_username,omitempty"`
	Company         string `json:"company,omitempty"`
	Location        string `json:"location,omitempty"`
	Hireable        bool   `json:"hireable,omitempty"`
	Bio             string `json:"bio,omitempty"`
}

// UpdateAuthenticated updates the profile of the authenticated user.
// This method allows you to modify the profile information of the
// currently authenticated user, including name, email, company,
// location, bio, and other profile fields. Only the provided
// fields will be updated.
func (s *UsersService) UpdateAuthenticated(ctx context.Context, body UserUpdateRequest) (*User, error) {
	path := "user"

	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ListAuthenticatedUserFollowers retrieves the followers of the authenticated user.
// This method returns a list of users who are following the authenticated user.
// The results can be paginated using the ListOptions parameter.
func (s *UsersService) ListAuthenticatedUserFollowers(ctx context.Context, opts *ListOptions) ([]*User, *Response, error) {
	path := "user/followers"

	if opts != nil {
		q := url.Values{}
		opts.paginateQuery(q)

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	users := new([]*User)
	res, err := s.client.Do(ctx, req, users)
	if err != nil {
		return nil, res, err
	}

	return *users, res, nil
}

// ListAuthenticatedUserFollowings retrieves the users that the authenticated user is following.
// This method returns a list of users that the authenticated user is following.
// The results can be paginated using the ListOptions parameter.
func (s *UsersService) ListAuthenticatedUserFollowings(ctx context.Context, opts *ListOptions) ([]*User, *Response, error) {
	path := "user/following"

	if opts != nil {
		q := url.Values{}
		opts.paginateQuery(q)

		if len(q) != 0 {
			path += "?" + q.Encode()
		}
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	users := new([]*User)
	res, err := s.client.Do(ctx, req, users)
	if err != nil {
		return nil, res, err
	}

	return *users, res, nil
}

// Follow starts following a user.
// This method allows the authenticated user to follow another GitHub user.
// Once followed, the target user will appear in the authenticated user's
// following list, and the authenticated user will appear in the target
// user's followers list.
func (s *UsersService) Follow(ctx context.Context, username string) error {
	path := fmt.Sprintf("user/following/%s", username)

	req, err := s.client.NewRequest(http.MethodPut, path, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

// Unfollow stops following a user.
// This method allows the authenticated user to unfollow a GitHub user
// they were previously following. This will remove the relationship
// between the users.
func (s *UsersService) Unfollow(ctx context.Context, username string) error {
	path := fmt.Sprintf("user/following/%s", username)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}
