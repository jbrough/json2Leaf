package json2Leaf

import (
	"strings"
)

func NewReplacer(data [][]string) (t *Replacer) {
	d := make(map[string]string)

	for _, strs := range data {
		s := strings.ToLower(strs[1])
		d[s] = strs[0]
	}

	return &Replacer{d}
}

type Replacer struct {
	data map[string]string
}

func (t Replacer) Do(ls []Leaf) (r []Leaf) {
	for _, l := range ls {
		l.Name = t.get(l.Name)
		l.Path = t.get(l.Path)
		switch l.Value.(type) {
		case string:
			l.Value = t.get(l.Value.(string))
		}
		r = append(r, l)
	}

	return
}

func (t Replacer) get(str string) string {
	lower := strings.ToLower(str)
	res, ok := t.data[lower]
	if ok {
		return res
	}

	return str
}
