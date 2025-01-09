package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	xml2json "github.com/basgys/goxml2json"
	"github.com/jbrough/json2Leaf"
	"github.com/jbrough/json2Leaf/schema"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <input_dir>")
		os.Exit(1)
	}
	inputDir := os.Args[1]
	config := json2Leaf.NewConfig()
	generator := schema.NewGenerator()
	f, err := os.Create("output.sql")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	generator.File = f
	generator.Writer = bufio.NewWriter(f)
	defer generator.Close()
	if err := generator.WriteInitScript(); err != nil {
		fmt.Printf("Error writing schema: %v\n", err)
		os.Exit(1)
	}
	files := []string{}
	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			switch strings.ToLower(filepath.Ext(path)) {
			case ".xml", ".json":
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d files\n", len(files))
	var totalLeaves int
	for i, path := range files {
		fmt.Printf("[%d/%d] Processing %s\n", i+1, len(files), filepath.Base(path))
		ext := strings.ToLower(filepath.Ext(path))
		var jsonData []byte
		if ext == ".xml" {
			fileData, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}
			unescaped := strings.ReplaceAll(string(fileData), "\\n", "\n")
			unescaped = strings.ReplaceAll(unescaped, "\\\"", "\"")
			converted, err := xml2json.Convert(strings.NewReader(unescaped))
			if err != nil {
				fmt.Println("Error converting XML to JSON:", err)
				return
			}

			jsonData = converted.Bytes()
		} else {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				continue
			}
			jsonData = data
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		mapper := json2Leaf.NewMapper(config)
		leaves, err := mapper.Do(name, jsonData)
		if err != nil {
			fmt.Printf("Error processing file: %v\n", err)
			continue
		}
		totalLeaves += len(leaves)
		fmt.Printf("Generated %d leaves (total: %d)\n", len(leaves), totalLeaves)
		if err := generator.WriteLeaves(leaves); err != nil {
			fmt.Printf("Error writing leaves: %v\n", err)
		}
	}
	fmt.Printf("Done! Processed %d files, generated %d leaves\n", len(files), totalLeaves)
}
