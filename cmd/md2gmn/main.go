// md2gmn converts Markdown text files to text/gemini.
package main

import (
	"flag"
	"io/ioutil"
	"os"

	gemini "git.tdem.in/tdemin/gmnhg"
)

func main() {
	var (
		input string
		file  *os.File
	)
	flag.StringVar(&input, "f", "", "input file")
	flag.Parse()

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

	os.Stdout.Write(gemini.RenderMarkdown(text))
}
