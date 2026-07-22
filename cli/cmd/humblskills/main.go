// Command humblskills installs agentskills.io-format skills into whichever
// agent platform you use — Claude Code, Cursor, and friends.
package main

import (
	"fmt"
	"os"

	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func main() {
	err := newRootCmd().Execute()
	// Tear down the interactive session program (if the router started one) and
	// restore the terminal on every path — os.Exit below skips defers, so this
	// runs explicitly first.
	tui.Shutdown()
	if err != nil {
		fmt.Fprintln(os.Stderr, "humblskills:", err)
		os.Exit(1)
	}
}
