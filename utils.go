package main

import (
	"io"
	"os"
	"path"
)

func copyFile(src, dst string, size int64) (err error) {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.CopyN(writer, reader, size)
	if err != nil {
		return err
	}
	return nil
}

func copyDir(src, dst string) (err error) {
	err = os.Mkdir(dst, dirPerm)
	if err != nil {
		return err
	}

	directory, err := os.Open(src)
	if err != nil {
		return err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		_src := path.Join(src, obj.Name())
		_dst := path.Join(dst, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(_src, _dst)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = copyFile(_src, _dst, obj.Size())
			if err != nil {
				return err
			}
		}
	}
	return
}
