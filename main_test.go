package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestGetHeadTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		expects string
	}{
		{name: "semver tag", tag: "v1.2.3", expects: "v1.2.3"},
		{name: "non-semver tag", tag: "not-a-version", expects: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoDir := t.TempDir()
			repo, err := git.PlainInit(repoDir, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(repoDir, "fixture"), []byte("test"), 0o644); err != nil {
				t.Fatal(err)
			}

			wt, err := repo.Worktree()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := wt.Add("."); err != nil {
				t.Fatal(err)
			}
			commit, err := wt.Commit("test", &git.CommitOptions{
				Author: &object.Signature{Name: "Test", Email: "test@example.com"},
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := repo.CreateTag(tt.tag, commit, nil); err != nil {
				t.Fatal(err)
			}

			tag := getHeadTag(repo, commit)
			if tag != tt.expects {
				t.Fatalf("getHeadTag() = %q, want %q", tag, tt.expects)
			}
		})
	}
}

func TestGetHeadTagIsFalseForUntaggedCommit(t *testing.T) {
	repoDir := t.TempDir()
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "fixture"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Add("."); err != nil {
		t.Fatal(err)
	}
	commit, err := wt.Commit("test", &git.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if tag := getHeadTag(repo, commit); tag != "" {
		t.Fatalf("getHeadTag() = %q, want no tag", tag)
	}
}
