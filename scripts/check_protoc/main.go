package main

import (
	"log"
	"os/exec"
	"regexp"
)

var libprotoRe = regexp.MustCompile(`^libprotoc 3.9.1\s$`)

func main() {
	out, err := exec.Command("protoc", "--version").Output()
	if err != nil {
		log.Fatal(err)
	}
	if !libprotoRe.Match(out) {
		log.Fatalf("bad protoc version: want %q, got %q", libprotoRe.String(), string(out))
	}
}
