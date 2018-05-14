// Copyright (c) 2017 Ernest Micklei
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"io"
	"os"

	"flag"

	"bytes"
	"io/ioutil"

	"github.com/emicklei/proto"
)

var (
	oOverwrite = flag.Bool("w", false, "write result to (source) file instead of stdout")
	oDebug     = flag.Bool("d", false, "debug mode")
)

// go run *.go unformatted.proto
func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(0)
	}
	for _, each := range flag.Args() {
		if err := readFormatWrite(each); err != nil {
			println(each, err.Error())
		}
	}
}

func readFormatWrite(filename string) error {
	// open for read
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	// buffer before write
	buf := new(bytes.Buffer)
	if err := format(file, buf); err != nil {
		return err
	}
	if *oOverwrite {
		// write back to input
		if err := ioutil.WriteFile(filename, buf.Bytes(), os.ModePerm); err != nil {
			return err
		}
	} else {
		// write to stdout
		if _, err := io.Copy(os.Stdout, bytes.NewReader(buf.Bytes())); err != nil {
			return err
		}
	}
	return nil
}

func format(input io.Reader, output io.Writer) error {
	parser := proto.NewParser(input)
	def, err := parser.Parse()
	if err != nil {
		return err
	}
	// if *oDebug {
	// 	spew.Dump(def)
	// }
	proto.NewFormatter(output, "  ").Format(def) // 2 spaces
	return nil
}
