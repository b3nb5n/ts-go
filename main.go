package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"github.com/baldwin-dev-co/ts-go/tsgo"
)

func main() {
	// read file
	file, err := os.Open("test.go")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// read the whole file in
	srcbuf, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	src := string(srcbuf)

	// file set
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "lib.go", src, 0)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		return
	}

	s := new(strings.Builder)
	err = tsgo.WriteFileDecls(s, node)
	fmt.Println(s.String())
	if err != nil {
		fmt.Printf("Error writing file declarations: %v\n", err)
		return
	}
}
