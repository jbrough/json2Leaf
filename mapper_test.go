package json2Leaf_test

import (
	"fmt"
	"sort"
	"testing"

	j "github.com/jbrough/json2Leaf"
	"github.com/stretchr/testify/assert"
)

func TestDoPathWithTableNameConfig(t *testing.T) {
	b := []byte(`
	{
		"foo": {
			"bar": [
				{
					"ObjectA": {
						"SubObject": { 
							"foo": {
								"bar": "a"
							}
						}
					},
					"ObjectB": {
						"SubObject": {
							"foo": {
								"bar": "b"
							}
						}
					}
				}
			]
		}
	}`)

	c := j.NewConfig()

	ls, err := j.NewMapper(c).DoPath("test", "foo.bar", b)
	if err != nil {
		t.Error(err)
	}

	sc := lander.NewSchema()
	for _, l := range ls {
		sc.Add(l)
	}

	var names []string
	var paths []string

	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}
		fmt.Printf("%+v\n", l)

		names = append(names, l.Name)
		paths = append(paths, l.Path)
	}

	sort.Strings(names)
	sort.Strings(paths)

	assert.Equal(t,
		[]string{"test", "test"},
		names,
	)

	assert.Equal(t,
		[]string{"object_a__sub_object__foo__bar", "object_b__sub_object__foo__bar"},
		paths,
	)

	c = j.Config{
		TableNames: []string{
			"ObjectA__SubObject",
			"ObjectB__SubObject",
		},
	}

	ls, err = j.NewMapper(c).DoPath("test", "foo.bar", b)
	if err != nil {
		t.Error(err)
	}

	sc = lander.NewSchema()
	for _, l := range ls {
		sc.Add(l)
	}

	names = []string{}
	paths = []string{}

	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}

		fmt.Printf("%+v\n", l)
		names = append(names, l.Name)
		paths = append(paths, l.Path)
	}

	sort.Strings(names)
	sort.Strings(paths)

	assert.Equal(t,
		[]string{"object_a__sub_object", "object_b__sub_object"},
		names,
	)

	assert.Equal(t,
		[]string{"foo__bar", "foo__bar"},
		paths,
	)
}

func TestMapJSON(t *testing.T) {
	b := []byte(`
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
				"foo2": 1
			}
		],
		"qux": [
			"a", 1, false
		]
	}`)

	c := j.NewConfig()

	ls, err := j.NewMapper(c).Do("doc", b)
	if err != nil {
		t.Error(err)
	}

	for _, l := range ls {
		fmt.Printf("%+v\n", l)
	}

	assert.Equal(t, 13, len(ls))

	checkTree := func(id, parent, name string) {
		var ok bool
		for _, v := range ls {
			if v.ID == id && v.ParentID == parent && v.Name == "_tree" {
				ok = true
				assert.Equal(t, name, v.Value)
			}
		}
		if !ok {
			t.Errorf("ID/ParentID %s/%s not found in _tree results", id, parent)
		}

	}

	checkAt := func(dataType, name, path string, value interface{}) {
		var ok bool
		for _, v := range ls {
			if v.Name == name && v.Path == path {
				ok = true
				assert.Equal(t, dataType, v.DataType)
				assert.Equal(t, value, v.Value)
				checkTree(v.ID, v.ParentID, v.Name)
			}
		}
		if !ok {
			t.Errorf("Name/Path %s/%s not found in results", name, path)
		}
	}

	checkIn := func(dataType, name string, value interface{}) {
		var ok bool
		for _, v := range ls {
			if v.Name == name && v.Value == value {
				ok = true
				assert.Equal(t, dataType, v.DataType)
				checkTree(v.ID, v.ParentID, v.Name)
			}
		}
		if !ok {
			t.Errorf("Name/val %s/%v not found in results", name, value)
		}
	}

	checkAt("string", "doc", "foo", "foo_v")
	checkAt("string", "doc", "bar__baz", "baz_v")
	checkAt("float64", "doc__baz", "foo2", 1.0)
	checkAt("float64", "doc__baz", "foo1", 1.3)
	checkIn("string", "doc__qux", "a")
	checkIn("float64", "doc__qux", 1.0)
	checkIn("bool", "doc__qux", false)
}

func TestBug(t *testing.T) {
	t.Skip()
	test1 := []byte(`
		{
			"flat": "one",
			"foo": [
				{
					"bar": {
						"baz": "qux"
					}
				},
				{
					"bar2": {
						"baz2": "qux2",
						"arrX": [
							{"x1": 1},
							{"x1": 12, "x2": 2, "x2A": 3}
						],
						"arr": [
							{ 
								"arr3": [3]
							},
							{
								"arr2": [1,2]
							}
						]
					}
				}
			]
		}
	`)

	c := j.NewConfig()

	ls, err := j.NewMapper(c).Do("test1", test1)
	if err != nil {
		t.Error(err)
	}

	// TODO shouldn't really test this via another package, but don't have time
	// to write detailed expectations. The functionality is already tested in
	// this file and better to have this now as a regression test than to not.
	sc := lander.NewSchema()
	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}

		sc.Add(l)
	}

	csv, err := sc.CSV()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t,
		"table,column,bool,int,real,text\ntest1__foo__bar2__arr__arr2,val,false,true,false,false\ntest1,flat,false,false,false,true\ntest1__foo,bar2__baz2,false,false,false,true\ntest1__foo,bar__baz,false,false,false,true\ntest1__foo__bar2__arrX,x2A,false,true,false,false\ntest1__foo__bar2__arrX,x1,false,true,false,false\ntest1__foo__bar2__arrX,x2,false,true,false,false\ntest1__foo__bar2__arr__arr3,val,false,true,false,false\n",
		csv,
	)
}

func TestColumnOverrideConfig(t *testing.T) {

	b := []byte(
		`{"foo": {"bar": 1}}`,
	)

	c := j.NewConfig()
	ls, err := j.NewMapper(c).Do("test", b)
	if err != nil {
		t.Error(err)
	}

	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}

		assert.Equal(t,
			"test",
			l.Name,
		)

		assert.Equal(t,
			"foo__bar",
			l.Path,
		)

		break
	}

	ls, err = j.NewMapper(j.Config{
		ColumnOverrides: [][]string{[]string{"test", "foo__bar", "a", "b"}},
	}).Do("test", b)
	if err != nil {
		t.Error(err)
	}

	for _, l := range ls {
		if l.Name == "_tree" {
			continue
		}

		assert.Equal(t,
			"a",
			l.Name,
		)

		assert.Equal(t,
			"b",
			l.Path,
		)

		break
	}

}
