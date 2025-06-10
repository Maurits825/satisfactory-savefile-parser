package main

import (
	"fmt"

	"github.com/Maurits825/satisfactory-savefile-parser/pkg/parser"
)

func main() {
	saveFile := "pkg/parser/testdata/test_benchmark.sav"
	body := parser.ParseSaveFile(saveFile)
	fmt.Println("Save file parsed successfully!", body.UncompressedSize)
}
