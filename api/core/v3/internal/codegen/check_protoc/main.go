package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

const protocVersion = "3.9.1"

var libprotoRe = regexp.MustCompile(fmt.Sprintf(`^libprotoc %s\s$`, protocVersion))

func main() {
	out, err := exec.Command("protoc", "--version").Output()
	if err != nil {
		log.Fatalf("error: looking for protoc %s: %s", protocVersion, err)
	}
	if !libprotoRe.Match(out) {
		log.Fatalf("bad protoc version: want %q, got %q", libprotoRe.String(), string(out))
	}
}
