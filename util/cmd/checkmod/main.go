package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/blang/semver/v4"
)

type Module struct {
	Path    string
	Version string
}

type GoMod struct {
	Module  ModPath
	Go      string
	Require []Require
	Exclude []Module
	Replace []Replace
	Retract []Retract
}

type ModPath struct {
	Path       string
	Deprecated string
}

type Require struct {
	Path     string
	Version  string
	Indirect bool
}

type Replace struct {
	Old Module
	New Module
}

type Retract struct {
	Low       string
	High      string
	Rationale string
}

func main() {
	if goBinaryPath, err := exec.LookPath("go"); err != nil {
		log.Fatal("this command requires the go compiler")
	} else {
		log.Printf("found go binary at %q", goBinaryPath)
	}
	if gitBinaryPath, err := exec.LookPath("git"); err != nil {
		log.Fatal("this command requires git")
	} else {
		log.Printf("found git binary at %q", gitBinaryPath)
	}
	mod, err := exec.Command("go", "mod", "edit", "-json").Output()
	if err != nil {
		log.Fatal(err)
	}
	var module GoMod
	if err := json.Unmarshal(mod, &module); err != nil {
		log.Fatal(err)
	}
	modPath := module.Module.Path
	log.Printf("found module %q", modPath)
	nestedModules := findNestedModules(&module, modPath)
	log.Printf("found %d nested modules", len(nestedModules))
	nestedModuleTags, err := findNestedModuleTags(modPath, nestedModules)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("found %d most recent nested module tags", len(nestedModuleTags))
	var exitStatus int
	for modPath, modTag := range nestedModuleTags {
		diffPath := strings.TrimSuffix(path.Join(path.Dir(modTag), strings.Split(path.Base(modTag), ".")[0]), "/v0")
		log.Printf("checking for module differences between HEAD and %s at %q", modTag, diffPath)
		output, err := exec.Command("git", "diff", "--stat", "HEAD", modTag, diffPath).CombinedOutput()
		if err != nil {
			log.Fatal(string(output))
		}
		if len(output) > 0 {
			exitStatus = 1
			log.Printf("module %q was changed since release %s", modPath, modTag)
			fmt.Println(string(output))
		}
	}
	if exitStatus > 0 {
		log.Println("one or more modules requires releasing")
	}
	os.Exit(exitStatus)
}

func findNestedModules(module *GoMod, modPath string) []Require {
	result := []Require{}
	for _, req := range module.Require {
		if strings.HasPrefix(req.Path, modPath) {
			log.Printf("found nested module %q", req.Path)
			result = append(result, req)
		}
	}
	return result
}

func findNestedModuleTags(rootMod string, nestedModules []Require) (map[string]string, error) {
	result := make(map[string]string)
	tagOutput, err := exec.Command("git", "tag", "--list").CombinedOutput()
	if err != nil {
		log.Println(string(tagOutput))
		return nil, err
	}
	for _, mod := range nestedModules {
		prefix := strings.TrimPrefix(mod.Path, rootMod+"/")
		scanner := bufio.NewScanner(bytes.NewReader(tagOutput))
		var topTag string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, prefix) {
				if topTag == "" {
					topTag = line
					continue
				}
				gt, err := versionGreater(line, topTag)
				if err != nil {
					return nil, err
				}
				if gt {
					topTag = line
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		if topTag == "" {
			log.Fatalf("could not find any tags for nested module %q", mod.Path)
		}
		result[mod.Path] = topTag
		log.Printf("most recent tag for module %q is %q", mod.Path, topTag)
	}
	return result, nil
}

func versionGreater(lhs, rhs string) (bool, error) {
	lhsVersion, err := semver.Parse(strings.TrimPrefix(path.Base(lhs), "v"))
	if err != nil {
		return false, fmt.Errorf("couldn't parse module version %q: %s", lhs, err)
	}
	rhsVersion, err := semver.Parse(strings.TrimPrefix(path.Base(rhs), "v"))
	if err != nil {
		return false, fmt.Errorf("couldn't parse module version %q: %s", rhs, err)
	}
	return lhsVersion.GT(rhsVersion), nil
}
