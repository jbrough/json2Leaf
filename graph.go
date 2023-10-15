package json2Leaf

import (
	"fmt"
	"sort"
	"strings"

	gv "github.com/awalterschulze/gographviz"
)

func NewGraph(name string) (g *Graph, err error) {
	gr := gv.NewGraph()

	if err = gr.SetName(name); err != nil {
		return
	}

	if err = gr.SetDir(true); err != nil {
		return
	}

	g = &Graph{name, gr}

	return
}

type Graph struct {
	name  string
	graph *gv.Graph
}

func (g Graph) AddSubGraph(name string, ls []Leaf) (err error) {
	if err = g.graph.AddSubGraph(g.name, name, nil); err != nil {
		return
	}

	nodes := newNodes(ls)

	for _, n := range nodes {
		if err = g.graph.AddNode(name, n.fullName, n.attrs()); err != nil {
			return
		}
	}

	for _, n := range nodes {
		if n.parent != "" {
			if err = g.graph.AddEdge(n.parent, n.fullName, true, nil); err != nil {
				return
			}
		}
	}

	return
}

func (g Graph) String() string {
	return g.graph.String()
}

func newNodes(ls []Leaf) (r []node) {
	var items []Leaf
	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}

		items = append(items, l)
	}

	nodes := make(map[string]node)
	names := make(map[string]string)

	for _, item := range items {
		names[item.ID] = item.Name
	}

	for _, item := range items {
		n, ok := nodes[item.Name]
		if !ok {
			n = newNode(item.Name, names[item.ParentID])
		}
		n.addAttr(item.Path, item.DataType)
		nodes[item.Name] = n
	}

	for _, v := range nodes {
		r = append(r, v)
	}

	return
}

func newNode(name, parent string) node {
	return node{
		fullName:   name,
		parent:     parent,
		attributes: make(map[string]string),
	}
}

type node struct {
	fullName   string
	attributes map[string]string
	parent     string
}

func (n *node) addAttr(fullName, dataType string) {
	n.attributes[fullName] = dataType
}

func (n *node) leaf(s string) string {
	a := strings.Split(s, "__")

	return a[len(a)-1]
}

func (n *node) label() string {
	var keys []string
	for k := range n.attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	s := fmt.Sprintf("{%s|", n.leaf(n.fullName))
	for _, k := range keys {
		s += fmt.Sprintf(`+ %s : %s\l`, k, n.attributes[k])
	}
	s += "}"

	return fmt.Sprintf(`"%s"`, s)
}

func (n *node) attrs() map[string]string {
	r := make(map[string]string)
	r["label"] = n.label()
	r["shape"] = "record"
	return r
}
