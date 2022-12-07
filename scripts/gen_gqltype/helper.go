package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"golang.org/x/tools/go/vcs"
)

type goModule struct {
	Require []struct {
		Path    string
		Version string
	}
}

func noop() {}

func WithGoModuleSource(packagePath string) (string, func(), error) {

	cleanup := noop

	cmd := exec.Command("go", "mod", "edit", "-json")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", cleanup, err
	}
	decoder := json.NewDecoder(&stdout)
	var module goModule
	if err := decoder.Decode(&module); err != nil {
		return "", cleanup, err
	}
	// order by module path length desc
	sort.Slice(module.Require, func(i, j int) bool { return len(module.Require[i].Path) > len(module.Require[j].Path) })

	var modPath, version string
	found := false
	for _, dep := range module.Require {
		if strings.HasPrefix(packagePath, dep.Path) {
			modPath, version = dep.Path, dep.Version
			found = true
			break
		}
	}
	if !found {
		return "", cleanup, fmt.Errorf("could not find package %s in go.mod", packagePath)
	}
	repo, err := vcs.RepoRootForImportPath(modPath, false)
	if err != nil {
		return "", cleanup, err
	}
	dir := base64.StdEncoding.EncodeToString([]byte(modPath + version))
	repodir := path.Join("./", dir)
	if err := repo.VCS.CreateAtRev(repodir, repo.Repo, version); err != nil {
		return "", cleanup, err
	}
	cleanup = func() {
		if err := os.RemoveAll(repodir); err != nil {
			log.Fatal(err)
		}
	}
	result := strings.Replace(packagePath, repo.Root, repodir, 1)
	return result, cleanup, nil
}
