package main

import (
	_ "embed"
	"go/build"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed tmpls/index.tmpl
var indexTmpl string

//go:embed tmpls/wasm.tmpl
var wasmTmpl string

func main() {
	if err := createWasmHTML("main.wasm"); err != nil {
		log.Fatal(err)
	}
	if err := createIndexHTML(); err != nil {
		log.Fatal(err)
	}
}

func createWasmHTML(defWASMFile string) error {
	const target = "./docs/wasm.html"

	wasmExecName := filepath.Join(build.Default.GOROOT, "misc/wasm/wasm_exec.js")
	// Read wasm_exec from system dist
	wasmExec, err := ioutil.ReadFile(wasmExecName)
	if err != nil {
		panic(err)
	}
	t, err := template.New("wasm").Parse(wasmTmpl)
	if err != nil {
		panic(err)
	}

	topts := map[string]interface{}{
		"wasmexec": string(wasmExec),
		"wasmfile": defWASMFile,
	}

	log.Println("Creating target:", target)
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close() // nolint: errcheck
	return t.Execute(f, topts)
}

func createIndexHTML() error {
	const target = "./docs/index.html"
	var files []string
	err := fs.WalkDir(os.DirFS("./docs"), "wasm", func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			return nil
		}
		log.Println("Adding:", p)
		files = append(files, p)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("").Parse(indexTmpl)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating target:", target)
	f, err := os.Create(target)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	return t.Execute(f, files)
}
