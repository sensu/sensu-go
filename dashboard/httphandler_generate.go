// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	dir := http.Dir("build")
	err := vfsgen.Generate(dir, vfsgen.Options{
		Filename:     "httpassets_release.go",
		PackageName:  "dashboard",
		BuildTags:    "release",
		VariableName: "httpAssets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
