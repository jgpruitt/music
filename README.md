# music

Personal project to clean up and deduplicate multiple copies of my music library.

## Background

I was recently switching machines and wanted a "clean" copy of my music collection on the new machine.
I had 6 copies of my music collection from previous machine moves. None of the 6 copies were complete and
there were many duplicate files. The 6 copies totaled 250GB on disk.

I wanted to get a list of all the music files into PostgreSQL, so that I could use SQL to produce a list
of the source files I wanted to copy into the new music collection with new "clean" file names/directories. 

## Disclaimer

This was a quick and dirty personal project that only really had to work *once*. As such, it is missing logging, 
configuration, and error handling features that would be present in something meant for professional production usage.

## What I did...

### Step 1

I ran the src/ext/ext.go program to produce a distinct list with counts of file extensions found in the 6
source directories. I used this list to decide on a list of the extensions I cared about. The rest would be ignored.

### Step 2

I ran the src/find/find.go program to recursively search the 6 source directories for files of the extensions I
cared about. For each file, it collected basic file information, ID3 tag information, whole file MD5 hashes, and 
checksums of the music content only. This information was written to a csv file.

This program was written with performance in mind. I didn't want to have to let it run overnight to process 250GB. I
used parallelism and concurrency to speed it up. It processed the 250GB in about 10 minutes. I was getting sustained
disk I/O of 435MB/s out of a consumer-grade SATA SSD. The CPU utilization stayed under 10%.

### Step 3

I loaded the csv from Step 2 into PostgreSQL and ran a bunch of SQL interactively to massage the data. I produced a
table of directories I wanted created in the new music collection. I produced a table of source file to new file
mappings to populate the music collection with music files.

### Step 4

I dumped the table of directories to a csv and ran the src/dirs/dirs.go program to create the directories.

### Step 5

I dumped the table of source file to new file mappings to a csv and ran the src/cpy/cpy.go program to copy these
source files into the new music collection.

