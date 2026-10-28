package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// checkPlatformsKept fails when a skill lists fewer platforms than it did in
// baseFile (the target branch's registry.json). An installed copy on a
// dropped platform can no longer be refreshed, so skills keep every platform
// they shipped; the fix is to add it back, not to drop it.
//
// An empty platforms[] means "every adapter", so narrowing from empty to an
// explicit list is a drop too. A dropped platform the binary no longer ships
// an adapter for is ignored: nothing can be installed there anyway.
// Skills are matched by name or previous_names, so a rename isn't a drop.
func checkPlatformsKept(skillsDir, baseFile string, knownAdapters map[string]struct{}) error {
	data, err := os.ReadFile(baseFile)
	if err != nil {
		return fmt.Errorf("--platforms-base: %w", err)
	}
	var base struct {
		Skills []struct {
			Name      string   `json:"name"`
			Platforms []string `json:"platforms"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return fmt.Errorf("--platforms-base %s: %w", baseFile, err)
	}
	basePlatforms := make(map[string][]string, len(base.Skills))
	for _, s := range base.Skills {
		basePlatforms[s.Name] = s.Platforms
	}

	parsed, err := walkSkills(skillsDir)
	if err != nil {
		return err
	}

	var drops []string
	for _, p := range parsed {
		before, ok := basePlatforms[p.fm.Name]
		for _, old := range p.fm.PreviousNames() {
			if ok {
				break
			}
			before, ok = basePlatforms[old]
		}
		if !ok {
			continue // new skill
		}
		after := p.fm.Platforms()
		if len(after) == 0 {
			continue // every adapter
		}
		if len(before) == 0 {
			for a := range knownAdapters {
				before = append(before, a)
			}
		}
		keep := make(map[string]struct{}, len(after))
		for _, a := range after {
			keep[a] = struct{}{}
		}
		var dropped []string
		for _, b := range before {
			_, kept := keep[b]
			_, known := knownAdapters[b]
			if !kept && known {
				dropped = append(dropped, b)
			}
		}
		if len(dropped) > 0 {
			sort.Strings(dropped)
			drops = append(drops, fmt.Sprintf("%s drops %s", p.fm.Name, strings.Join(dropped, ", ")))
		}
	}
	if len(drops) > 0 {
		sort.Strings(drops)
		return fmt.Errorf("skills must keep every platform they shipped; add these back to metadata.platforms:\n  - %s",
			strings.Join(drops, "\n  - "))
	}
	fmt.Printf("platforms kept (%d skills checked against %s)\n", len(parsed), baseFile)
	return nil
}
