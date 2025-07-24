package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type UsersService struct {
	client *Client
}

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

func (s *UsersService) Get(ctx context.Context, username string) (*User, error) {
	path := fmt.Sprintf("users/%s", username)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return user, nil
}

func (s *UsersService) GetAuthenticated(ctx context.Context) (*User, error) {
	path := "user"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return user, nil
}

type UsersListOptions struct {
	Since int
	*ListOptions
}

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
		return nil, nil, fmt.Errorf("request creating error: %w", err)
	}

	users := new([]*User)
	res, err := s.client.Do(ctx, req, users)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *users, res, nil
}

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

func (s *UsersService) UpdateAuthenticated(ctx context.Context, body UserUpdateRequest) (*User, error) {
	path := "user"
	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, fmt.Errorf("request creating error: %w", err)
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return user, nil
}

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
		return nil, nil, fmt.Errorf("request creating error: %w", err)
	}

	users := new([]*User)
	res, err := s.client.Do(ctx, req, users)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *users, res, nil
}

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
		return nil, nil, fmt.Errorf("request creating error: %w", err)
	}

	users := new([]*User)
	res, err := s.client.Do(ctx, req, users)
	if err != nil {
		return nil, res, fmt.Errorf("repsponse parsing error: %w", err)
	}

	return *users, res, nil
}

func (s *UsersService) Follow(ctx context.Context, username string) error {
	path := fmt.Sprintf("user/following/%s", username)

	req, err := s.client.NewRequest(http.MethodPut, path, nil)
	if err != nil {
		return fmt.Errorf("request creating error: %w", err)
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("repsponse parsing error: %w", err)
	}

	return nil
}

func (s *UsersService) Unfollow(ctx context.Context, username string) error {
	path := fmt.Sprintf("user/following/%s", username)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("request creating error: %w", err)
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return fmt.Errorf("repsponse parsing error: %w", err)
	}

	return nil
}
