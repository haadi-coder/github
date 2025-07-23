package github

import (
	"context"
	"fmt"
	"net/http"
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
type UsersListOptions struct {
	Since int
	*ListOptions
}

func (s *UsersService) Get(ctx context.Context, username string) (*User, error) {
	url := s.client.baseUrl.JoinPath("users/", username).String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) GetAuthenticated(ctx context.Context) (*User, error) {
	url := s.client.baseUrl.JoinPath("user").String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) List(ctx context.Context, opts *UsersListOptions) ([]*User, *Response, error) {
	rawUrl := s.client.baseUrl.JoinPath("users")

	if opts != nil {
		q := rawUrl.Query()
		opts.paginateQuery(q)
		if opts.Since != 0 {
			q.Set("since", fmt.Sprintf("%d", opts.Since))
		}

		rawUrl.RawQuery = q.Encode()
	}

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
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

func (s *UsersService) UpdateAuthenticated(ctx context.Context, body UserUpdateRequest) (*User, error) {
	url := s.client.baseUrl.JoinPath("user").String()
	req, err := s.client.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.Do(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) ListAuthenticatedUserFollowers(ctx context.Context, opts *ListOptions) ([]*User, *Response, error) {
	rawUrl := s.client.baseUrl.JoinPath("user", "followers")

	if opts != nil {
		q := rawUrl.Query()
		opts.paginateQuery(q)
		rawUrl.RawQuery = q.Encode()
	}

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
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

func (s *UsersService) ListAuthenticatedUserFollowings(ctx context.Context, opts *ListOptions) ([]*User, *Response, error) {
	rawUrl := s.client.baseUrl.JoinPath("user/following")

	if opts != nil {
		q := rawUrl.Query()
		opts.paginateQuery(q)
		rawUrl.RawQuery = q.Encode()
	}

	url := rawUrl.String()
	req, err := s.client.NewRequest(http.MethodGet, url, nil)
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

func (s *UsersService) Follow(ctx context.Context, username string) error {
	url := s.client.baseUrl.JoinPath("user", "following", username).String()
	req, err := s.client.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (s *UsersService) Unfollow(ctx context.Context, username string) error {
	url := s.client.baseUrl.JoinPath("user", "following", username).String()
	req, err := s.client.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	if _, err = s.client.Do(ctx, req, nil); err != nil {
		return err
	}

	return nil
}
