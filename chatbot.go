package chatbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/chat/v1"
)

const (
	user                     = "container-tools-bot"
	project                  = "Container Tools Team In Progress"
	organization             = "GoogleContainerTools"
	columnName               = "Waiting Code Review"
	mediaTypeProjectsPreview = "application/vnd.github.inertia-preview+json"
	url                      = "https://github.com/orgs/GoogleContainerTools/projects/1"
)

// Chatbot receives requests and responds with the number of PRs awaiting review
// in the GoogleContainerTools project
func Chatbot(w http.ResponseWriter, r *http.Request) {
	space, err := retrieveSpace(r)
	if err != nil {
		log.Print(err)
		return
	}
	resp, err := generateResponseMessage(space)
	if err != nil {
		log.Print(err)
		return
	}
	if err := respondToChat(resp, space); err != nil {
		log.Print(err)
		return
	}
}

// retrieves the space (hangouts chat ID) from the request
func retrieveSpace(r *http.Request) (string, error) {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	var msg chat.Message
	err = json.Unmarshal(contents, &msg)
	if err != nil {
		return "", err
	}
	if msg.Space == nil {
		return "", errors.New("no space provided in request")
	}
	return msg.Space.Name, nil
}

// gets the number of PRs awaiting code review and generates a reponse message
func generateResponseMessage(space string) (*chat.Message, error) {
	client := NewGithubClient()
	cards, err := client.RetrieveCards()
	var msg string
	if err != nil {
		log.Printf("errors responding to chat: %v", err)
		return nil, err
	}
	msg = fmt.Sprintf("There are %d PRs awaiting code review \n%s", len(cards), url)
	return &chat.Message{
		Text: msg,
	}, nil
}

// responds to a chat with the response message
func respondToChat(resp *chat.Message, space string) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://chat.googleapis.com/v1/%s/messages", space)

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey())

	client := &http.Client{}

	_, err = client.Do(req)
	return err
}

func apiKey() string {
	return os.Getenv("API_KEY")
}

func accessToken() string {
	return os.Getenv("GITHUB_ACCESS_TOKEN")
}

type GithubClient struct {
	ctx    context.Context
	client *github.Client
}

// NewGithubClient returns a client with the necessary auth
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
