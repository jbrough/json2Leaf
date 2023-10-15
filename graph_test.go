package json2Leaf_test

import (
	"fmt"
	"strings"
	"testing"

	j "github.com/jbrough/json2Leaf"
)

func TestGraph(t *testing.T) {
	test1 := []byte(`
		{
			"foo": "foo_v",
			"bar": {
				"baz": "baz_v"
			},
			"baz": [
				{
					"foo1": 1.3
				},
				{
					"foo2": 1,
					"foo3": {
						"bar": [
							{
								"baz": "qux"
							}
						]
					}
				}
			],
			"qux": [
				"a", 1, false
			]
		}
	`)

	test2 := []byte(`
		{
			"baz": "123",
			"bar": "abc",
			"foo": [
				"a", 1, false
			]
		}
	`)

	g, err := j.NewGraph("test")
	if err != nil {
		t.Error(err)
	}

	c := j.Config{}
	ls, err := j.NewMapper(c).Do("test1", test1)
	if err != nil {
		t.Error(err)
	}

	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}
		fmt.Printf("%+v\n", l)
	}

	if err = g.AddSubGraph("test1", ls); err != nil {
		t.Error(err)
	}

	ls, err = j.NewMapper(c).Do("test2", test2)
	if err != nil {
		t.Error(err)
	}

	if err = g.AddSubGraph("test2", ls); err != nil {
		t.Error(err)
	}

	// g.String() outputs the DOT format I'd like to test, but ordering of
	// certain elements isn't deterministic which makes it difficult to test.
	// Attempting a sanity-check, at least.

	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}

		fmt.Printf("%+v\n", l)
	}

	actual := g.String()

	var tests = []struct {
		in string
	}{
		{`digraph test {`},
		{`test1->test1__baz;`},
		{`test1__baz->test1__baz__foo3__bar;`},
		{`test1->test1__qux;`},
		{`test2->test2__foo;`},
		{`subgraph test1 {`},
		{`baz__foo3__bar [ label="{bar|+ baz : string\l}", shape=record ];`},
		{`test1 [ label="{test1|+ bar__baz : string\l+ foo : string\l}", shape=record ];`},
		{`test1__baz [ label="{baz|+ foo1 : float64\l+ foo2 : float64\l}", shape=record ];`},
		{`test1__qux [ label="{qux|+ val : bool\l}", shape=record ];`},
		{`subgraph test2 {`},
		{`test2 [ label="{test2|+ bar : string\l+ baz : string\l}", shape=record ];`},
		{`test2__foo [ label="{foo|+ val : bool\l}", shape=record ];`},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if !strings.Contains(actual, tt.in) {
				t.Errorf("%s, should contain %s", actual, tt.in)
			}
		})
	}
}
