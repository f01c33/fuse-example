package main

import (
	"context"
	"data"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	_ "modernc.org/sqlite"
)

type Node struct {
	inode int64
	name  string
}

// var inode int64

func NewInode() int64 {
	return int64(rand.Int63n(math.MaxInt64))
}

//go:generate sqlc generate

//go:embed schema.sql
var ddl string

func buildFile(file data.File) *File {
	f := File{Node: Node{inode: file.Inode, name: file.Name}, data: file.Data, parent: file.Parent}
	return &f
}

func buildDir(dir data.Directory) (*Dir, error) {
	d := Dir{Node: Node{inode: dir.Inode, name: dir.Name}, directories: []*Dir{}, files: []*File{}}
	dirs, err := DB.SelectDirectoriesParent(context.Background(), dir.Inode)
	if err != nil {
		return nil, err
	}
	// fmt.Println("dirs", dirs)
	for _, dir := range dirs {
		fmt.Println(dir.Name, "dir ", dir)
		dr, err := buildDir(dir)
		if err != nil {
			return nil, err
		}
		d.directories = append(d.directories, dr)
	}
	files, err := DB.SelectFilesParent(context.Background(), dir.Inode)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		fmt.Println(dir.Name, "file", f)
		file := buildFile(f)
		d.files = append(d.files, file)
	}
	fmt.Println(dir.Name, "directories", d.directories, "files", d.files)
	return &d, nil
}

func buildFS() (*FS, error) {
	fs := FS{}
	dir, err := DB.SelectOneDirectoryInode(context.Background(), 0)
	if err != nil {
		return nil, err
	}
	fs.root, err = buildDir(dir)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

var DB *data.Queries

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
	DB = data.New(db)
	_ = DB.InsertDirectory(context.Background(), data.InsertDirectoryParams{
		Inode:  0,
		Name:   "root",
		Parent: -1,
	})

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
	filesys, err := buildFS()
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
