package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

type repoInfo struct {
	Path       string
	StartBytes int64
	EndBytes   int64
	Error      string
}

func main() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		gopath = filepath.Join(user.HomeDir, "go")
	}
	results := map[string]repoInfo{}
	arr, err := listRepos(filepath.Join(gopath, "src"))
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range arr[:1] {
		var info repoInfo
		info.Path = path
		info.StartBytes, _ = dirSizeInBytes(path)
		if err := shallowize(path); err != nil {
			info.Error = err.Error()
			fmt.Fprintln(os.Stderr, "error with repo: ", path)
			fmt.Fprintln(os.Stderr, err)
		}
		info.EndBytes, _ = dirSizeInBytes(path)
		results[path] = info
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(results)
}

func shallowize(repo string) error {
	cwd := func(c *exec.Cmd) *exec.Cmd {
		c.Dir = repo
		return c
	}
	// skip dirty repos
	if out, err := cwd(exec.Command("git", "diff", "--shortstat")).CombinedOutput(); err != nil || string(out) != "" {
		return fmt.Errorf("might be dirty:\n\tout: %s\n\terr: %s", out, err)
	}
	// find remote origin
	out, err := cwd(exec.Command("git", "remote", "get-url", "origin")).CombinedOutput()
	url := string(out)
	if err != nil || url == "" {
		return fmt.Errorf("unable to read url:\n\tout: %s\n\terr: %s", out, err)
	}
	dir, err := ioutil.TempDir("", "shallow-clone")
	if err != nil {
		return fmt.Errorf("unable to create temp dir:\n\t%s", err)
	}
	defer os.RemoveAll(dir)
	// do a shallow clone
	if out, err := cwd(exec.Command("git", "clone", "--depth=1", url, dir)).CombinedOutput(); err != nil {
		return fmt.Errorf("unable to re-clone:\n\tout: %s\n\terr: %s", out, err)
	}
	// TODO: copy temp files to repo path
	return nil
}

func listRepos(root string) (repos []string, err error) {
	return repos, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ok, err := isGitRepo(path); !ok {
			return err
		}
		repos = append(repos, path)
		return filepath.SkipDir
	})
}

func isGitRepo(folder string) (bool, error) {
	info, err := os.Stat(filepath.Join(folder, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	return info != nil && info.IsDir(), err
}

func dirSizeInBytes(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
