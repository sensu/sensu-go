//+build !windows

package main

import (
	"github.com/spf13/cobra"
)

func addRootPlatformArguments(*cobra.Command)  {}
func addStartPlatformArguments(*cobra.Command) {}
