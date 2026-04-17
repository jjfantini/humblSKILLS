// Package frontmatter parses SKILL.md YAML frontmatter and the humblSKILLS
// extension keys layered on top of the agentskills.io base format.
package frontmatter

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

const delimiter = "---"

// Frontmatter is the typed view of a SKILL.md frontmatter block. Unknown keys
// in the YAML are accepted but not surfaced; humblSKILLS only validates the
// fields it owns.
type Frontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Version     string   `yaml:"version"`
	Requires    []string `yaml:"requires,omitempty"`
	Platforms   []string `yaml:"platforms,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Preserve    []string `yaml:"preserve,omitempty"`
}

// Parse splits the leading `---` YAML frontmatter block from the body and
// unmarshals it. Returns the parsed frontmatter, the remaining body bytes, and
// any parse error.
func Parse(data []byte) (Frontmatter, []byte, error) {
	var fm Frontmatter

	rest := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // strip UTF-8 BOM
	rest = bytes.TrimLeft(rest, " \t\r\n")

	if !bytes.HasPrefix(rest, []byte(delimiter)) {
		return fm, nil, errors.New("missing leading '---' frontmatter delimiter")
	}
	rest = rest[len(delimiter):]
	rest = bytes.TrimLeft(rest, " \t\r")
	if len(rest) == 0 || rest[0] != '\n' {
		return fm, nil, errors.New("opening '---' must be followed by a newline")
	}
	rest = rest[1:]

	end := findClosingDelimiter(rest)
	if end < 0 {
		return fm, nil, errors.New("missing closing '---' frontmatter delimiter")
	}

	yamlBlock := rest[:end]
	body := rest[end+len(delimiter):]
	body = bytes.TrimLeft(body, " \t\r\n")

	if err := yaml.Unmarshal(yamlBlock, &fm); err != nil {
		return fm, nil, fmt.Errorf("yaml parse: %w", err)
	}
	return fm, body, nil
}

// findClosingDelimiter returns the byte offset of a line containing exactly
// "---" (possibly with trailing whitespace), or -1 if not found.
func findClosingDelimiter(data []byte) int {
	idx := 0
	for idx < len(data) {
		lineEnd := bytes.IndexByte(data[idx:], '\n')
		var line []byte
		if lineEnd < 0 {
			line = data[idx:]
		} else {
			line = data[idx : idx+lineEnd]
		}
		if bytes.Equal(bytes.TrimRight(line, " \t\r"), []byte(delimiter)) {
			return idx
		}
		if lineEnd < 0 {
			return -1
		}
		idx += lineEnd + 1
	}
	return -1
}
