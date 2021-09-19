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

// md2gmn converts Markdown text files to text/gemini.
package main

import (
	"flag"
	"io/ioutil"
	"os"

	gemini "github.com/tdemin/gmnhg"
	"github.com/tdemin/gmnhg/internal/gmnhg"
)

var version = "v0+HEAD"

func main() {
	var (
		input        string
		file         *os.File
		isVersionCmd bool
	)
	flag.StringVar(&input, "f", "", "input file")
	flag.BoolVar(&isVersionCmd, "version", false, "display version")
	flag.Parse()

	if isVersionCmd {
		println("md2gmn", version)
		return
	}

	if input != "" {
		var err error
		file, err = os.Open(input)
		if err != nil {
			panic(err)
		}
	} else {
		file = os.Stdin
	}
	text, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	content, _ := gmnhg.ParseMetadata(text)
	geminiContent, err := gemini.RenderMarkdown(content, gemini.Defaults)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(geminiContent)
}
