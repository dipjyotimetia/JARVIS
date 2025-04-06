package github

import (
	"context"
	"errors"

	"github.com/google/go-github/v70/github"
	"golang.org/x/oauth2"
)

type Client struct {
	ctx    context.Context
	client *github.Client
}

func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		ctx:    ctx,
		client: github.NewClient(tc),
	}, nil
}
