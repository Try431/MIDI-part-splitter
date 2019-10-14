package main

import (
	"strconv"
	"syscall/js"
)

// func add(i []js.Value) {
// 	js.Global().Set("output", js.ValueOf(i[0].Int()+i[1].Int()))
// 	println(js.ValueOf(i[0].Int() + i[1].Int()).String())
// }

// function definition
func add(this js.Value, i []js.Value) interface{} {
	value1 := js.Global().Get("document").Call("getElementById", i[0].String()).Get("value").String()
	value2 := js.Global().Get("document").Call("getElementById", i[1].String()).Get("value").String()

	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)
	sum := int1 + int2

	js.Global().Get("document").Call("getElementById", i[2].String()).Set("value", sum)
	return nil
}

// func subtract(i []js.Value) {
// 	js.Global().Set("output", js.ValueOf(i[0].Int()-i[1].Int()))
// 	println(js.ValueOf(i[0].Int() - i[1].Int()).String())
// }

// function definition
func subtract(this js.Value, i []js.Value) interface{} {
	value1 := js.Global().Get("document").Call("getElementById", i[0].String()).Get("value").String()
	value2 := js.Global().Get("document").Call("getElementById", i[1].String()).Get("value").String()

	int1, _ := strconv.Atoi(value1)
	int2, _ := strconv.Atoi(value2)
	diff := int1 - int2

	js.Global().Get("document").Call("getElementById", i[2].String()).Set("value", diff)
	return nil
}

func registerFunctions() {
	// exposing to JS
	js.Global().Set("add", js.FuncOf(add))
	js.Global().Set("subtract", js.FuncOf(subtract))

}

func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerFunctions()
	<-c
}
