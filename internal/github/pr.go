package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type PullRequest struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	Base    PRRef  `json:"base"`
	Head    PRRef  `json:"head"`
	User    User   `json:"user"`
}

type PRRef struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo Repo   `json:"repo"`
}

type Repo struct {
	FullName string `json:"full_name"`
}

type User struct {
	Login string `json:"login"`
}

func (c *Client) FindPullRequestsForBranch(ctx context.Context, owner, repo, branch string) ([]PullRequest, error) {
	query := url.Values{}
	query.Set("state", "open")
	query.Set("head", owner+":"+branch)

	path := fmt.Sprintf("/repos/%s/%s/pulls?%s", url.PathEscape(owner), url.PathEscape(repo), query.Encode())

	var prs []PullRequest
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &prs); err != nil {
		return nil, err
	}
	return prs, nil
}
