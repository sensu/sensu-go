package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sensu/sensu-go/internal/astutil"
)

var packageDeclRe = regexp.MustCompile(`([\/\*]*[pP]ackage)[ ]+([_a-z0-9]+)`)

// copyPackage naively copies a package from one place to another, updating its
// package declaration.
func copyPackage(fromPackage, toPackage string) error {
	from := astutil.PackagePath(fromPackage)
	to := astutil.PackagePath(toPackage)
	if err := os.MkdirAll(to, os.ModeDir|0755); err != nil {
		return fmt.Errorf("couldn't copy package: %s", err)
	}
	files, err := ioutil.ReadDir(from)
	if err != nil {
		return fmt.Errorf("couldn't copy package: %s", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), "_generated.go") || strings.HasSuffix(f.Name(), "_generated_test.go") {
			// Generated code from internal packages does not belong in
			// versioned packages.
			continue
		}
		oldFileName := filepath.Join(from, f.Name())
		newFileName := filepath.Join(to, f.Name())
		packageName := filepath.Base(to)
		if strings.HasSuffix(f.Name(), ".go") {
			if err := copyGoFile(oldFileName, newFileName, packageName); err != nil {
				return fmt.Errorf("couldn't copy package: %s", err)
			}
		} else if strings.HasSuffix(f.Name(), ".proto") {
			if err := copyProtoFile(oldFileName, newFileName, toPackage); err != nil {
				return fmt.Errorf("couldn't copy package: %s", err)
			}
		} else {
			if err := copyFile(oldFileName, newFileName); err != nil {
				return fmt.Errorf("couldn't copy package: %s", err)
			}
		}
	}

	return nil
}

var protoPackageDeclRe = regexp.MustCompile(`(package) [_\.a-z0-9]+`)
var protoGoPackageDeclRe = regexp.MustCompile(`(option go_package =) "[_a-z0-9]+"`)

func copyProtoFile(from, to, toPackage string) error {
	b, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}

	newProtoPackageName := strings.Replace(toPackage, "/", ".", -1)
	newProtoPackageName = strings.Replace(newProtoPackageName, "-", "_", -1)
	newBytes := protoPackageDeclRe.ReplaceAll(b, []byte(fmt.Sprintf(`$1 %s`, newProtoPackageName)))

	newProtoGoPackageName := path.Base(toPackage)
	newBytes = protoGoPackageDeclRe.ReplaceAll(newBytes, []byte(fmt.Sprintf(`$1 "%s"`, newProtoGoPackageName)))

	return ioutil.WriteFile(to, newBytes, 0644)
}

func copyGoFile(from, to, packageName string) error {
	b, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}
	newBytes := packageDeclRe.ReplaceAll(b, []byte(fmt.Sprintf(`$1 %s`, packageName)))
	return ioutil.WriteFile(to, newBytes, 0644)
}

func copyFile(from, to string) error {
	b, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(to, b, 0644)
}
