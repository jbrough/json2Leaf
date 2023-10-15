package json2Leaf

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Jeffail/gabs"
	"github.com/google/uuid"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	replacer := strings.NewReplacer("-", "", "#", "")
	str = replacer.Replace(str)
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	snake = strings.Replace(snake, "___", "__", -1)

	return strings.ToLower(snake)
}

func NewConfig() Config {
	return Config{}
}

type Config struct {
	ColumnOverrides [][]string
	ColumnSubs      [][]string
	TableNames      []string
	TableSubs       [][]string
}

type Leaf struct {
	DataType string
	Name     string
	ID       string
	ParentID string
	Path     string
	Value    interface{}
}

func NewMapper(c Config) *Mapper {
	overrides := make(map[string]map[string][]string)

	for _, o := range c.ColumnOverrides {
		if _, ok := overrides[o[0]]; !ok {
			overrides[o[0]] = make(map[string][]string)
		}

		m := overrides[o[0]]
		m[o[1]] = o[2:]
	}

	return &Mapper{
		config:    c,
		nodes:     make(map[string]interface{}),
		overrides: overrides,
	}
}

type Mapper struct {
	config     Config
	leaves     []Leaf
	nodes      map[string]interface{}
	nodesMutex sync.RWMutex
	overrides  map[string]map[string][]string
}

func (m *Mapper) DoPath(name, path string, b []byte) (ls []Leaf, err error) {
	d, err := gabs.ParseJSON(b)
	if err != nil {
		return
	}

	err = m.do(name, "", "", "", d.Path(path))

	return m.leaves, err
}

func (m *Mapper) Do(name string, b []byte) (ls []Leaf, err error) {
	d, err := gabs.ParseJSON(b)
	if err != nil {
		return
	}

	err = m.do(name, "", "", "", d)

	return m.leaves, err
}

func (m *Mapper) hasNode(node, parent string) (ok bool) {
	m.nodesMutex.RLock()
	defer m.nodesMutex.RUnlock()

	_, ok = m.nodes[node+parent]
	return
}

func (m *Mapper) addNode(node, parent string) {
	m.nodesMutex.Lock()
	defer m.nodesMutex.Unlock()

	m.nodes[node+parent] = struct{}{}
}

func (m *Mapper) add(name, path, node, parent string, d *gabs.Container) {
	replacer := strings.NewReplacer("-", "", "#", "")
	path = replacer.Replace(path)

	path = toSnakeCase(path)
	for _, strs := range m.config.ColumnSubs {
		path = strings.Replace(path, strs[0], strs[1], -1)
	}

	name = toSnakeCase(name)
	for _, strs := range m.config.TableSubs {
		name = strings.Replace(name, strs[0], strs[1], -1)
	}

	oldNode := node

	paths, ok := m.overrides[name]
	if ok {
		override, ok := paths[path]
		if ok {
			name = override[0]
			path = override[1]
			parent = oldNode
			node = uuid.New().String()
		}
	}

	m.leaves = append(m.leaves, Leaf{
		DataType: fmt.Sprintf("%T", d.Data()),
		Name:     name,
		ID:       node,
		ParentID: parent,
		Path:     path,
		Value:    d.Data(),
	})

	if !m.hasNode(node, parent) {
		m.leaves = append(m.leaves, Leaf{
			DataType: "string",
			ID:       node,
			Name:     "_tree",
			ParentID: parent,
			Path:     "name",
			Value:    name,
		})

		m.addNode(node, parent)
	}

}

func (m *Mapper) do(name, path, node, parent string, d *gabs.Container) error {
	if d.Data() == nil {
		return nil
	}

	for _, tn := range m.config.TableNames {
		if path == tn {
			prevNode := node
			name = path
			path = ""
			parent = prevNode
			node = uuid.New().String()
		}
	}

	if node == "" {
		node = uuid.New().String()
	}

	switch d.Data().(type) {
	case string, float64, bool:
		if path == "" {
			path = "val"
		}
		m.add(name, path, node, parent, d)

		return nil

	case []interface{}:
		c, err := d.Children()
		if err != nil {
			return err
		}

		if path != "" && name != "" {
			name = fmt.Sprintf("%s__%s", name, path)
		}

		for _, child := range c {
			m.do(name, "", uuid.New().String(), node, child)
		}

	case map[string]interface{}:
		cm, err := d.ChildrenMap()
		if err != nil {
			return err
		}

		for key, child := range cm {
			nextPath := key
			if path != "" {
				nextPath = fmt.Sprintf("%s__%s", path, key)
			}
			m.do(name, nextPath, node, parent, child)
		}

	default:
		panic("default case. We should not get here.")
	}
	return nil
}
