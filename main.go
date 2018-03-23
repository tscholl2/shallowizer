package main

import (
	"encoding/json"
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
	for _, path := range arr {
		info := repoInfo{path, 0, 0}
		info.StartBytes, err = dirSizeInBytes(path)
		if err != nil {
			log.Fatal(err)
		}
		if err := shallowize(path); err != nil {
			log.Fatal(err)
		}
		info.EndBytes, err = dirSizeInBytes(path)
		if err != nil {
			log.Fatal(err)
		}
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
	if _, err := cwd(exec.Command("git", "pull", "--depth=1")).Output(); err != nil {
		return err
	}
	_, err := cwd(exec.Command("git", "gc", "--prune=all")).Output()
	return err
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
