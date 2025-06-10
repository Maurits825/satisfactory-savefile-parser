package parser_test

import (
	"testing"

	. "github.com/Maurits825/satisfactory-savefile-parser/pkg/parser"
)

func BenchmarkParser(b *testing.B) {
	save := "testdata/test_benchmark.sav"
	// save := "testdata/test_creative_v1.1_exp.sav"
	b.Run(save, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ParseSaveFile(save)
		}
	})
}
