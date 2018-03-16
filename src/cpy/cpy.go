package main

import (
	"os"
	"encoding/csv"
	"io"
	"fmt"
)

/*
	A file named "cpy.csv" will have two columns per row.
	The first column will be the path to the existing music file. The second column
	will be where we want to copy the file to in the new "clean" music collection.
	The "cpy.csv" file dumped from a Postgres table that was created manually/interactively.

	For instance:
	Column 1: C:\music\music4\The Beastie Boys\Paul's Boutique [Explicit]\05 High Plains Drifter [Explicit].mp3
	Column 2: C:\Users\jpruitt\Music\Beastie Boys\Pauls Boutique\05 High Plains Drifter.mp3
*/


func main() {
	i, err := os.Open("cpy.csv")
	if err != nil {
		panic(err)
	}
	defer i.Close()
	r := csv.NewReader(i)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if len(record) != 2 {
			fmt.Println("Bad field count:", len(record), record)
			continue
		}
		cpyit(record[0], record[1])
	}
	fmt.Println("DONE!")
}

func cpyit(fpath, tpath string) {
	fmt.Println(fpath, " -> ", tpath)
	f, err := os.Open(fpath)
	if err != nil {
		fmt.Println("FAILED to open", fpath)
		return
	}
	defer f.Close()
	t, err := os.Create(tpath)
	if err != nil {
		fmt.Println("FAILED to create", tpath)
		return
	}
	defer t.Close()
	_, err = io.Copy(t, f)
	if err != nil {
		fmt.Println("FAILED to copy", fpath, "to", tpath, err)
		return
	}
	t.Sync()
}
