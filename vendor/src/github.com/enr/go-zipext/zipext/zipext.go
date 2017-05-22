package zipext

import (
	"archive/zip"
	"fmt"
	"github.com/enr/go-files/files"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// WalkFunc is the type of the function called for each file or directory
// visited by Walk.
//
// If there was a problem walking to the file or directory named by path, the
// incoming error will describe the problem and the function can decide how
// to handle that error (and Walk will not descend into that directory). If
// an error is returned, processing stops.
//
// TODO: this signature requires client code import archive/zip
type WalkFunc func(file *zip.File, err error) error

// walk recursively descends path, calling walkFn.
func walk(fileName string, walkFn WalkFunc) error {
	r, err := zip.OpenReader(fileName)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		err := walkFn(f, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// Walk walks the zip file rooted at root, calling walkFn for each file or
// directory in the zip, including root.
// All errors that arise visiting files and directories are filtered by walkFn.
func Walk(path string, walkFn WalkFunc) error {
	root := strings.TrimSpace(path)
	_, err := os.Lstat(root)
	if err != nil {
		return walkFn(nil, err)
	}
	return walk(root, walkFn)
}

func Extract(archivePath string, extractPath string) error {
	zipPath := strings.TrimSpace(archivePath)
	destinationPath := strings.TrimSpace(extractPath)
	if zipPath == "" || destinationPath == "" {
		return fmt.Errorf("path or destination is empty")
	}
	if !files.Exists(zipPath) {
		return fmt.Errorf("%s not found", zipPath)
	}
	if !files.IsDir(dirname(destinationPath)) {
		return fmt.Errorf("%s invalid path", destinationPath)
	}
	r, err := zip.OpenReader(zipPath)
	defer r.Close()
	if err != nil {
		return err
	}
	destinationBaseDir := filepath.ToSlash(destinationPath)
	fi, err := os.Stat(destinationBaseDir)
	if err != nil {
		return err
	}
	if files.Exists(destinationPath) && !fi.IsDir() {
		return fmt.Errorf("%s exists but it is NOT a directory!", destinationPath)
	}
	for _, f := range r.File {
		destination := fmt.Sprintf("%s/%s", destinationBaseDir, f.Name)
		basepath := dirname(destination)
		if err := os.MkdirAll(basepath, 0755); err != nil {
			return err
		}
		s, err := f.Open()
		defer s.Close()
		if err != nil {
			return err
		}
		dst := fmt.Sprintf("%s/%s", destinationBaseDir, f.Name)
		if files.Exists(dst) {
			continue
		}
		d, err := os.Create(dst)
		defer d.Close()
		if err != nil {
			return err
		}
		if _, err := io.Copy(d, s); err != nil {
			return err
		}
	}
	return nil
}

func dirname(path string) string {
	pathIndex := strings.LastIndex(path, "/")
	if pathIndex != -1 {
		return path[:pathIndex]
	}
	return "."
}

func addToZip(_path string, tw *zip.Writer, fi os.FileInfo, internalPath string) error {
	fr, err := os.Open(_path)
	defer fr.Close()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	header.Name = internalPath
	header.UncompressedSize64 = uint64(fi.Size())
	w, err := tw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, fr)
	if err != nil {
		return err
	}
	return nil
}

// Preferred ReadDir to filepath.Walk because...
// From filepath.Walk docs:
// for very large directories Walk can be inefficient. Walk does not follow symbolic links.
func walkDirectory(startPath string, tw *zip.Writer, basePath string, flat bool) error {
	dirPath := filepath.ToSlash(startPath)
	dir, err := os.Open(dirPath)
	defer dir.Close()
	if err != nil {
		return err
	}
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()
		if fi.IsDir() {
			err = walkDirectory(curPath, tw, basePath, flat)
			if err != nil {
				return err
			}
		} else {
			baseName := filepath.Base(basePath)
			if flat {
				baseName = ""
			}
			internalPath := strings.Replace(curPath, basePath, baseName, 1)
			internalPath = strings.TrimLeft(internalPath, "/")
			err = addToZip(curPath, tw, fi, internalPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateFlat(inputPath string, zipPath string) error {
	return createZip(inputPath, zipPath, true)
}

// if inputPath is a directory the zip will contain the directory
func Create(inputPath string, zipPath string) error {
	return createZip(inputPath, zipPath, false)
}

func createZip(inputPath string, zipPath string, flat bool) error {
	inPath := strings.TrimSpace(inputPath)
	outFilePath := strings.TrimSpace(zipPath)
	if inPath == "" || outFilePath == "" {
		return fmt.Errorf("path or destination is empty")
	}
	if !files.Exists(inPath) {
		return fmt.Errorf("invalid path %s", inPath)
	}
	if !files.IsDir(dirname(outFilePath)) {
		return fmt.Errorf("invalid path %s", outFilePath)
	}
	fw, err := os.Create(outFilePath)
	defer fw.Close()
	if err != nil {
		return err
	}
	zw := zip.NewWriter(fw)
	defer zw.Close()
	if files.IsDir(inPath) {
		err = walkDirectory(inPath, zw, inPath, flat)
		if err != nil {
			return err
		}
	} else {
		fi, err := os.Stat(inPath)
		if err != nil {
			return err
		}
		err = addToZip(inPath, zw, fi, filepath.Base(inPath))
		if err != nil {
			return err
		}
	}
	return nil
}
