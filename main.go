package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

var (
	relReleasePattern = regexp.MustCompile(`^(rel|release)/`)
	featFeaturePattern = regexp.MustCompile(`^(feat|feature)/`)
)

func main() {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		fatal("error opening repository: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		fatal("error getting HEAD: %v", err)
	}

	if tag := getHeadTag(repo, head.Hash()); tag != "" {
		fmt.Println(tag)
		return
	}

	latestTag, commitsSince, err := findLatestTagAndCount(repo, head.Hash())
	if err != nil {
		fatal("no semver tag found in history: %v", err)
	}

	branchName, err := getBranchName(repo, head)
	if err != nil {
		fatal("error determining branch: %v", err)
	}

	nextVersion := calculateNextVersion(latestTag, branchName, calculateBranchID(branchName), commitsSince)
	fmt.Println(nextVersion)
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func getHeadTag(repo *git.Repository, headHash plumbing.Hash) string {
	tagRefs, err := repo.Tags()
	if err != nil {
		return ""
	}

	var match string
	_ = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		commitHash, err := resolveTagCommit(repo, ref)
		if err != nil {
			return nil
		}
		if *commitHash == headHash {
			name := ref.Name().Short()
			if _, err := semver.NewVersion(name); err == nil {
				match = name
				return storer.ErrStop
			}
		}
		return nil
	})
	return match
}

func resolveTagCommit(repo *git.Repository, ref *plumbing.Reference) (*plumbing.Hash, error) {
	obj, err := repo.Object(plumbing.AnyObject, ref.Hash())
	if err != nil {
		return nil, err
	}
	switch o := obj.(type) {
	case *object.Commit:
		h := o.Hash
		return &h, nil
	case *object.Tag:
		commit, err := o.Commit()
		if err != nil {
			return nil, err
		}
		h := commit.Hash
		return &h, nil
	}
	return nil, fmt.Errorf("unexpected object type for tag %s", ref.Name().Short())
}

func findLatestTagAndCount(repo *git.Repository, headHash plumbing.Hash) (*semver.Version, int, error) {
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, 0, err
	}

	tagByCommit := make(map[plumbing.Hash]*semver.Version)
	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		commitHash, err := resolveTagCommit(repo, ref)
		if err != nil {
			return nil
		}
		name := ref.Name().Short()
		v, err := semver.NewVersion(name)
		if err == nil {
			tagByCommit[*commitHash] = v
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	commit, err := repo.CommitObject(headHash)
	if err != nil {
		return nil, 0, err
	}

	commitsSince := 0
	for {
		if v, ok := tagByCommit[commit.Hash]; ok {
			return v, commitsSince, nil
		}
		parents := commit.ParentHashes
		if len(parents) == 0 {
			break
		}
		commit, err = repo.CommitObject(parents[0])
		if err != nil {
			return nil, 0, err
		}
		commitsSince++
	}

	return nil, 0, fmt.Errorf("no semver tag found in history")
}

func getBranchName(repo *git.Repository, head *plumbing.Reference) (string, error) {
	name := head.Name().String()
	if strings.HasPrefix(name, "refs/heads/") {
		return head.Name().Short(), nil
	}

	branches, err := repo.Branches()
	if err != nil {
		return "", err
	}

	var match string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if ref.Hash() == head.Hash() {
			match = ref.Name().Short()
			return storer.ErrStop
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if match != "" {
		return match, nil
	}

	if ref := os.Getenv("GITHUB_REF_NAME"); ref != "" {
		return ref, nil
	}
	if ref := os.Getenv("GITHUB_REF"); ref != "" {
		parts := strings.Split(ref, "/")
		if len(parts) > 2 {
			return strings.Join(parts[2:], "/"), nil
		}
	}

	return "", fmt.Errorf("could not determine branch name")
}

func calculateBranchID(branchName string) string {
	h := sha256.Sum256([]byte(branchName))
	id := (uint64(h[0]) | uint64(h[1])<<8 | uint64(h[2])<<16 | uint64(h[3])<<24) % 10000
	return fmt.Sprintf("b%04d", id)
}

func calculateNextVersion(latestTag *semver.Version, branchName, branchID string, commitsSince int) string {
	var major, minor, patch uint64

	switch {
	case relReleasePattern.MatchString(branchName):
		major = latestTag.Major() + 1
		minor = 0
		patch = 0
	case featFeaturePattern.MatchString(branchName):
		major = latestTag.Major()
		minor = latestTag.Minor() + 1
		patch = 0
	default:
		major = latestTag.Major()
		minor = latestTag.Minor()
		patch = latestTag.Patch() + 1
	}

	if branchName == "main" || branchName == "master" {
		return fmt.Sprintf("v%d.%d.%d+%d", major, minor, patch, commitsSince)
	}

	return fmt.Sprintf("v%d.%d.%d.%s+%d", major, minor, patch, branchID, commitsSince)
}
