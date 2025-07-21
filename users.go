package main

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

type UserRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Blog            string `json:"blog"`
	TwitterUsername string `json:"twitter_username"`
	Company         string `json:"company"`
	Location        string `json:"location"`
	Hireable        bool   `json:"hireable"`
	Bio             string `json:"bio"`
}

type UsersListOptions struct {
	Since int
	ListOptions
}

func (s *UsersService) Get(ctx context.Context, username string) (*User, error) {
	path := "users/" + username
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.fetch(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) GetAuthenticated(ctx context.Context) (*User, error) {
	path := "user"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.fetch(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) List(ctx context.Context, opts *UsersListOptions) ([]*User, *Response, error) {
	path := "users"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		q := req.URL.Query()
		opts.paginateQuery(q)
		if opts.Since != 0 {
			q.Set("since", fmt.Sprintf("%d", opts.Since))
		}

		req.URL.RawQuery = q.Encode()
	}

	users := new([]*User)
	res, err := s.client.fetch(ctx, req, users)
	if err != nil {
		return nil, res, err
	}

	return *users, res, nil
}

func (s *UsersService) UpdateAuthenticated(ctx context.Context, body UserRequest) (*User, error) {
	path := "user"
	req, err := s.client.NewRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}

	user := new(User)
	if _, err := s.client.fetch(ctx, req, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersService) ListAuthenticatedUserFollowers(ctx context.Context, opts *ListOptions) ([]*User, *Response, error) {
	path := "user/followers"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		q := req.URL.Query()
		opts.paginateQuery(q)
		req.URL.RawQuery = q.Encode()
	}

	users := new([]*User)
	res, err := s.client.fetch(ctx, req, users)
	if err != nil {
		return nil, res, err
	}

	return *users, res, nil
}

func (s *UsersService) ListAuthenticatedUserFollowings(ctx context.Context, opts *ListOptions) ([]*User, *Response, error) {
	path := "user/following"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		q := req.URL.Query()
		opts.paginateQuery(q)
		req.URL.RawQuery = q.Encode()
	}

	users := new([]*User)
	res, err := s.client.fetch(ctx, req, users)
	if err != nil {
		return nil, res, err
	}

	return *users, res, nil
}

func (s *UsersService) Follow(ctx context.Context, username string) error {
	path, err := url.JoinPath("user/following", username)
	if err != nil {
		return err
	}

	req, err := s.client.NewRequest(http.MethodPut, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.fetch(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersService) Unfollow(ctx context.Context, username string) error {
	path, err := url.JoinPath("user/following", username)
	if err != nil {
		return err
	}

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.fetch(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
