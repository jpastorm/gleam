package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type PrivateUser struct {
	Token    string
	Username string
}

func NewPrivateUser() *PrivateUser {
	return &PrivateUser{}
}

func (p *PrivateUser) Setup(token, username string) {
	p.Token = token
	p.Username = username
}

func (p PrivateUser) ListRepositories() ([]string, error) {
	url := "https://api.github.com/user/repos"

	cmd := exec.Command("curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", p.Token), url)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var repos []struct {
		URL string `json:"clone_url"`
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

func (p PrivateUser) CloneRepositories(repoURLs []string) error {
	for _, repoURL := range repoURLs {
		repoName := getRepoName(repoURL)

		err := os.RemoveAll(repoName)
		if err != nil {
			return err
		}

		authURL := fmt.Sprintf("https://%s:%s@%s", p.Username, p.Token, strings.TrimPrefix(repoURL, "https://"))

		cmd := exec.Command("git", "clone", authURL)

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
