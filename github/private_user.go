package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type PrivateUser struct {
	Token string
}

func NewPrivateUser() *PrivateUser {
	return &PrivateUser{}
}

func (p *PrivateUser) Setup(token, username string) {
	p.Token = token
}

func (p PrivateUser) ListRepositories() ([]string, error) {
	url := "https://api.github.com/user/repos"

	cmd := exec.Command("curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", p.Token), url)

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

func (p PrivateUser) CloneRepositories(repoURLs []string) error {
	for _, repoURL := range repoURLs {
		err := os.RemoveAll(getRepoName(repoURL))
		if err != nil {
			return err
		}

		gitCmd := fmt.Sprintf("git clone %s", repoURL)

		cmd := exec.Command("bash", "-c", gitCmd)

		cmd.Env = append(cmd.Env, fmt.Sprintf("GITHUB_TOKEN=%s", p.Token))

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
