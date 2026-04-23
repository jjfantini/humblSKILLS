package adapters

// PreferredDefaults returns the pre-selected adapters when the user hasn't
// explicitly chosen any. Used by both the install TUI's pre-check state and
// the non-interactive `--yes` path so the two stay symmetric.
//
// Policy (issue #84): Cursor can read ~/.claude/skills directly when its
// "Include Third-Party Plugins, Skills and other configs" setting is on,
// so installing to both by default creates duplicate copies that drift.
// When both claude-code and cursor are detected, we prefer claude-code
// and drop cursor from the defaults. Cursor stays a visible opt-in in the
// TUI and remains explicit via `--platform cursor`.
//
// Profile defaults always win — a user who saved a preference gets it back.
func PreferredDefaults(adapterList []Adapter, detected map[string]bool, profileDefaults []string) []string {
	if len(profileDefaults) > 0 {
		known := NameSet(adapterList)
		out := make([]string, 0, len(profileDefaults))
		for _, name := range profileDefaults {
			if _, ok := known[name]; ok {
				out = append(out, name)
			}
		}
		return out
	}

	out := make([]string, 0, len(adapterList))
	for _, a := range adapterList {
		if detected[a.Name] {
			out = append(out, a.Name)
		}
	}

	if detected["claude-code"] && detected["cursor"] {
		filtered := make([]string, 0, len(out))
		for _, name := range out {
			if name != "cursor" {
				filtered = append(filtered, name)
			}
		}
		out = filtered
	}

	return out
}
