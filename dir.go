package main

import (
	"data"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context" // need this cause bazil lib doesn't use syslib context lib
)

type Dir struct {
	Node
	files       []*File
	directories []*Dir
	parent      int64
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("Requested Attr for Directory", d.name)
	a.Inode = uint64(d.inode)
	a.Mode = os.ModeDir | 0444
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Requested lookup for ", name)
	if d.files != nil {
		for _, n := range d.files {
			if n.name == name {
				log.Println("Found match for directory lookup with size", len(n.data))
				return n, nil
			}
		}
	}
	if d.directories != nil {
		for _, n := range d.directories {
			if n.name == name {
				log.Println("Found match for directory lookup")
				return n, nil
			}
		}
	}
	return nil, syscall.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Reading all dirs")
	dir, err := DB.SelectOneDirectoryInode(context.Background(), d.inode)
	if err != nil {
		return nil, err
	}
	d, err = buildDir(dir)
	if err != nil {
		return nil, err
	}
	var children []fuse.Dirent
	if d.files != nil {
		for _, f := range d.files {
			children = append(children, fuse.Dirent{Inode: uint64(f.inode), Type: fuse.DT_File, Name: f.name})
		}
	}
	if d.directories != nil {
		for _, dir := range d.directories {
			children = append(children, fuse.Dirent{Inode: uint64(dir.inode), Type: fuse.DT_Dir, Name: dir.name})
		}
		log.Println(len(children), " children for dir", d.name)
	}
	return children, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Create request for name", req.Name)
	f := &File{Node: Node{name: req.Name, inode: NewInode()}}
	files := []*File{f}
	if d.files != nil {
		files = append(files, d.files...)
	}
	d.files = files
	DB.InsertFile(context.Background(), data.InsertFileParams{
		Inode:  f.inode,
		Name:   f.name,
		Parent: d.inode,
		Data:   f.data,
	})
	return f, f, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Println("Remove request for ", req.Name)
	if req.Dir && d.directories != nil {
		newDirs := []*Dir{}
		for _, dir := range d.directories {
			if dir.name != req.Name {
				newDirs = append(newDirs, dir)
			} else if dir.name == req.Name {
				err := DB.DeleteDirectoryInode(context.Background(), dir.inode)
				if err != nil {
					return err
				}
			}
		}
		d.directories = newDirs
		return nil
	} else if !req.Dir && d.files != nil {
		newFiles := []*File{}
		for _, f := range d.files {
			if f.name != req.Name {
				newFiles = append(newFiles, f)
			} else if f.name == req.Name {
				err := DB.DeleteFileInode(context.Background(), f.inode)
				if err != nil {
					return err
				}
			}
		}
		d.files = newFiles
		return nil
	}
	return syscall.ENOENT
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdir request for name", req.Name)
	dir := &Dir{Node: Node{name: req.Name, inode: NewInode()}}
	directories := []*Dir{dir}
	if d.directories != nil {
		directories = append(d.directories, directories...)
	}
	d.directories = directories
	DB.InsertDirectory(context.Background(), data.InsertDirectoryParams{
		Inode:  dir.inode,
		Name:   dir.name,
		Parent: d.inode,
	})
	return dir, nil

}
