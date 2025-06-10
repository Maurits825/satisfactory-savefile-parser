package main

import (
	"fmt"

	"github.com/Maurits825/satisfactory-savefile-parser/pkg/parser"
)

func main() {
	//test
	//52-14
	saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\test_creative_v1.1_exp_040625-171915.sav"

	//working
	//46-13
	// saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\Leggo v1.sav"
	// saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\Leggov1.1.sav"

	//46-13
	// saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\Leggo v1_continue.sav"
	//52-14
	// saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\Here we go again.sav"
	//51-14
	// saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\Leggo_autosave_1.sav"

	//>=44 for inventory item to work -> format is different (only a objProp)
	//42-13
	// saveFile := "C:\\Users\\Maurits\\AppData\\Local\\FactoryGame\\Saved\\SaveGames\\76561198083442458\\Leggo_300624-011310.sav"

	body := parser.ParseSaveFile(saveFile)
	fmt.Println("Save file parsed successfully!", body.UncompressedSize)
}
