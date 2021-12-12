package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

func printFile(out io.Writer, file os.FileInfo, printFiles bool, prefix string, header string) error {
	fileSize := file.Size()
	var size string
	if fileSize <= 0 {
		size = "empty"
	} else {
		size = strconv.Itoa(int(fileSize)) + "b"
	}
	if printFiles {
		_, err := fmt.Fprintf(out, "%s%s%s (%s)\n", prefix, header, file.Name(), size)
		if err != nil {
			return err
		}
	}
	return nil
}

func printDir(out io.Writer, dir os.FileInfo, prefix string, header string) error {
	_, err := fmt.Fprintf(out, "%s%s%s\n", prefix, header, dir.Name())
	if err != nil {
		return err
	}
	return nil
}

func header(lastEntry bool) string {
	header := "├───"
	lastHeader := "└───"
	head := header
	if lastEntry {
		head = lastHeader
	}

	return head
}

func cleanupList(list []os.FileInfo, printFiles bool) (result []os.FileInfo) {
	if !printFiles {
		for _, f := range list {
			if f.IsDir() {
				result = append(result, f)
			}
		}
	} else {
		result = append(result, list...)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return
}

func walkDir(out io.Writer, path string, printFiles bool, prefix string) error {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	cleanEntries := cleanupList(entries, printFiles)
	for i, entry := range cleanEntries {
		lastEntry := i == (len(cleanEntries) - 1)
		if entry.IsDir() {
			if err := printDir(out, entry, prefix, header(lastEntry)); err != nil {
				return err
			}
			var newPrefix string
			if lastEntry {
				newPrefix = prefix + "\t"
			} else {
				newPrefix = prefix + "│\t"
			}
			newPath := filepath.Join(path, entry.Name())
			if err = walkDir(out, newPath, printFiles, newPrefix); err != nil {
				return err
			}
		} else {
			err := printFile(out, entry, printFiles, prefix, header(lastEntry))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	_ = walkDir(out, path, printFiles, "")

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
