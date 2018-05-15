package main

import (
	"fmt"

	"github.com/willf/pad"
	padUtf8 "github.com/willf/pad/utf8"
)

func main() {
	fmt.Println(pad.Right("Hello", 20, "!"))
	fmt.Println(padUtf8.Left("Exit now", 20, "→"))
}
