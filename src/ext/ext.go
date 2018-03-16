package main

import (
	"path/filepath"
	"os"
	"fmt"
	"strings"
	"encoding/csv"
)

/*
	I wanted to know what kinds of files existed in the source directories.
	This little program creates a distinct list of file extensions found
	and a count of the files per extension. It dumps it to "ext.csv".
*/

func main() {
	roots := []string{
		`C:\music\music1\`,
		`C:\music\music2\`,
		`C:\music\music3\`,
		`C:\music\music4\`,
		`C:\music\music5\`,
		`C:\music\music6\`,
	}

	m := make(map[string]int64)

	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			ext := filepath.Ext(path)
			if _, ok := m[ext]; !ok {
				m[ext] = 1
			} else {
				m[ext] = m[ext] + 1
			}
			return nil
		})
		if err != nil {
			fmt.Println("FAILED to walk", root, err)
		}
	}

	o, err := os.Create("ext.csv")
	if err != nil {
		panic(err)
	}
	defer o.Close()
	w := csv.NewWriter(o)
	for k, v := range m {
		w.Write([]string{k, fmt.Sprintf("%d", v)})
	}
	w.Flush()
}
