package selfupdate

import (
	"errors"
	"fmt"
	"os"
)

// ReplaceBinary atomically swaps newBinaryPath onto targetPath, the classic
// self-update trick that also works while targetPath is the currently
// running executable: rename the old file out of the way first (POSIX lets
// you rename or delete an open file; Windows lets you rename — just not
// delete — an open file), then rename the new one into place, then
// best-effort clean up the old file (ignored if it's still locked, e.g. on
// Windows while this process is still running — it gets cleaned up next
// run).
func ReplaceBinary(targetPath, newBinaryPath string) error {
	old := targetPath + ".old"
	_ = os.Remove(old) // leftover from a previous run that couldn't clean up

	if err := os.Rename(targetPath, old); err != nil {
		return fmt.Errorf("rename current binary aside: %w", err)
	}
	if err := os.Rename(newBinaryPath, targetPath); err != nil {
		// Best-effort revert so we never leave the user with no binary.
		if rerr := os.Rename(old, targetPath); rerr != nil {
			return fmt.Errorf("install new binary: %w (revert also failed: %v)", err, rerr)
		}
		return fmt.Errorf("install new binary: %w", err)
	}
	_ = os.Remove(old) // best-effort; a locked .old on Windows is cleaned up next run
	return nil
}

// IsPermissionError reports whether err indicates ReplaceBinary failed
// because targetPath's directory isn't writable by the current user, so
// callers can suggest sudo or Homebrew instead of printing a raw OS error.
func IsPermissionError(err error) bool {
	return errors.Is(err, os.ErrPermission)
}
