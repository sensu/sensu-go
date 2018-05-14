package test_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/httpfs/vfsutil"
	"github.com/shurcooL/httpgzip"
)

//go:generate go run test_gen.go

// Basic functionality test.
func Example_basic() {
	var fs http.FileSystem = assets

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}

		fmt.Println(path)
		if fi.IsDir() {
			return nil
		}

		b, err := vfsutil.ReadFile(fs, path)
		fmt.Printf("%q %v\n", string(b), err)
		return nil
	}

	err := vfsutil.Walk(fs, "/", walkFn)
	if err != nil {
		panic(err)
	}

	// Output:
	// /
	// /folderA
	// /folderA/file1.txt
	// "Stuff in /folderA/file1.txt." <nil>
	// /folderA/file2.txt
	// "Stuff in /folderA/file2.txt." <nil>
	// /folderB
	// /folderB/folderC
	// /folderB/folderC/file3.txt
	// "Stuff in /folderB/folderC/file3.txt." <nil>
	// /not-worth-compressing-file.txt
	// "Its normal contents are here." <nil>
	// /sample-file.txt
	// "This file compresses well. Blaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaah!" <nil>
}

func Example_compressed() {
	// Compressed file system.
	var fs http.FileSystem = assets

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}

		fmt.Println(path)
		if fi.IsDir() {
			return nil
		}

		f, err := fs.Open(path)
		if err != nil {
			fmt.Printf("fs.Open(%q): %v\n", path, err)
			return nil
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		fmt.Printf("%q %v\n", string(b), err)

		if gzipFile, ok := f.(httpgzip.GzipByter); ok {
			b := gzipFile.GzipBytes()
			fmt.Printf("%q\n", string(b))
		} else {
			fmt.Println("<not compressed>")
		}
		return nil
	}

	err := vfsutil.Walk(fs, "/", walkFn)
	if err != nil {
		panic(err)
	}

	// Output:
	// /
	// /folderA
	// /folderA/file1.txt
	// "Stuff in /folderA/file1.txt." <nil>
	// <not compressed>
	// /folderA/file2.txt
	// "Stuff in /folderA/file2.txt." <nil>
	// <not compressed>
	// /folderB
	// /folderB/folderC
	// /folderB/folderC/file3.txt
	// "Stuff in /folderB/folderC/file3.txt." <nil>
	// <not compressed>
	// /not-worth-compressing-file.txt
	// "Its normal contents are here." <nil>
	// <not compressed>
	// /sample-file.txt
	// "This file compresses well. Blaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaah!" <nil>
	// "\x1f\x8b\b\x00\x00\x00\x00\x00\x00\xff\n\xc9\xc8,VH\xcb\xccIUH\xce\xcf-(J-.N-V(O\xcd\xc9\xd1Sp\xcaI\x1c\xd4 C\x11\x10\x00\x00\xff\xff\xe7G\x81:\xbd\x00\x00\x00"
}

func Example_readTwoOpenedCompressedFiles() {
	var fs http.FileSystem = assets

	f0, err := fs.Open("/sample-file.txt")
	if err != nil {
		panic(err)
	}
	defer f0.Close()
	_ = f0.(httpgzip.GzipByter)
	f1, err := fs.Open("/sample-file.txt")
	if err != nil {
		panic(err)
	}
	defer f1.Close()
	_ = f1.(httpgzip.GzipByter)

	_, err = io.CopyN(os.Stdout, f0, 9)
	if err != nil {
		panic(err)
	}
	_, err = io.CopyN(os.Stdout, f1, 9)
	if err != nil {
		panic(err)
	}

	// Output:
	// This fileThis file
}

