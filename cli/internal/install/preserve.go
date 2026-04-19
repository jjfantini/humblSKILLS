package install

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
)

// loadLocalPreserve reads SKILL.md at installPath and returns the preserve
// list the user currently has on disk, validated. The semantics are:
//
//   - ok == true: list is safe to apply (may be empty, which means "preserve
//     nothing"). Callers must honor the empty list - it's how users opt into
//     a clean overwrite of previously-preserved paths.
//   - ok == false: something went wrong parsing or validating; reason holds a
//     short diagnostic suitable for a warning event. Callers should fall back
//     to the registry's preserve list to avoid accidentally destroying data
//     over a broken YAML edit.
func loadLocalPreserve(installPath string) (entries []string, ok bool, reason string) {
	skillMD := filepath.Join(installPath, "SKILL.md")
	data, err := os.ReadFile(skillMD)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, "SKILL.md missing"
		}
		return nil, false, "SKILL.md read failed: " + err.Error()
	}
	fm, _, err := frontmatter.Parse(data)
	if err != nil {
		return nil, false, "SKILL.md frontmatter parse failed: " + err.Error()
	}
	if perrs := frontmatter.ValidatePreserve(fm.Preserve); len(perrs) > 0 {
		return nil, false, "invalid preserve list: " + strings.Join(perrs, "; ")
	}
	return fm.Preserve, true, ""
}

const frontmatterDelimiter = "---"

// mergePreserveIntoSkillMD rewrites path so its frontmatter's `preserve:` key
// matches preserve. Every other frontmatter key and the body below are kept
// byte-for-byte from the existing file, save for YAML re-serialization of the
// mapping itself (which may normalize whitespace and drops comments inside
// the YAML block).
//
// This is what lets users edit only the `preserve:` list and still receive
// upstream changes to the rest of SKILL.md on every update. Whenever the
// caller has determined the user "owns" their preserve list, it should call
// this against the per-target staging SKILL.md before the final replaceDir.
//
// When preserve is empty (nil or len 0), the key is written as an explicit
// empty flow sequence (`preserve: []`) so the next update still sees it as
// user-authored rather than defaulting back to the registry's list.
func mergePreserveIntoSkillMD(path string, preserve []string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	yamlStart, yamlEnd, bodyStart, err := frontmatterBounds(data)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data[yamlStart:yamlEnd], &root); err != nil {
		return fmt.Errorf("%s: parse frontmatter: %w", path, err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return fmt.Errorf("%s: unexpected frontmatter shape", path)
	}
	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("%s: frontmatter is not a mapping", path)
	}

	setPreserveKey(mapping, preserve)

	var yamlOut bytes.Buffer
	enc := yaml.NewEncoder(&yamlOut)
	enc.SetIndent(2)
	if err := enc.Encode(&root); err != nil {
		return fmt.Errorf("%s: encode frontmatter: %w", path, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("%s: close yaml encoder: %w", path, err)
	}

	var out bytes.Buffer
	out.WriteString(frontmatterDelimiter + "\n")
	out.Write(yamlOut.Bytes())
	out.WriteString(frontmatterDelimiter + "\n")
	out.Write(data[bodyStart:])

	fi, statErr := os.Stat(path)
	mode := os.FileMode(0o644)
	if statErr == nil {
		mode = fi.Mode().Perm()
	}
	return os.WriteFile(path, out.Bytes(), mode)
}

// frontmatterBounds locates the YAML block inside a SKILL.md-shaped file.
// Returns offsets into data: [yamlStart, yamlEnd) is the YAML body (between
// the two `---` delimiters, excluding them), bodyStart is the first byte
// after the trailing `---\n` (where the markdown body begins).
func frontmatterBounds(data []byte) (yamlStart, yamlEnd, bodyStart int, err error) {
	// Strip BOM + leading whitespace for the delimiter search, but we need
	// to return offsets into the original data so the markdown body stays
	// byte-exact.
	stripped := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	prefixSkipped := len(data) - len(stripped)
	ws := len(stripped) - len(bytes.TrimLeft(stripped, " \t\r\n"))
	cursor := prefixSkipped + ws
	if !bytes.HasPrefix(data[cursor:], []byte(frontmatterDelimiter)) {
		return 0, 0, 0, errors.New("missing leading '---' frontmatter delimiter")
	}
	cursor += len(frontmatterDelimiter)
	// Accept optional trailing whitespace on the opening line.
	for cursor < len(data) && (data[cursor] == ' ' || data[cursor] == '\t' || data[cursor] == '\r') {
		cursor++
	}
	if cursor >= len(data) || data[cursor] != '\n' {
		return 0, 0, 0, errors.New("opening '---' must be followed by a newline")
	}
	cursor++
	yamlStart = cursor

	// Find the closing delimiter on its own line.
	for cursor < len(data) {
		lineEnd := bytes.IndexByte(data[cursor:], '\n')
		var line []byte
		if lineEnd < 0 {
			line = data[cursor:]
		} else {
			line = data[cursor : cursor+lineEnd]
		}
		if bytes.Equal(bytes.TrimRight(line, " \t\r"), []byte(frontmatterDelimiter)) {
			yamlEnd = cursor
			if lineEnd < 0 {
				bodyStart = len(data)
			} else {
				bodyStart = cursor + lineEnd + 1
			}
			return yamlStart, yamlEnd, bodyStart, nil
		}
		if lineEnd < 0 {
			break
		}
		cursor += lineEnd + 1
	}
	return 0, 0, 0, errors.New("missing closing '---' frontmatter delimiter")
}

// setPreserveKey replaces or inserts the `preserve` key in a mapping node.
// Unknown keys stay intact and in their original positions; the mapping keeps
// its existing order (new key appended at the end).
func setPreserveKey(mapping *yaml.Node, preserve []string) {
	value := preserveSeqNode(preserve)
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == "preserve" {
			mapping.Content[i+1] = value
			return
		}
	}
	key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "preserve"}
	mapping.Content = append(mapping.Content, key, value)
}

// preserveSeqNode builds a YAML sequence node for a preserve list. An empty
// list becomes `[]` (flow style) so the key stays visible and machine-readable
// in the rewritten SKILL.md.
func preserveSeqNode(preserve []string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	if len(preserve) == 0 {
		seq.Style = yaml.FlowStyle
		return seq
	}
	for _, s := range preserve {
		seq.Content = append(seq.Content, &yaml.Node{
			Kind: yaml.ScalarNode, Tag: "!!str", Value: s,
		})
	}
	return seq
}
