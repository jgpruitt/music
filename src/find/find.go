package main

import (
	"sync"
	"os"
	"fmt"
	"encoding/csv"
	"time"
	"github.com/dhowden/tag"
	"path/filepath"
	"crypto/md5"
	"io"
	"strings"
)

/*
	This program recursively searches 6 directories to find music files matching a list of file extensions.
	For each matching file, it:
		1. collects basic information about the file
		2. reads the ID3 tag data out of the file
		3. MD5 hashes the entire file (so we can find true file duplicates)
		4. checksums the music content in the file (so we can find duplicates which may differ only in ID3 tag data)
		5. writes the information to a csv file
*/

type record struct {
	file string
	ext string
	size int64
	mod time.Time
	md5 string       // hash of the entire file
	musichash string // checksum of just the music content within the file
	format string
	filetype string
	title string
	album string
	artist string
	albumartist string
	composer string
	genre string
	year int
	tracknbr int
	tracktot int
	disknbr int
	disktot int
}

var exts map[string]struct{}

func main() {
	// the 6 source directories to process
	roots := []string{
		`C:\music\music1\`,
		`C:\music\music2\`,
		`C:\music\music3\`,
		`C:\music\music4\`,
		`C:\music\music5\`,
		`C:\music\music6\`,
	}

	// a "set" for the file extensions we care about
	exts = make(map[string]struct{})
	exts[`.m4a`] = struct{}{}
	exts[`.m4p`] = struct{}{}
	exts[`.m4v`] = struct{}{}
	exts[`.mp3`] = struct{}{}
	exts[`.ogg`] = struct{}{}
	exts[`.wav`] = struct{}{}
	exts[`.wma`] = struct{}{}

	// a channel to pass records from the finders to the processors
	records1 := make(chan *record, 20)

	// 6 goroutines (1 per source dir) to find all the music files
	wg1 := &sync.WaitGroup{}
	wg1.Add(len(roots))
	for _, root := range roots {
		go find(root, wg1, records1)
	}

	// a channel to pass records from the processors to the csv writer
	records2 := make(chan *record, 20)

	// 8 goroutines to process the music files
	//   * hash the whole file
	//   * pull the meta data out of the id3 tag
	//   * hash just the music content in the file
	wg2 := &sync.WaitGroup{}
	wg2.Add(8)
	for i := 0; i < 8; i++ {
		go process(wg2, records1, records2)
	}

	// a channel to signal that the goroutine that is writing the csv file has completed
	done := make(chan struct{})
	// 1 goroutine to write all the information to a csv file
	go out(done, records2)

	// wait for the finders to complete
	wg1.Wait()
	fmt.Println("Found all files!")
	close(records1)
	// wait for the processors to complete
	wg2.Wait()
	fmt.Println("Processed all files!")
	close(records2)
	// wait for the csv writer to complete
	<-done
	fmt.Println("Wrote all output!")
	fmt.Println("Done.")
}

// find recursively searches a directory for files with certain extensions and gets basic file information for matching files
func find(root string, wg *sync.WaitGroup, records chan *record) {
	defer wg.Done()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := exts[ext]; !ok {
			return nil
		}
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		r := &record{}
		r.file = path
		r.ext = filepath.Ext(path)
		r.size = info.Size()
		r.mod = info.ModTime()
		records<-r
		return err
	})
	if err != nil {
		fmt.Println("FAILED to walk", root, err)
		return
	}
}

// process reads the id3 tag data, hashes the file, and checksums the music content
func process(wg *sync.WaitGroup, records1 chan *record, records2 chan *record) {
	defer wg.Done()
	for r := range records1 {
		hash(r)
		// the tag library doesn't understand wma files
		if strings.ToLower(r.ext) != ".wma" {
			meta(r)
		}
		records2<-r
	}
}

// hash MD5 hashes the entire file
func hash(r *record) {
	i, err := os.Open(r.file)
	if err != nil {
		fmt.Println("FAILED to open", r.file, err)
		return
	}
	defer i.Close()

	h := md5.New()
	if _, err := io.Copy(h, i); err != nil {
		fmt.Println("FAILED to hash", r.file, err)
		return
	}
	r.md5 = fmt.Sprintf("%X", h.Sum(nil))
}

// meta parses the id3 tag data
func meta(r *record) {
	i, err := os.Open(r.file)
	if err != nil {
		fmt.Println("FAILED to open", r.file, err)
		return
	}
	defer i.Close()

	m, err := tag.ReadFrom(i)
	if err != nil {
		fmt.Println("FAILED to read metadata from", r.file, err)
		return
	}
	r.format = string(m.Format())
	r.filetype = string(m.FileType())
	if r.format != string(tag.UnknownFormat) && r.filetype != string(tag.UnknownFileType) {
		r.title = clean(m.Title())
		r.album = clean(m.Album())
		r.artist = clean(m.Artist())
		r.albumartist = clean(m.AlbumArtist())
		r.composer = clean(m.Composer())
		r.year = m.Year()
		r.genre = m.Genre()
		r.tracknbr, r.tracktot = m.Track()
		r.disknbr, r.disktot = m.Disc()
		checksum(r)
	}
}

// checksum produces a checksum of the music content within the file
func checksum(r *record) {
	// unfortunately the tag library seemed to fail if the tag.Sum() function
	// didn't use a "fresh" file handle
	// seeking an existing open file handle to the beginning of the file didn't work
	i, err := os.Open(r.file)
	if err != nil {
		fmt.Println("FAILED to open", r.file, err)
		return
	}
	defer i.Close()
	r.musichash, err = tag.Sum(i)
	if err != nil {
		fmt.Println("FAILED to checksum music content in", r.file, err)
	}
}

// clean restricts the characters we'll allow in the title, album, artist, albumartist, and composer fields
func clean(str string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case 'A' <= r && r <= 'Z':
			return r
		case 'a' <= r && r <= 'z':
			return r
		case '0' <= r && r <= '9':
			return r
		case r == ' ':
			return r
		}
		return -1
	}, str)
}

// out writes the data to a csv
func out(done chan struct{}, records chan *record) {
	var nbr int64
	nbr = 0
	defer func() {
		done<-struct{}{}
	}()
	o, err := os.Create("files.csv")
	if err != nil {
		fmt.Println("FAILED to create files.csv", err)
		return
	}
	defer o.Close()
	w := csv.NewWriter(o)
	for r := range records {
		w.Write([]string{
			r.file,
			r.ext,
			fmt.Sprintf("%d", r.size),
			r.mod.Format("2006-01-02 15:04:05.000 -0700"),
			r.md5,
			r.musichash,
			r.format,
			r.filetype,
			r.title,
			r.album,
			r.albumartist,
			r.artist,
			r.composer,
			r.genre,
			fmt.Sprintf("%d", r.year),
			fmt.Sprintf("%d", r.tracknbr),
			fmt.Sprintf("%d", r.tracktot),
			fmt.Sprintf("%d", r.disknbr),
			fmt.Sprintf("%d", r.disktot),
		})
		if err := w.Error(); err != nil {
			fmt.Println("FAILED to write", r.file, "to csv!", err)
		}
		nbr = nbr + 1
		if nbr % 50 == 0 {
			fmt.Println(nbr)
		}
	}
	fmt.Println(nbr)
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("FAILED to flush csv writer!", err)
	}
}
