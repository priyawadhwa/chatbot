package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
)

const (
	user                     = "priyawadhwa"
	project                  = "Project"
	organization             = "priya-test"
	columnName               = "Waiting Code Review"
	mediaTypeProjectsPreview = "application/vnd.github.inertia-preview+json"
)

func accessToken() string {
	return os.Getenv("GITHUB_ACCESS_TOKEN")
}

type GithubClient struct {
	ctx    context.Context
	client *github.Client
}

func NewGithubClient() *GithubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken()},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return &GithubClient{
		ctx:    ctx,
		client: client,
	}
}

func (c *GithubClient) RetrieveCards() ([]Card, error) {
	projects, _, err := c.client.Organizations.ListProjects(c.ctx, organization, nil)
	if err != nil {
		return nil, err
	}
	var projectID int64
	for _, p := range projects {
		if p.Name != nil && *p.Name == project {
			projectID = p.GetID()
		}
	}

	columns, err := c.columns(projectID)
	var columnID int64
	for _, c := range columns {
		if c.Name == columnName {
			columnID = c.ID
		}
	}
	return c.cards(columnID)
}

type Column struct {
	Name string
	ID   int64
}

// columns returns all columns within the associated projectID
func (c *GithubClient) columns(projectId int64) ([]Column, error) {
	u := fmt.Sprintf("/projects/%v/columns", projectId)
	req, err := c.client.NewRequest("GET", u, nil)
	req.Header.Set("Accept", mediaTypeProjectsPreview)
	var columns []Column
	_, err = c.client.Do(c.ctx, req, &columns)
	return columns, err
}

type Creator struct {
	Login string
}

type Card struct {
	Name    string
	ID      int64
	Note    string
	Creator Creator
}

func (c *GithubClient) cards(columnId int64) ([]Card, error) {
	u := fmt.Sprintf("/projects/columns/%v/cards", columnId)
	req, err := c.client.NewRequest("GET", u, nil)
	req.Header.Set("Accept", mediaTypeProjectsPreview)
	var cards []Card
	_, err = c.client.Do(c.ctx, req, &cards)
	return cards, err
}
