package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	aw "github.com/deanishe/awgo"
)

var wf *aw.Workflow

func main() {
	wf = aw.New()
	wf.Run(run)
}

func run() {
	wf.Args()
	flag.Parse()
	query := flag.Arg(0)

	client := BitriseAPIClient{
		client: &http.Client{Timeout: time.Second * 10},
		// TODO: set this in keychain in a dedicated subcommand
		authToken: "",
	}

	apps, err := client.getAppsByTitle(query)
	if err != nil {
		wf.FatalError(err)
	}

	for _, app := range apps {
		wf.NewItem(fmt.Sprintf("%s/%s", app.Owner.Name, app.Title)).
			Subtitle(app.RepoInfo()).
			Arg(app.URL()).
			UID(app.URL()).
			Valid(true)
	}

	wf.WarnEmpty("No matching items", "Try a different query?")
	wf.SendFeedback()
}

type App struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	RepoOwner string `json:"repo_owner"`
	RepoSlug  string `json:"repo_slug"`
	Owner     Owner  `json:"owner"`
}

func (a App) URL() string {
	return fmt.Sprintf("http://app.bitrise.io/app/%s", a.Slug)
}

func (a App) RepoInfo() string {
	return fmt.Sprintf("%s/%s", a.RepoOwner, a.RepoSlug)
}

type Owner struct {
	Name string
}

type BitriseAPIClient struct {
	authToken string
	client    *http.Client
}

func (c *BitriseAPIClient) getAppsByTitle(title string) ([]App, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.bitrise.io/v0.1/apps?limit=10&title=%s&sort_by=last_build_at", title), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", c.authToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bitrise API request failed with status code: %d", resp.StatusCode)
	}

	var response struct {
		Data []App `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
