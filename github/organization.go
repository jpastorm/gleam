package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Organization struct {
	Token string
	Org   string
}

func NewOrganization() *Organization {
	return &Organization{}
}

func (o *Organization) Setup(token, username string) {
	o.Token = token
	o.Org = username
}

func (o Organization) ListRepositories() ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", o.Org)
	cmd := exec.Command("curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", o.Token), url)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var repos []struct {
		URL string `json:"ssh_url"`
	}

	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, err
	}

	var repoURLs []string
	for _, repo := range repos {
		repoURLs = append(repoURLs, repo.URL)
	}

	return repoURLs, nil
}

func (o Organization) CloneRepositories(repoURLs []string) error {
	for _, repoURL := range repoURLs {
		err := os.RemoveAll(getRepoName(repoURL))
		if err != nil {
			return err
		}

		gitCmd := fmt.Sprintf("git clone %s", repoURL)

		cmd := exec.Command("bash", "-c", gitCmd)

		cmd.Env = append(cmd.Env, fmt.Sprintf("GITHUB_TOKEN=%s", o.Token))

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func getRepoName(repoURL string) string {
	parts := strings.Split(repoURL, "/")
	return strings.Split(parts[len(parts)-1], ".")[0]
}
