package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type catalogEntry struct {
	Name string `json:"name"`
	File string `json:"file"`
}

var parentheticalSuffix = regexp.MustCompile(`\s*\([^)]*\)`)

func main() {
	if len(os.Args) != 3 {
		fail(errors.New("usage: romcatalog <source directory> <output directory>"))
	}

	sourceDirectory := os.Args[1]
	outputDirectory := os.Args[2]
	if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
		fail(err)
	}

	files, err := romFiles(sourceDirectory)
	if err != nil {
		fail(err)
	}

	catalog := make([]catalogEntry, 0, len(files))
	for _, file := range files {
		if err := copyFile(filepath.Join(sourceDirectory, file), filepath.Join(outputDirectory, file)); err != nil {
			fail(err)
		}
		catalog = append(catalog, catalogEntry{Name: displayName(file), File: file})
	}

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		fail(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(outputDirectory, "catalog.json"), data, 0o644); err != nil {
		fail(err)
	}
}

func romFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.EqualFold(filepath.Ext(entry.Name()), ".nes") {
			files = append(files, entry.Name())
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i]) < strings.ToLower(files[j])
	})
	return files, nil
}

func displayName(file string) string {
	name := strings.TrimSuffix(file, filepath.Ext(file))
	name = strings.ReplaceAll(name, "_", " ")
	name = parentheticalSuffix.ReplaceAllString(name, "")
	words := strings.Fields(name)
	for index, word := range words {
		letters := []rune(word)
		letters[0] = unicode.ToUpper(letters[0])
		words[index] = string(letters)
	}
	return strings.Join(words, " ")
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
