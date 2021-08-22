package gmnhg

import gemini "github.com/tdemin/gmnhg"

type Post struct {
	Post     []byte
	Metadata gemini.HugoMetadata
	Link     string
}

// Posts implements sort.Interface.
type Posts []Post

func (p Posts) Len() int {
	return len(p)
}

func (p Posts) Less(i, j int) bool {
	return p[i].Metadata.PostDate.Before(p[j].Metadata.PostDate)
}

func (p Posts) Swap(i, j int) {
	t := p[i]
	p[i] = p[j]
	p[j] = t
}