func Example_readTwoOpenedUncompressedFiles() {
	var fs http.FileSystem = assets

	f0, err := fs.Open("/not-worth-compressing-file.txt")
	if err != nil {
		panic(err)
	}
	defer f0.Close()
	_ = f0.(httpgzip.NotWorthGzipCompressing)
	f1, err := fs.Open("/not-worth-compressing-file.txt")
	if err != nil {
		panic(err)
	}
	defer f1.Close()
	_ = f1.(httpgzip.NotWorthGzipCompressing)

	_, err = io.CopyN(os.Stdout, f0, 9)
	if err != nil {
		panic(err)
	}
	_, err = io.CopyN(os.Stdout, f1, 9)
	if err != nil {
		panic(err)
	}

	// Output:
	// Its normaIts norma
}

func Example_modTime() {
	var fs http.FileSystem = assets

	f, err := fs.Open("/sample-file.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}

	fmt.Println(fi.ModTime())

	// Output:
	// 0001-01-01 00:00:00 +0000 UTC
}

func Example_seek() {
	var fs http.FileSystem = assets

	f, err := fs.Open("/sample-file.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.CopyN(os.Stdout, f, 5)
	if err != nil {
		panic(err)
	}
	_, err = f.Seek(22, io.SeekCurrent)
	if err != nil {
		panic(err)
	}
	_, err = io.CopyN(os.Stdout, f, 10)
	if err != nil {
		panic(err)
	}
	fmt.Print("...")
	_, err = f.Seek(-4, io.SeekEnd)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		panic(err)
	}
	_, err = f.Seek(3, io.SeekStart)
	if err != nil {
		panic(err)
	}
	_, err = f.Seek(1, io.SeekCurrent)
	if err != nil {
		panic(err)
	}
	_, err = io.CopyN(os.Stdout, f, 22)
	if err != nil {
		panic(err)
	}

	// Output:
	// This Blaaaaaaaa...aah! file compresses well.
}

type fisStringer []os.FileInfo

func (fis fisStringer) String() string {
	var s = "[ "
	for _, fi := range fis {
		s += fi.Name() + " "
	}
	return s + "]"
}

func Example_seekDir1() {
	var fs http.FileSystem = assets

	f, err := fs.Open("/")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fis, err := f.Readdir(0)
	fmt.Println(fisStringer(fis), err)

	// Output:
	// [ folderA folderB not-worth-compressing-file.txt sample-file.txt ] <nil>
}

func Example_seekDir2() {
	var fs http.FileSystem = assets

	f, err := fs.Open("/")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fis, err := f.Readdir(2)
	fmt.Println(fisStringer(fis), err)
	fis, err = f.Readdir(1)
	fmt.Println(fisStringer(fis), err)
	_, err = f.Seek(0, io.SeekStart)
	fmt.Println(err)
	fis, err = f.Readdir(2)
	fmt.Println(fisStringer(fis), err)
	_, err = f.Seek(0, io.SeekStart)
	fmt.Println(err)
	fis, err = f.Readdir(1)
	fmt.Println(fisStringer(fis), err)
	fis, err = f.Readdir(10)
	fmt.Println(fisStringer(fis), err)
	fis, err = f.Readdir(10)
	fmt.Println(fisStringer(fis), err)

	// Output:
	// [ folderA folderB ] <nil>
	// [ not-worth-compressing-file.txt ] <nil>
	// <nil>
	// [ folderA folderB ] <nil>
	// <nil>
	// [ folderA ] <nil>
	// [ folderB not-worth-compressing-file.txt sample-file.txt ] <nil>
	// [ ] EOF
}

func Example_notExist() {
	var fs http.FileSystem = assets

	_, err := fs.Open("/does-not-exist")
	fmt.Println("os.IsNotExist:", os.IsNotExist(err))
	fmt.Println(err)

	// Output:
	// os.IsNotExist: true
	// open /does-not-exist: file does not exist
}

func Example_pathCleaned() {
	var fs http.FileSystem = assets

	f, err := fs.Open("//folderB/../folderA/file1.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}
	fmt.Println(fi.Name())

	b, err := ioutil.ReadAll(f)
	fmt.Printf("%q %v\n", string(b), err)

	// Output:
	// file1.txt
	// "Stuff in /folderA/file1.txt." <nil>
}
