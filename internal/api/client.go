package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("gitee api error: status=%d body=%s", e.StatusCode, e.Body)
}

func (c *Client) Request(ctx context.Context, method, endpoint string, query map[string]string, body any, out any) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, endpoint)

	q := u.Query()
	for k, v := range query {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(payload)}
	}
	if out == nil || len(payload) == 0 {
		return nil
	}
	return json.Unmarshal(payload, out)
}

type User struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

type Repo struct {
	ID          int64  `json:"id"`
	FullName    string `json:"full_name"`
	HumanName   string `json:"human_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	DefaultBr   string `json:"default_branch"`
}

type Issue struct {
	ID      int64  `json:"id"`
	Number  string `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	User    User   `json:"user"`
}

type PullRequest struct {
	ID      int64  `json:"id"`
	Number  int64  `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	User    User   `json:"user"`
}

type Release struct {
	ID      int64  `json:"id"`
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var u User
	if err := c.Request(ctx, http.MethodGet, "user", nil, nil, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *Client) ListRepos(ctx context.Context, org, visibility string, page, perPage int) ([]Repo, error) {
	query := map[string]string{
		"type":      visibility,
		"sort":      "updated",
		"direction": "desc",
		"page":      fmt.Sprintf("%d", page),
		"per_page":  fmt.Sprintf("%d", perPage),
	}
	endpoint := "user/repos"
	if org != "" {
		endpoint = fmt.Sprintf("orgs/%s/repos", org)
	}
	var repos []Repo
	if err := c.Request(ctx, http.MethodGet, endpoint, query, nil, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*Repo, error) {
	var r Repo
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s", owner, repo), nil, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) CreateRepo(ctx context.Context, name, desc, org string, private bool) (*Repo, error) {
	endpoint := "user/repos"
	if org != "" {
		endpoint = fmt.Sprintf("orgs/%s/repos", org)
	}
	body := map[string]any{"name": name, "description": desc, "private": private}
	var r Repo
	if err := c.Request(ctx, http.MethodPost, endpoint, nil, body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) ListIssues(ctx context.Context, owner, repo, state string, page, perPage int) ([]Issue, error) {
	query := map[string]string{
		"state":    state,
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}
	var issues []Issue
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s/issues", owner, repo), query, nil, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Client) GetIssue(ctx context.Context, owner, repo, number string) (*Issue, error) {
	var issue Issue
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s/issues/%s", owner, repo, number), nil, nil, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string) (*Issue, error) {
	payload := map[string]any{"title": title, "body": body}
	var issue Issue
	if err := c.Request(ctx, http.MethodPost, fmt.Sprintf("repos/%s/%s/issues", owner, repo), nil, payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) UpdateIssueState(ctx context.Context, owner, repo, number, state string) (*Issue, error) {
	payload := map[string]any{"state": state}
	var issue Issue
	if err := c.Request(ctx, http.MethodPatch, fmt.Sprintf("repos/%s/%s/issues/%s", owner, repo, number), nil, payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) CreateIssueComment(ctx context.Context, owner, repo, number, body string) error {
	payload := map[string]any{"body": body}
	return c.Request(ctx, http.MethodPost, fmt.Sprintf("repos/%s/%s/issues/%s/comments", owner, repo, number), nil, payload, nil)
}

func (c *Client) ListPRs(ctx context.Context, owner, repo, state string, page, perPage int) ([]PullRequest, error) {
	query := map[string]string{
		"state":    state,
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}
	var prs []PullRequest
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s/pulls", owner, repo), query, nil, &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

func (c *Client) GetPR(ctx context.Context, owner, repo string, number int64) (*PullRequest, error) {
	var pr PullRequest
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), nil, nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) CreatePR(ctx context.Context, owner, repo, title, head, base, body string) (*PullRequest, error) {
	payload := map[string]any{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
	}
	var pr PullRequest
	if err := c.Request(ctx, http.MethodPost, fmt.Sprintf("repos/%s/%s/pulls", owner, repo), nil, payload, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) MergePR(ctx context.Context, owner, repo string, number int64, title string) error {
	payload := map[string]any{"merge_commit_message": title}
	return c.Request(ctx, http.MethodPut, fmt.Sprintf("repos/%s/%s/pulls/%d/merge", owner, repo, number), nil, payload, nil)
}

func (c *Client) ClosePR(ctx context.Context, owner, repo string, number int64) (*PullRequest, error) {
	payload := map[string]any{"state": "closed"}
	var pr PullRequest
	if err := c.Request(ctx, http.MethodPatch, fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number), nil, payload, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) CreatePRComment(ctx context.Context, owner, repo string, number int64, body string) error {
	payload := map[string]any{"body": body}
	return c.Request(ctx, http.MethodPost, fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, number), nil, payload, nil)
}

func (c *Client) ListReleases(ctx context.Context, owner, repo string, page, perPage int) ([]Release, error) {
	query := map[string]string{
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}
	var releases []Release
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s/releases", owner, repo), query, nil, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (c *Client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*Release, error) {
	var rel Release
	if err := c.Request(ctx, http.MethodGet, fmt.Sprintf("repos/%s/%s/releases/tags/%s", owner, repo, tag), nil, nil, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (c *Client) CreateRelease(ctx context.Context, owner, repo, tag, name, body, target string, draft bool) (*Release, error) {
	payload := map[string]any{
		"tag_name":         tag,
		"name":             name,
		"body":             body,
		"target_commitish": target,
		"draft":            draft,
	}
	var rel Release
	if err := c.Request(ctx, http.MethodPost, fmt.Sprintf("repos/%s/%s/releases", owner, repo), nil, payload, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (c *Client) DeleteRelease(ctx context.Context, owner, repo string, id int64) error {
	return c.Request(ctx, http.MethodDelete, fmt.Sprintf("repos/%s/%s/releases/%d", owner, repo, id), nil, nil, nil)
}
