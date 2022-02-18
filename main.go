package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Directory struct {
	name     string
	children []Tree
}

func (d Directory) String() string {
	return d.name
}

func (f File) String() string {
	if f.size == 0 {
		return f.name + " (empty)"
	} else {
		return f.name + " (" + strconv.FormatInt(f.size, 10) + "b)"
	}
}

type File struct {
	name string
	size int64
}

type Tree interface {
	String() string
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

func dirTree(out io.Writer, path string, printFiles bool) error {
	var tree []Tree
	if !printFiles {
		tree = walkDir(path, []Tree{})
	} else {
		tree = walkFilesAndDir(path, []Tree{})
	}
	getTree(out, tree, []string{})
	return nil
}

func getTree(out io.Writer, files []Tree, prefixes []string) {
	const (
		staging = "├───"
		last    = "└───"
		tab     = "│\t"
		space   = "\t"
	)
	for idx, line := range files {
		if len(files)-1 == idx {
			_, err := fmt.Fprintf(out, "%s%s%s\n", strings.Join(prefixes, ""), last, line)
			if err != nil {
				return
			}
			if directory, ok := line.(Directory); ok {
				getTree(out, directory.children, append(prefixes, space))
			}
		} else {
			_, err := fmt.Fprintf(out, "%s%s%s\n", strings.Join(prefixes, ""), staging, line)
			if err != nil {
				return
			}
			if directory, ok := line.(Directory); ok {
				getTree(out, directory.children, append(prefixes, tab))
			}
		}
	}
}

func readFilesAndDir(path string) (error, []os.FileInfo) {
	file, err := os.Open(path)
	if err != nil {
		return err, nil
	}
	files, err := file.Readdir(-1)
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	return err, files
}

func walkDir(path string, filesAndDir []Tree) []Tree {
	err, files := readFilesAndDir(path)
	if err != nil {
		return nil
	}
	for _, file := range files {
		if file.IsDir() {
			children := walkDir(filepath.Join(path, file.Name()), []Tree{})
			filesAndDir = append(filesAndDir, Directory{file.Name(), children})
		}
	}
	return filesAndDir
}

func walkFilesAndDir(path string, filesAndDir []Tree) []Tree {
	err, files := readFilesAndDir(path)
	if err != nil {
		return nil
	}
	for _, file := range files {
		var line Tree
		if file.IsDir() {
			children := walkFilesAndDir(filepath.Join(path, file.Name()), []Tree{})
			line = Directory{file.Name(), children}
		} else {
			line = File{file.Name(), file.Size()}
		}
		filesAndDir = append(filesAndDir, line)
	}
	return filesAndDir
}
