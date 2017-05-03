package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	d, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("usage: sleep seconds")
		os.Exit(1)
	}
	time.Sleep(time.Duration(d) * time.Second)
}
