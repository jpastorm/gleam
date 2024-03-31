package github

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type PublicUser struct {
	Username string
}

func NewPublicUser() *PublicUser {
	return &PublicUser{}
}

func (p *PublicUser) Setup(token, username string) {
	p.Username = username
}

func (p PublicUser) ListRepositories() ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos", p.Username)

	cmd := exec.Command("curl", "-s", url)

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

func (p PublicUser) CloneRepositories(repoURLs []string) error {
	for _, repoURL := range repoURLs {
		err := os.RemoveAll(getRepoName(repoURL))
		if err != nil {
			return err
		}
		gitCmd := fmt.Sprintf("git clone %s", repoURL)

		cmd := exec.Command("bash", "-c", gitCmd)

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
