package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	diffFile, err := ioutil.ReadFile("./gen/diff.js")
	if err != nil {
		log.Printf("ERROR: cannt read gen/diff.js: %v", err)
		os.Exit(1)
	}
	diffV2File, err := ioutil.ReadFile("./gen/diff_v2.js")
	if err != nil {
		log.Printf("ERROR: cannt read gen/diff_v2.js: %v", err)
		os.Exit(1)
	}

	diffGojaFile, err := ioutil.ReadFile("./gen/diff_goja.js")
	if err != nil {
		log.Printf("ERROR: cannt read gen/diff_goja.js: %v", err)
		os.Exit(1)
	}
	gojaPolyfill := string(diffGojaFile) + "globalThis.run();"
	content := fmt.Sprintf("// Code generated by go-coverage;DO NOT EDIT.\n\npackage vscode\n\nvar diffJSCode = %q\nvar diffV2JSCode = %q\nvar DiffGojaCode = %q\n", string(diffFile), string(diffV2File), gojaPolyfill)
	err = ioutil.WriteFile("./diff_js_gen.go", []byte(content), 0777)
	if err != nil {
		log.Printf("ERROR: cannt write diff_js_gen.go: %v", err)
		os.Exit(1)
	}
}
