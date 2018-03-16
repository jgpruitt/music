package main

import (
	"os"
	"encoding/csv"
	"fmt"
	"io"
)

/*
	A file named "dirs.csv" will have one column per row.
	The first column will be a directory to create in the new music collection.
	We'll go ahead and create all the directories ahead of time so that we don't have
	to check for their existence and create them while copying the music files into the
	new music collection.

	The "dirs.csv" file dumped from a Postgres table that was created manually/interactively.

	For instance:
	Column 1: C:\Users\jpruitt\Music\Beastie Boys\Pauls Boutique
*/

func main() {
	i, err := os.Open("dirs.csv")
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

		fmt.Println(record[0])
		err = os.MkdirAll(record[0], 0777)
		if err != nil {
			panic(err)
		}
	}
}
