package github

import (
	"errors"
	"log/slog"
	"os"

	"github.com/google/go-github/v70/github"
)

var (
	owner = os.Getenv("GITHUB_OWNER")
	repo  = os.Getenv("GITHUB_REPO")
)

func (c *Client) ChekPR() (string, error) {
	pr, _, err := c.client.PullRequests.List(c.ctx, owner, repo,
		&github.PullRequestListOptions{
			State:       "open",
			Head:        "main",
			Sort:        "updated",
			Direction:   "desc",
			ListOptions: github.ListOptions{PerPage: 100},
		})
	if err != nil {
		return "", err
	}
	if len(pr) == 0 {
		return "", errors.New("no pull requests found")
	}
	for _, p := range pr {
		ok, err := checkPrReference(p)
		if err != nil {
			return "", err
		}
		if ok {
			prDetails, _, err := c.client.PullRequests.Get(c.ctx, owner, repo, p.GetNumber())
			if err != nil {
				return "", err
			}
			if prDetails.GetMerged() {
				return "", errors.New("pull request already merged")
			}
			diff, _, err := c.client.PullRequests.GetRaw(c.ctx, owner, repo, *prDetails.Number, github.RawOptions{
				Type: github.Diff,
			})
			if err != nil {
				slog.Error("Error fetching diff", "error", err)
				return "", err
			}
			if len(diff) == 0 {
				return "", errors.New("no diff found")
			}
			return diff, nil
		}
	}
	return "", errors.New("no suitable pull request found")
}

func checkPrReference(pr *github.PullRequest) (bool, error) {
	switch pr.GetHead().GetRef() {
	case "main":
		return false, nil
	case "master":
		return false, nil
	case "develop":
		return true, nil
	case "release":
		return false, nil
	case "hotfix":
		return true, nil
	case "feature":
		return true, nil
	case "bugfix":
		return true, nil
	case "chore":
		return true, nil
	case "docs":
		return false, nil
	case "test":
		return true, nil
	default:
		return false, nil
	}
}
