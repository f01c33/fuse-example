package main

import (
	"context"
	"data"
	"database/sql"
	_ "embed"
	"flag"
	"log"
	"math"
	"math/rand"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	_ "modernc.org/sqlite"
)

type Node struct {
	inode uint64
	name  string
}

// var inode uint64

func NewInode() uint64 {
	return uint64(rand.Int63n(math.MaxInt64))
}

//go:generate sqlc generate

//go:embed schema.sql
var ddl string

func buildFile(file data.File) *File {
	f := File{Node: Node{inode: uint64(file.Inode), name: file.Name}, data: file.Data}
	return &f
}

func buildDir(DB *data.Queries, dir data.Directory) (*Dir, error) {
	d := Dir{Node: Node{inode: uint64(dir.Inode), name: dir.Name}}
	dirs, err := DB.SelectDirectoriesParent(context.Background(), dir.Inode)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		d, err := buildDir(DB, dir)
		if err != nil {
			return nil, err
		}
		*d.directories = append(*d.directories, d)
	}
	files, err := DB.SelectFilesParent(context.Background(), dir.Inode)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		file := buildFile(f)
		*d.files = append(*d.files, file)
	}
	return &d, nil
}

func buildFS(DB *data.Queries) (*FS, error) {
	fs := FS{}
	dir, err := DB.SelectOneDirectoryInode(context.Background(), "root")
	if err != nil {
		return nil, err
	}
	fs.root, err = buildDir(DB, dir)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func main() {
	mountPoint := ""
	dbArg := ""
	flag.StringVar(&mountPoint, "m", "fs", "the folder to mount to")
	flag.StringVar(&dbArg, "db", "db.sqlite", "The db to mount")
	flag.Parse()

	db, err := sql.Open("sqlite", dbArg)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := db.ExecContext(context.Background(), ddl); err != nil {
		log.Fatal(err)
	}
	DB := data.New(db)
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mounting ", dbArg, "in directory", mountPoint)
	defer c.Close()
	// if p := c.Protocol(); !p.HasInvalidate() {
	// 	log.Panicln("kernel FUSE support is too old to have invalidations: version %v", p)
	// }
	srv := fs.New(c, nil)
	filesys, err := buildFS(DB)
	if err != nil {
		log.Fatalln(err)
	}
	// &Dir{Node: Node{name: "head", inode: NewInode()}, files: &[]*File{
	// 	&File{Node: Node{name: "hello", inode: NewInode()}, data: []byte("hello world!")},
	// 	&File{Node: Node{name: "aybbg", inode: NewInode()}, data: []byte("send notes")},
	// }, directories: &[]*Dir{
	// 	&Dir{Node: Node{name: "left", inode: NewInode()}, files: &[]*File{
	// 		&File{Node: Node{name: "yo", inode: NewInode()}, data: []byte("ayylmaooo")},
	// 	},
	// 	},
	// 	&Dir{Node: Node{name: "right", inode: NewInode()}, files: &[]*File{
	// 		&File{Node: Node{name: "hey", inode: NewInode()}, data: []byte("heeey, thats pretty good")},
	// 	},
	// 	},
	// },
	// 	db: DB},
	// DB}
	log.Println("About to serve fs")
	if err := srv.Serve(filesys); err != nil {
		log.Panicln(err)
	}
	// Check if the mount process has an error to report.
	// <-c.
	// if err := c.; err != nil {
	// 	log.Panicln(err)
	// }
}
