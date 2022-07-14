# TS Go

TS Go is a tool for generating typescript declarations from go ast nodes. All go types are supported with the exception of struct methods.

## Example
```go
// source.go

package main

// Primitive Types
const boolLit = true
const strLit = "hello world"
const runeLit = 'x'
const intLit = 178
const floatLit = 56.93

type boolType bool 
type strType string
type intType int
type floatType float64
type errorType error 
type anyType interface{} 

type ptrType *uint16 
type sliceType []int64
type arrayType [8]float32
type mapType map[float32]byte
type structType struct {
	A string
	B complex128
	unexported bool
}
type embeddedStructType struct {
	structType
	C float64
}
type genericStructType[T1 string, T2 any] struct {
	A T1
	B T2
	C string
}
type emptyInterfaceType interface {}
type interfaceType interface {
	MethodA(string, int) error
	MethodB(rune, int32)
}
type embeddedInterfaceType interface {
	interfaceType // ast.Ident
	MethodC() string
}
type onlyEmbeddedInterfaceType interface {
	interfaceType
	embeddedInterfaceType
}
type typeSetType interface {
	~string | int
	interfaceType
}

// Func types
func funcLit(A int8, B string) error { return nil }
func (recv *structType) methodLit(A uint32) {}
func multipleResults() (string, int, error) { return "", 0, nil }
func namedResults() (str string, err error) { return "", nil }
func genericFuncLit[T1 typeSetType, T2 ~int]() {}

// Nested types
type nestedMap map[string]struct {A, B int}
type nestedSlice []map[int]map[string][]any
type nestedArray [8][]uint64
type nestedStruct struct {
	A map[string]int
	B []complex64
}
type recursiveStruct struct {
	A float64
	B onlyEmbeddedInterfaceType
	C *recursiveStruct
}
```

```ts
// src.d.ts

declare const boolLit = true;
declare const strLit = "hello world";
declare const runeLit = "x";
declare const intLit = 178;
declare const floatLit = 56.93;

declare type boolType = boolean;
declare type strType = string;
declare type intType = number;
declare type floatType = number;
declare type errorType = Error;
declare type anyType = any;

declare type ptrType = number | undefined;
declare type sliceType = number[];
declare type arrayType = number[];
declare type mapType = { [key: number]: number };
declare type structType = { A: string; B: number; unexported: boolean; };

declare type embeddedStructType = { C: number; } & (structType);
declare type genericStructType<T1 extends string, T2 extends any> = { A: T1; B: T2; C: string; };
declare type emptyInterfaceType = any;
declare type interfaceType = { MethodA: (_: string, __: number) => Error; MethodB: (_: string, __: number) => void; };
declare type embeddedInterfaceType = { MethodC: () => string; } & (interfaceType);
declare type onlyEmbeddedInterfaceType = { } & (interfaceType) & (embeddedInterfaceType);
declare type typeSetType = { } & (string | number) & (interfaceType);

declare const funcLit: (A: number, B: string) => Error;
declare const methodLit: (A: number) => void;
declare const multipleResults: () => [string, number, Error];
declare const namedResults: () => { str: string, err: Error };
declare const genericFuncLit: ()<T1 extends typeSetType, T2 extends number> => void;

declare type nestedMap = { [key: string]: { A: number; B: number; } };
declare type nestedSlice = { [key: number]: { [key: string]: any[] } }[];
declare type nestedArray = number[][];
declare type nestedStruct = { A: { [key: string]: number }; B: number[]; };
declare type recursiveStruct = { A: number; B: onlyEmbeddedInterfaceType; C: recursiveStruct | undefined; };
```

## Usage
```go
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"github.com/baldwin-dev-co/ts-go"
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
```
