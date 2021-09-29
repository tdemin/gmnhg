# Hugo-to-Gemini converter

[![PkgGoDev](https://pkg.go.dev/badge/github.com/tdemin/gmnhg)](https://pkg.go.dev/github.com/tdemin/gmnhg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tdemin/gmnhg)](https://goreportcard.com/report/github.com/tdemin/gmnhg)
[![Push to GHCR](https://github.com/tdemin/gmnhg/actions/workflows/docker.yml/badge.svg)](https://github.com/tdemin/gmnhg/actions/workflows/docker.yml)

This repo holds a converter of Hugo Markdown posts to
[text/gemini][Gemtext] (also named Gemtext in this README). The
converter is supposed to make people using [Hugo](https://gohugo.io)'s
entrance to [Project Gemini][Gemini], the alternate web, somewhat
simpler.

[Gemini]: https://gemini.circumlunar.space
[Gemtext]: https://gemini.circumlunar.space/docs/specification.html

The renderer uses the [gomarkdown][gomarkdown] library for parsing
Markdown. gomarkdown has a few quirks at this time, the most notable one
being unable to parse links/images inside other links.

At this time, gmnhg can convert these Markdown elements to Gemtext:

* paragraphs, converting them to soft wrap as per Gemini spec p. 5.4.1;
* inline text formatting (bold, emphasis, strikethrough, code,
  subscript, superscript), which stays in the text to preserve stylistic
  context;
* headings;
* blockquotes;
* preformatted blocks;
* tables, displayed as ASCII preformatted blocks;
* lists (as Gemini doesn't allow lists of level >= 2, those will be
  reflected with an extra indentation level): ordered, numbered,
  definition;
* links and images, rendered as Gemtext links (inline links are rendered
  after their parent paragraph or other block element in a links block
  sorted by element type);
* footnotes, rendered as paragraphs;
* horizontal rules.

The renderer will also treat lists of links and paragraphs consisting of
links only the special way: it will render only the links block for
them.

To get a better idea of how source Markdown looks like after the
conversion to Gemtext, see [testdata](testdata) directory.

[gomarkdown]: https://github.com/gomarkdown/markdown

## gmnhg

This program converts Hugo Markdown content files from `content/` in
accordance with templates found in `gmnhg/` to the output dir. It
also copies static files from `static/` to the output dir.

For more details about the rendering process, see the
[doc](cmd/gmnhg/main.go) attached to the program.

```
Usage of gmnhg:
  -output string
        output directory (will be created if missing) (default "output/")
  -working string
        working directory (defaults to current directory)
```

## md2gmn

This program reads Markdown input from either text file (if `-f
filename` is given), or stdin. The resulting Gemtext goes to stdout.

```
Usage of md2gmn:
  -f string
        input file
```

md2gmn is mainly made to facilitate testing the Gemtext renderer but
can be used as a standalone program as well.

## Site configuration

gmnhg will pick up some attributes such as site title, base URL, and
language code from your Hugo configuration file (`config.toml`,
`config.yaml`, or `config.json`). Presently these are used in the
default RSS template.

gmnhg provides a way to override these attributes by defining a
`gmnhg` section in the configuration file and nesting the attributes
to override underneath this section. Presently you can override both
`baseUrl` and `title` in this manner.

For example, you could add the following to your `config.toml` to
override your `baseUrl`:

```
[gmnhg]
baseUrl = "gemini://mysite.com"
```

This is recommended, as it will ensure that RSS links on your Gemini
site use the correct URL.

## License

This program is redistributed under the terms and conditions of the GNU
General Public License, more specifically version 3 of the License. For
details, see [COPYING](COPYING).
