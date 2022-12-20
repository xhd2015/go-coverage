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
	content := fmt.Sprintf("// Code generated by go-coverage;DO NOT EDIT.\n\npackage vscode\n\nvar diffJSCode = %q\n", string(diffFile))
	err = ioutil.WriteFile("./diff_js_gen.go", []byte(content), 0777)
	if err != nil {
		log.Printf("ERROR: cannt write diff_js_gen.go: %v", err)
		os.Exit(1)
	}
}
