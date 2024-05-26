package main

import (
	"data"

	"bazil.org/fuse/fs"
)

type FS struct {
	root *Dir
	db   *data.Queries
}

func (f *FS) Root() (fs.Node, error) {
	// dir, err := f.db.SelectOneDirectoryName(context.Background(), "root")
	// if err != nil {
	// 	return nil, err
	// }
	return f.root, nil
}
