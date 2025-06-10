package main

import (
	"path/filepath"
	"testing"
)

func testReadSaveFile(file string, t *testing.T) {
	saveFile := filepath.Join("testdata", file)
	body := readSaveFile(saveFile)
	if body == nil {
		t.Error(file, "Body is nil")
		return
	}

	if len(body.Levels) == 0 {
		t.Error(file, "Body empty levels")
	}
	if body.Zero != 0 {
		t.Error(file, "Body zero field is not zero")
	}
	if body.NoneString1 != "None" || body.NoneString2 != "None" {
		t.Error(file, "Body noneString field is not None")
	}
}

func TestReadSaveFile(t *testing.T) {
	saveFiles := []string{
		"test_creative_v1.1_exp.sav",
	}

	for _, save := range saveFiles {
		testReadSaveFile(save, t)
	}
}
