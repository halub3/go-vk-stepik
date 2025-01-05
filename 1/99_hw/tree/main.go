package main

import (
	"fmt"
	"io"
	"os"
)

var (
	Trait    string = "├───"
	HTrait   string = "└───"
	Stick    string = "│"
	StickTab string = "│\t"
	Tab      string = "\t"
)

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

func dirTree(w io.Writer, dir string, isFiles bool) (err error) {
	return dirTreeRec(w, dir, isFiles, "")
}

func dropFiles(in []os.DirEntry) []os.DirEntry {
	var res []os.DirEntry
	for _, v := range in {
		if v.IsDir() {
			res = append(res, v)
		}
	}
	return res
}

func dirTreeRec(w io.Writer, dir string, isFiles bool, prefix string) (err error) {

	objs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	if !isFiles {
		objs = dropFiles(objs)
	}

	lastIdx := len(objs) - 1
	for idx, obj := range objs {
		var startString, nextPrefix string

		isLast := idx == lastIdx

		if isLast {
			startString = prefix + HTrait
			nextPrefix = prefix + Tab
		} else {
			startString = prefix + Trait
			nextPrefix = prefix + StickTab
		}

		fmt.Fprint(w, startString+obj.Name())

		if obj.IsDir() {
			fmt.Fprint(w, "\n")
			err = dirTreeRec(w, dir+"\\"+obj.Name(), isFiles, nextPrefix)
			if err != nil {
				return
			}
		} else {
			info, err := obj.Info()
			if err != nil {
				return err
			}
			size := info.Size()
			if size > 0 {
				fmt.Fprint(w, " (", size, "b)\n")
			} else {
				fmt.Fprint(w, " (empty)\n")
			}
		}
	}

	return

}
