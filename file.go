package main

import (
	"data"
	"fmt"
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
)

type File struct {
	Node
	data   []byte
	parent int64
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Requested Attr for File", f.name, "has data size", len(f.data))
	a.Inode = uint64(f.inode)
	a.Mode = 0777
	a.Size = uint64(len(f.data))
	return nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("Requested Read on File", f.name)
	file, err := DB.SelectOneFileName(context.Background(), f.name)
	if err != nil {
		return err
	}
	fuseutil.HandleRead(req, resp, file.Data[req.Offset:req.Offset+int64(req.Size)])
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("Reading all of file", f.name)
	file, err := DB.SelectOneFileInode(context.Background(), int64(f.inode))
	if err != nil {
		return nil, err
	}
	return []byte(file.Data), nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("Trying to write to ", f.name, "offset", req.Offset, "dataSize:", len(req.Data), "data: ", string(req.Data))
	resp.Size = len(req.Data)
	// f.data = append(f.data[req.Offset:], append(req.Data, f.data[len(req.Data)+int(req.Offset):]...)...)
	f.data = req.Data
	log.Println("Wrote to file", f, fmt.Sprintf("%#v", req))
	return DB.UpdateFile(context.Background(), data.UpdateFileParams{
		Inode:   f.inode,
		Name:    f.name,
		Parent:  f.parent,
		Data:    f.data,
		Inode_2: f.inode,
	})
}

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	log.Println("Flushing file", f.name)
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Println("Open call on file", f.name)

	return f, nil
}

func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	log.Println("Release requested on file", f.name)
	return nil
}

func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Println("Fsync call on file", f.name)
	return nil
}
