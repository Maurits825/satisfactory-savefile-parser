package main

import (
	"fmt"
	"slices"
)

type sloopInfo struct {
	name     string
	count    int32
	location Vector
}

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

	body := readSaveFile(saveFile)
	sloops := findSloops(body)

	fmt.Println("Sloops found:")
	for _, s := range sloops {
		fmt.Printf("Name: %s, count: %d\n", s.name, s.count)
	}
}

func findSloops(body *SaveFileBody) []sloopInfo {
	fmt.Println("Finding sloops...")
	//only need to check persitent level??
	persistentLevel := body.Levels[len(body.Levels)-1]

	inventoryRefs := make([]string, 0)
	for i, actorHeader := range persistentLevel.ActorHeaders {
		if actorHeader.TypePath == "/Game/FactoryGame/Buildable/Factory/QuantumEncoder/Build_QuantumEncoder.Build_QuantumEncoder_C" {
			for _, p := range persistentLevel.ActorObjects[i].Properties {
				if p.Name == "mInventoryPotential" {
					//todo store index also to get actor head/obj
					inventoryRefs = append(inventoryRefs, p.Value.(ObjectReference).PathName)
				}
			}
		}
	}

	sloops := make([]sloopInfo, 0)
	for i, componentHeader := range persistentLevel.ComponentHeaders {
		if slices.Contains(inventoryRefs, componentHeader.Name) {
			for _, p := range persistentLevel.ComponentObjects[i].Properties {
				if p.Name == "mInventoryStacks" {
					inventStacks := p.Value.(ArrayStructProperty)
					for _, item := range inventStacks.Value {
						//is the order the same always?
						itemName := item.([]Property)[0].Value.(InventoryItem).Reference.PathName
						if itemName == "/Game/FactoryGame/Prototype/WAT/Desc_WAT1.Desc_WAT1_C" {
							sloopCount := item.([]Property)[1].Value.(int32)
							//todo for location we have to find the actor header?
							s := sloopInfo{name: componentHeader.Name, count: sloopCount}
							sloops = append(sloops, s)
						}
					}
				}
			}
			// fmt.Printf("Header: %+v\n", persistentLevel.ComponentHeaders[i])
			// fmt.Printf("Component props: %+v\n", persistentLevel.ComponentObjects[i].Properties)
			// fmt.Println("")
		}
	}
	return sloops
}
