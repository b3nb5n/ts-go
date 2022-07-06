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
type anyType any

// Container types
// var ptrLit *string = nil
// var sliceLit = []string{}
// var arrayLit = [8]int{}
// var mapLit = map[string]float64{}
// var structLit = struct { A rune; B bool } {}

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
// type emptyInterfaceType interface {}
// type interfaceType interface {
// 	MethodA(string, int) error
// 	MethodB(rune, int32)
// }
// type embeddedInterfaceType interface {
// 	interfaceType
// 	MethodC() string
// }
// type typeSetType interface {
// 	string | int
// }

// // Func types
func funcLit(A int8, B string) error { return nil }
func (recv *structType) methodLit(A uint32) {}
func multipleResults() (string, int, error) { return "", 0, nil }
func namedResults() (str string, err error) { return "", nil }
// func genericFuncLit[T1 typeSetType, T2 ~int]() {}

// Nested types
type nestedMap map[string]struct {A, B int}
type nestedSlice []map[int]map[string][]any
type nestedArray [8][]uint64
type nestedStruct struct {
	A map[string]int
	B []complex64
}
