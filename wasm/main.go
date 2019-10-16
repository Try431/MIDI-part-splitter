package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func setupRoutes() {
	http.HandleFunc("/uploaded", uploadFile)
	fmt.Println("setup routes")
	http.ListenAndServe(":8080", nil)
}

func main() {
	// c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerFunctions()
	setupRoutes()
	// <-c
}
