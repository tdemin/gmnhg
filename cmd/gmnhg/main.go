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

// gmnhg converts Hugo content files to a Gemini site. This program is
// to be started in the top level directory of a Hugo site (the one
// containing config.toml).
//
// gmngh will read layout template files (with .gotmpl extension) and
// then apply them to content files ending with .md by the following
// algorithm (layout file names are relative to gmnhg/):
//
// 1. If the .md file specifies its own layout, the relevant layout file
// is applied. If not, the default template is applied (single). If the
// layout file does not exist, the file is skipped. Draft posts are not
// rendered. _index.md files are also skipped.
//
// 2. For every top-level content directory an index.gmi is generated,
// the corresponding template is taken from top/{directory_name}.gotmpl.
// Its content is taken from _index.gmi.md in that dir. If there's no
// matching template or no _index.gmi.md, the index won't be rendered.
//
// 3. The very top index.gmi is generated from index.gotmpl and
// top-level _index.gmi.
//
// The program will then copy static files from static/ directory to the
// output dir.
//
// Templates are passed the following data:
//
// 1. Single pages are given .Post, which contains the entire post
// rendered, .Metadata, which contains the metadata crawled from it (see
// HugoMetadata), and .Link, which contains the filename relative to
// content dir (with .md replaced with .gmi).
//
// 2. Directory index pages are passed .Posts, which is a slice over
// post metadata crawled (see HugoMetadata), .Dirname, which is
// directory name relative to content dir, and .Content, which is
// rendered from directory's _index.gmi.md.
//
// 3. The top-level index.gmi is passed with the .PostData map whose
// keys are top-level content directories names and values are slices
// over the same post props as specified in 1, and .Content, which is
// rendered from top-level _index.gmi.md.
//
// This program provides some extra template functions, documented in
// templates.go.
//
// One might want to ignore _index.gmi.md files with the following Hugo
// config option in config.toml:
//
//  ignoreFiles = [ "_index\\.gmi\\.md$" ]
//
// Limitations:
//
// * For now, the program will only recognize YAML front matter, while
// Hugo supports it in TOML, YAML, JSON, and org-mode formats.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	gemini "github.com/tdemin/gmnhg"
)

const (
	defaultPageTemplate = "single"
	indexMdFilename     = "_index.gmi.md"
	indexFilename       = "index.gmi"
)

const (
	contentBase  = "content/"
	templateBase = "gmnhg/"
	staticBase   = "static/"
	outputBase   = "output/"
)

var (
	tmplNameRegex    = regexp.MustCompile("^" + templateBase + `([\w-_ /]+)\.gotmpl$`)
	leafIndexRegex   = regexp.MustCompile("^" + contentBase + `([\w-_ /]+)/index\.[\w]+$`)
	pagePathRegex    = regexp.MustCompile("^" + contentBase + `([\w-_ /]+)/([\w-_ ]+)\.md$`)
)

var hugoConfigFiles = []string{"config.toml", "config.yaml", "config.json"}

type post struct {
	Post     []byte
	Metadata gemini.HugoMetadata
	Link     string
}

