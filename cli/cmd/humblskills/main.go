// Command humblskills installs agentskills.io-format skills into whichever
// agent platform you use — Claude Code, Cursor, and friends.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "humblskills:", err)
		os.Exit(1)
	}
}
