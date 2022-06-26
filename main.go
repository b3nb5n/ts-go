package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/baldwin-dev-co/ts-go/tsgo"
)

func main() {
	// read file
	file, err := os.Open("test.go")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	// read the whole file in
	srcbuf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return
	}
	src := string(srcbuf)

	// file set
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "lib.go", src, 0)
	if err != nil {
		log.Println(err)
		return
	}

	gen := tsgo.NewGenerator(nil)
	gen.ParseAstNode(node)
	fmt.Println(gen.String())
}