func copyFile(dst, src string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()
	if p := path.Dir(dst); p != "" {
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return nil
}

func writeFile(dst string, contents []byte) error {
	if p := path.Dir(dst); p != "" {
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()
	if _, err := output.Write(contents); err != nil {
		return err
	}
	return nil
}

func hasSubPath(paths []string, path string) bool {
  for _, p := range paths {
		if strings.HasPrefix(path, p + "/") {
			return true
		}
  }
  return false
}

var version = "v0+HEAD"

func main() {
	var (
		outputDir, workingDir string
		isVersionCmd          bool
	)
	flag.StringVar(&outputDir, "output", outputBase, "output directory (will be created if missing)")
	flag.StringVar(&workingDir, "working", "", "working directory (defaults to current directory)")
	flag.BoolVar(&isVersionCmd, "version", false, "display version")
	flag.Parse()

	if isVersionCmd {
		println("gmnhg", version)
		return
	}

	if workingDir != "" {
		if err := os.Chdir(workingDir); err != nil {
			panic(err)
		}
	}

	configFound := false
	for _, filename := range hugoConfigFiles {
		if fileInfo, err := os.Stat(filename); !(os.IsNotExist(err) || fileInfo.IsDir()) {
			configFound = true
			break
		}
	}
	if !configFound {
		panic(fmt.Errorf("no Hugo config in %v found; not in a Hugo site dir?", hugoConfigFiles))
	}

	// build templates
	templates := make(map[string]*template.Template)
	if _, err := os.Stat(templateBase); !os.IsNotExist(err) {
		if err := filepath.Walk(templateBase, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := tmplNameRegex.FindStringSubmatch(path)
			if name == nil || len(name) != 2 {
				return nil
			}
			tmplName := name[1]
			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			tmpl, err := template.New(tmplName).Funcs(defineFuncMap()).Parse(string(contents))
			if err != nil {
				return err
			}
			templates[tmplName] = tmpl
			return nil
		}); err != nil {
			panic(err)
		}
	}

	// collect leaf node paths (directories containing an index.* file)
	leafIndexPaths := []string{}
	if err := filepath.Walk(contentBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matches := leafIndexRegex.FindStringSubmatch(path); matches != nil {
			leafIndexPaths = append(leafIndexPaths, contentBase + matches[1])
		}
		return nil
	}); err != nil {
		panic(err)
	}

	// render posts to Gemtext and collect top level posts data
	posts := make(map[string]*post)
	topLevelPosts := make(map[string][]*post)
	if err := filepath.Walk(contentBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if n := info.Name(); info.IsDir() || !strings.HasSuffix(n, ".md") || n == "_index.md" || n == indexMdFilename {
			return nil
		}
		// do not render resources under leaf paths unless they are an index
		if info.Name() != "index.md" && hasSubPath(leafIndexPaths, path) {
			return nil
		}
		fileContent, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		gemText, metadata, err := gemini.RenderMarkdown(fileContent, gemini.Defaults)
		// skip drafts from rendering
		if errors.Is(err, gemini.ErrPostIsDraft) {
			return nil
		} else if err != nil {
			return err
		}
		key := strings.TrimPrefix(strings.TrimSuffix(path, ".md"), contentBase) + ".gmi"
		p := post{
			Post:     gemText,
			Link:     key,
			Metadata: metadata,
		}
		posts[key] = &p
		if matches := pagePathRegex.FindStringSubmatch(path); matches != nil {
			dirs := strings.Split(matches[1], "/")
			// if this is a leaf node, the dir is the name of the post
			if info.Name() == "index.md" {
				dirs = dirs[:len(dirs)-1]
			}
			// include post in all subdirectory indices
			for i, dir := range dirs {
				if i > 0 {
					dirs[i] = dirs[i - 1] + "/" + dir
				}
			}
			for _, dir := range dirs {
				topLevelPosts[dir] = append(topLevelPosts[dir], &p)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}

	// clean up output dir beforehand
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			panic(err)
		}
	} else {
		dir, err := ioutil.ReadDir(outputDir)
		if err != nil {
			panic(err)
		}
		for _, d := range dir {
			os.RemoveAll(path.Join(outputDir, d.Name()))
		}
	}

	var singleTemplate = defaultSingleTemplate
	if tmpl, hasTmpl := templates[defaultPageTemplate]; hasTmpl {
		singleTemplate = tmpl
	}

	// render posts to files
	for fileName, post := range posts {
		var tmpl = singleTemplate
		if pl := post.Metadata.PostLayout; pl != "" {
			t, ok := templates[pl]
			if !ok {
				// no point trying to render pages with no layout
				continue
			}
			tmpl = t
		}
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, &post); err != nil {
			panic(err)
		}
		if err := writeFile(path.Join(outputDir, fileName), buf.Bytes()); err != nil {
			panic(err)
		}
	}
	// render indexes for top-level dirs
	for dirname, posts := range topLevelPosts {
		tmpl, hasTmpl := templates["top/"+dirname]
		if !hasTmpl {
			continue
		}
		content, err := ioutil.ReadFile(path.Join(contentBase, dirname, indexMdFilename))
		if err != nil {
			// skip unreadable index files
			continue
		}
		gemtext, _, err := gemini.RenderMarkdown(content, gemini.Defaults)
		if errors.Is(err, gemini.ErrPostIsDraft) {
			continue
		} else if err != nil {
			panic(err)
		}
		cnt := map[string]interface{}{
			"Posts":   posts,
			"Dirname": dirname,
			"Content": gemtext,
		}
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, cnt); err != nil {
			panic(err)
		}
		if err := writeFile(path.Join(outputDir, dirname, indexFilename), buf.Bytes()); err != nil {
			panic(err)
		}
	}
	// render index page
	var indexTmpl = defaultIndexTemplate
	if t, hasIndexTmpl := templates["index"]; hasIndexTmpl {
		indexTmpl = t
	}
	indexContent, err := ioutil.ReadFile(path.Join(contentBase, indexMdFilename))
	if err != nil {
		panic(err)
	}
	gemtext, _, err := gemini.RenderMarkdown(indexContent, gemini.Defaults)
	if err != nil && !errors.Is(err, gemini.ErrPostIsDraft) {
		panic(err)
	}
	buf := bytes.Buffer{}
	cnt := map[string]interface{}{"PostData": topLevelPosts, "Content": gemtext}
	if err := indexTmpl.Execute(&buf, cnt); err != nil {
		panic(err)
	}
	if err := writeFile(path.Join(outputDir, indexFilename), buf.Bytes()); err != nil {
		panic(err)
	}

	// copy static files to output dir unmodified
	if err := filepath.Walk(staticBase, func(p string, info os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return copyFile(path.Join(outputDir, strings.TrimPrefix(p, staticBase)), p)
	}); err != nil {
		panic(err)
	}
}
