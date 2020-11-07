// gmnhg converts Hugo posts to gemini content.
//
// TODO: it is yet to actually do that.
package main

import (
	"fmt"

	"git.tdem.in/tdemin/gmnhg/internal/gemini"
	"github.com/davecgh/go-spew/spew"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

var text = `
# Some document

This is some markdown [text](https://tdem.in). This is some more text.

![This is some image](https://tdem.in/favicon.ico)

[This is some full-blown link.](https://tdem.in/nyaa)

This is some more plain text. More of it!

+ Unordered list item
+ Another list item
	* Indented list item.
	* Another one.
+ Third.

1. Ordered list item.
2. Another one.
	* and another inset list.
	* text.
3. Yay.

## Subheading 2

More text!

> Some weird blockquote. More text.
> More quote text.
`

func main() {
	ast := markdown.Parse([]byte(text), parser.NewWithExtensions(parser.CommonExtensions))
	spew.Dump(ast)
	geminiContent := markdown.Render(ast, gemini.NewRenderer())
	fmt.Printf("---\noriginal:\n---\n%s---\ngemini:\n---\n%s", text, geminiContent)
}
