package github

import "github.com/google/go-github/v59/github"

type Comment struct {
	Id     int
	Body   string
	PostId int
}

func (c *Client) ChekPRComment(prNumber int) (string, error) {
	prComment, _, err := c.client.PullRequests.ListComments(c.ctx, owner, repo, prNumber, &github.PullRequestListCommentsOptions{})
	if err != nil {
		return "", err
	}
	for _, comment := range prComment {
		// if *comment.User.Login == c.user {
		// 	return *comment.Body, nil
		// }
		*comment.Body = "test"
		return *comment.Body, nil
	}
	return "", nil
}
