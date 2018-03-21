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
		Filename:     "assets_generated.go",
		PackageName:  "dashboard",
		BuildTags:    "release",
		VariableName: "generatedAssets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
