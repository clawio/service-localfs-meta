package main

import (
	"github.com/clawio/service.auth/lib"
	"io"
	"os"
	"path"
	"strings"
)

// getHome returns the user home directory.
// the logical home has this layout.
// local/users/<letter>/<pid>
// Example: /local/users/o/ourense
// idt.Pid must be always non-empty
func getHome(idt *lib.Identity) string {

	pid := path.Clean(idt.Pid)

	if pid == "" {
		panic("idt.Pid must not be empty")
	}

	return path.Join("local", "users", string(pid[0]), pid)
}

// isUnderHome checks is the path is under a user home dir or not.
func isUnderHome(p string, idt *lib.Identity) bool {

	p = path.Clean(p)

	if strings.HasPrefix(p, getHome(idt)) {
		return true
	}

	return false
}

// copyFile copies a file from src to dst.
// src and dst are physycal paths.
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

// copyDir copies a dir from src to dst.
// src and dst are physycal paths.
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
