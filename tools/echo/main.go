package main

import (
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Println("usage: echo ...[string]")
}

func main() {
	if len(os.Args) > 1 {
		line := strings.Join(os.Args[1:], " ")
		lines := strings.Split(line, `\n`)
		for _, l := range lines {
			fmt.Println(l)
		}
	} else {
		usage()
	}
}
