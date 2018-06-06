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
			_, err := os.Stdout.WriteString(l + "\n")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	} else {
		usage()
	}
}
