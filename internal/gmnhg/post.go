// This file is part of gmnhg.

// gmnhg is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gmnhg is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gmnhg. If not, see <https://www.gnu.org/licenses/>.

package gmnhg

import (
	"bytes"
	"encoding/json"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type Post struct {
	Post     []byte
	Metadata Metadata
	Link     string
}

// Posts implements sort.Interface.
type Posts []Post

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	return p[i].Metadata.Date.Before(p[j].Metadata.Date)
}

func (p Posts) Swap(i, j int) {
	t := p[i]
	p[i] = p[j]
	p[j] = t
}

// Metadata contains all recognized Hugo properties.
type Metadata struct {
	Title      string    `yaml:"title" toml:"title" json:"title" org:"title"`
	IsDraft    bool      `yaml:"draft" toml:"draft" json:"draft" org:"draft"`
	Layout     string    `yaml:"layout" toml:"layout" json:"layout" org:"layout"`
	Date       time.Time `yaml:"date" toml:"date" json:"date" org:"date"`
	Summary    string    `yaml:"summary" toml:"summary" json:"summary" org:"summary"`
	IsHeadless bool      `yaml:"headless" toml:"headless" json:"headless" org:"headless"`
}

var (
	yamlDelimiter   = []byte("---\n")
	tomlDelimiter   = []byte("+++\n")
	jsonObjectRegex = regexp.MustCompile(`\A(\{[\s\S]*\})\n\n`)
	orgModeRegex    = regexp.MustCompile(`\A((?:#\+\w+: ?\S*\n)*)`)
)

// ParseMetadata extracts TOML/JSON/YAML/org-mode format front matter
// from Markdown text. If no metadata is found, markdown will be equal
// to source.
//
// TOML front matter is identified as +++ symbols at the very start of
// the text, followed by TOML content, followed by another +++ (YAML is
// the same, but with ---). JSON front matter is identified as a JSON
// object followed by two newline symbols. org-mode front matter is
// identified as a set of #+KEY: VALUE lines, the first line started
// with anything else but #+ ends the front matter.
func ParseMetadata(source []byte) (markdown []byte, metadata Metadata) {
	var (
		blockEnd        int
		metadataContent []byte
	)
	markdown = source
	// block start is always 0, as front matter is only permitted at the
	// very start of the file
	if bytes.Index(source, yamlDelimiter) == 0 {
		blockEnd = bytes.Index(source[len(yamlDelimiter):], yamlDelimiter)
		if blockEnd == -1 {
			return
		}
		metadataContent = source[len(yamlDelimiter) : blockEnd+len(yamlDelimiter)]
		if err := yaml.Unmarshal(metadataContent, &metadata); err != nil {
			return
		}
		markdown = source[blockEnd+len(yamlDelimiter)*2:]
	} else if bytes.Index(source, tomlDelimiter) == 0 {
		blockEnd = bytes.Index(source[len(tomlDelimiter):], tomlDelimiter)
		if blockEnd == -1 {
			return
		}
		metadataContent = source[len(tomlDelimiter) : blockEnd+len(tomlDelimiter)]
		if err := toml.Unmarshal(metadataContent, &metadata); err != nil {
			return
		}
		markdown = source[blockEnd+len(yamlDelimiter)*2:]
	} else if match := jsonObjectRegex.FindIndex(source); match != nil {
		blockEnd = match[1]
		metadataContent = source[:blockEnd]
		if err := json.Unmarshal(metadataContent, &metadata); err != nil {
			return
		}
		markdown = source[blockEnd+1:] // JSON end + \n\n - 1
	} else if match := orgModeRegex.FindIndex(source); match != nil {
		blockEnd = match[1]
		metadataContent = source[:blockEnd]
		if err := unmarshalORG(metadataContent, &metadata); err != nil {
			return
		}
		markdown = source[blockEnd:]
	}
	return
}
