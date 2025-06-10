package parser

import (
	"fmt"
	"os"

	"github.com/Maurits825/satisfactory-savefile-parser/internal/readsave"
	"github.com/Maurits825/satisfactory-savefile-parser/pkg/saveformat"
)

func ParseSaveFile(saveFileName string) *saveformat.SaveFileBody {
	file, err := os.Open(saveFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer file.Close()

	body := readsave.ReadSave(file)
	return body
}
