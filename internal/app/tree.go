package app

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// tnode is a file or folder in a drawn tree.
type tnode struct {
	name string
	note string
	kids map[string]*tnode
}

// add puts a slash-separated path into the tree and returns its last
// part. Names are compared without case, as on Windows.
func (n *tnode) add(rel string) *tnode {
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." {
			continue
		}
		if n.kids == nil {
			n.kids = map[string]*tnode{}
		}
		k := strings.ToLower(part)
		if n.kids[k] == nil {
			n.kids[k] = &tnode{name: part}
		}
		n = n.kids[k]
	}
	return n
}

func (n *tnode) sorted() []*tnode {
	out := make([]*tnode, 0, len(n.kids))
	for _, k := range n.kids {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].name) < strings.ToLower(out[j].name) })
	return out
}

// lines draws the tree like tree(1): one entry per line, notes after it.
func (n *tnode) lines(others int) [][2]string {
	var out [][2]string
	var walk func(n *tnode, indent string)
	walk = func(n *tnode, indent string) {
		kids := n.sorted()
		for i, k := range kids {
			last := i == len(kids)-1 && (indent != "" || others == 0)
			branch, more := "├── ", "│   "
			if last {
				branch, more = "└── ", "    "
			}
			out = append(out, [2]string{indent + branch + k.name, k.note})
			walk(k, indent+more)
		}
	}
	walk(n, "")
	if others > 0 {
		out = append(out, [2]string{fmt.Sprintf("└── … %d more, unchanged", others), ""})
	}
	return out
}

func drawTree(b *strings.Builder, title string, n *tnode, others int) {
	ls := n.lines(others)
	width := 0
	for _, l := range ls {
		if l[1] != "" {
			width = max(width, utf8.RuneCountInString(l[0]))
		}
	}
	width = min(width, 60)
	b.WriteString(title + "\n")
	for _, l := range ls {
		if l[1] == "" {
			b.WriteString(l[0] + "\n")
			continue
		}
		pad := max(width-utf8.RuneCountInString(l[0]), 0)
		b.WriteString(l[0] + strings.Repeat(" ", pad) + "   " + l[1] + "\n")
	}
}

// tree draws the parts of the real folder that the request touches, as
// they are now and as they will be after the commit.
func (q *request) tree() string {
	before, after := &tnode{}, &tnode{}
	for _, c := range q.Changes {
		switch c.kind {
		case kReplace:
			before.add(c.rel)
			after.add(c.rel).note = "~ replaced"
		case kAdd:
			parent := c.top
			if parent == "" {
				parent = c.rel
			}
			if d := path.Dir(parent); d != "." {
				before.add(d)
			}
			if c.top != "" {
				after.add(c.top).note = "+ new folder"
			}
			note := "+ new"
			if l := originLabel(c.origin); l != "" {
				note += ", " + l
			}
			after.add(c.rel).note = note
		case kDiffers:
			before.add(c.rel)
			after.add(c.rel).note = "≠ unchanged; the working copy's version differs and is not copied"
		case kDup:
			before.add(c.dup)
			after.add(c.dup).note = fmt.Sprintf("= unchanged; %q of the working copy is the same company, not copied", path.Base(c.top))
		}
	}
	var b strings.Builder
	drawTree(&b, "Before: "+filepath.Clean(q.To), before, q.Others)
	b.WriteString("\n")
	drawTree(&b, "After the commit:", after, q.Others)
	return b.String()
}
