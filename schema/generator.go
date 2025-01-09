package schema

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/jbrough/json2Leaf"
)

// sometimes we may want to srongly name files when creating them by appending
// a hash of their contents that provides for multiple versions of the same
// document. But these hashes aren't useful in the node name.
// TODO: this hash is not useful for the sql import but we should add a version
// id to all rows when inserting new records from file. A time sortable ksuid?
var (
	mdHashDoubleUnderscore = regexp.MustCompile(`__[a-f0-9]{32}__`)
	mdHashEnd              = regexp.MustCompile(`__[a-f0-9]{32}$`)
)

func cleanName(name string) string {
	name = mdHashDoubleUnderscore.ReplaceAllString(name, "__")
	return mdHashEnd.ReplaceAllString(name, "")
}

type Generator struct {
	File    *os.File
	Writer  *bufio.Writer
	mu      sync.Mutex
	types   map[string]string
	started bool
}

func NewGenerator() *Generator {
	return &Generator{
		types: map[string]string{
			"string":  "TEXT",
			"float64": "NUMERIC",
			"bool":    "BOOLEAN",
		},
	}
}

func (g *Generator) Close() error {
	if g.started {
		if _, err := g.Writer.WriteString("\\.\n"); err != nil {
			return err
		}
	}
	if err := g.Writer.Flush(); err != nil {
		return err
	}
	return g.File.Close()
}

func (g *Generator) WriteInitScript() error {
	schema := `DROP TABLE IF EXISTS nodes CASCADE;

CREATE TABLE nodes (
    id VARCHAR,
    parent_id VARCHAR, 
    name VARCHAR NOT NULL,
    path VARCHAR NOT NULL,
    data_type VARCHAR NOT NULL,
    value TEXT
);
`
	_, err := g.Writer.WriteString(schema)
	return err
}

func (g *Generator) WriteLeaves(leaves []json2Leaf.Leaf) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.started {
		if _, err := g.Writer.WriteString("COPY nodes (id, parent_id, name, path, data_type, value) FROM stdin;\n"); err != nil {
			return err
		}
		g.started = true
	}

	for _, leaf := range leaves {
		value := leaf.Value
		if leaf.Name == "_tree" {
			value = cleanName(fmt.Sprintf("%v", value))
		}
		value = formatValue(value, leaf.DataType)

		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n",
			escapeCopy(leaf.ID),
			nullableStringCopy(leaf.ParentID),
			escapeCopy(cleanName(leaf.Name)),
			escapeCopy(leaf.Path),
			escapeCopy(leaf.DataType),
			value,
		)
		if _, err := g.Writer.WriteString(row); err != nil {
			return err
		}
	}

	return g.Writer.Flush()
}

func formatValue(v interface{}, dataType string) string {
	if v == nil {
		return "\\N"
	}

	switch dataType {
	case "string":
		return escapeCopy(fmt.Sprintf("%v", v))
	case "bool":
		return strings.ToLower(fmt.Sprintf("%v", v))
	case "float64":
		return fmt.Sprintf("%v", v)
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "\\N"
		}
		return escapeCopy(string(jsonBytes))
	}
}

func escapeCopy(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

func nullableStringCopy(s string) string {
	if s == "" {
		return "\\N"
	}
	return escapeCopy(s)
}
