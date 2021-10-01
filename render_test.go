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

package gemini

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/tdemin/gmnhg/internal/gmnhg"
)

var fileList []string

var (
	mdFilenameRegex = regexp.MustCompile(`^(.+)\.md$`)
)

func TestMain(m *testing.M) {
	// go test implicitly sets cwd to tested package directory; sadly,
	// this fact is undocumented
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		panic(err)
	}
	for _, fileInfo := range files {
		if match := mdFilenameRegex.FindStringSubmatch(fileInfo.Name()); !fileInfo.IsDir() && match != nil {
			fileList = append(fileList, match[1])
		}
	}
	os.Exit(m.Run())
}

func TestRenderer(t *testing.T) {
	for _, testName := range fileList {
		t.Logf("testing %s", testName)
		mdContents, err := ioutil.ReadFile(path.Join("testdata", testName+".md"))
		if err != nil {
			t.Fatalf("failed to open Markdown test %s: %v", testName, err)
		}
		gmiContents, err := ioutil.ReadFile(path.Join("testdata", testName+".gmi"))
		if err != nil {
			t.Logf("%s: cannot open Gemtext file, skipping: %v", testName, err)
			continue
		}
		content, _ := gmnhg.ParseMetadata(mdContents)
		geminiContent, err := RenderMarkdown(content, Defaults)
		if err != nil {
			t.Errorf("failed to convert %s Markdown to Gemtext: %v", testName, err)
		}
		if !bytes.Equal(geminiContent, gmiContents) {
			diff := myers.ComputeEdits(span.URIFromPath("a.gmi"),
				string(geminiContent), string(gmiContents))
			t.Errorf("content mismatch on %s, diff:\n%s", testName,
				gotextdiff.ToUnified("a.gmi", "b.gmi", string(geminiContent), diff))
		}
	}
}
