package github

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

// CheckVersion checks the latest version from GitHub repository
// and prints it to stdout
func CheckVersion() {
	latestVersion, err := GetLatestVersion()
	if err != nil {
		fmt.Println("Error checking version:", err)
		return
	}

	if latestVersion != nil {
		fmt.Println("Latest Tag:", latestVersion.String())
	}
}

// GetLatestVersion returns the latest version from GitHub repository
func GetLatestVersion() (*semver.Version, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting user information: %w", err)
	}

	sshKeyPath := filepath.Join(usr.HomeDir, ".ssh", "id_ed25519")

	sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("error loading SSH key: %w", err)
	}

	repo, err := gogit.Clone(memory.NewStorage(), memfs.New(), &gogit.CloneOptions{
		Auth:          sshAuth,
		URL:           "git@github.com:dipjyotimetia/jarvis.git",
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName("refs/heads/main"),
		SingleBranch:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}

	tagrefs, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", err)
	}

	versions := make([]*semver.Version, 0)
	tagrefs.ForEach(func(t *plumbing.Reference) error {
		tagName := t.Name().Short()
		v, err := semver.NewVersion(tagName)
		if err == nil {
			versions = append(versions, v)
		}
		return nil
	})

	if len(versions) == 0 {
		return nil, fmt.Errorf("no valid SemVer tags found")
	}

	sort.Sort(semver.Collection(versions))
	latestTag := versions[len(versions)-1]

	return latestTag, nil
}
