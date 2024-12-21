// mvd2text is for parsing MVD2 demos in their text format
package demo

import (
	"errors"
	"os"
	"strings"
)

type DemoFile struct {
	Blocks []Block
}

type ServerData struct {
	MajorVersion string
	MinorVersion string
	ServerCount  int
	GameDir      string
	ClientNum    int
	NoMessages   bool
	BaseStrings  []ConfigString
}

type ConfigString struct {
	Index  int
	String string
}

type BaseFrame struct {
	PortalBits int
	Players    []Player
	Entities   []Entity
}

type Player struct {
	Number       int
	OriginXY     [2]int
	OriginZ      int
	ViewOffset   [3]int
	ViewAnglesXY [2]int
	KickAngles   [3]int
	WeaponIndex  int
	WeaponFrame  int
	FOV          int
	Stats        []Stat
}

type Stat struct {
	Index int
	Value int
}

type Block struct {
	Frame BaseFrame
}

type Entity struct {
	Number      int
	ModelIndex  int
	ModelIndex2 int
	Frame       int
	Skin        int
	Solid       int
	OriginX     int
	OriginY     int
	OriginZ     int
	RenderFX    int // hex
	Sound       int
	Effects     int // hex
	AngleX      int
	AngleY      int
	AngleZ      int
}

func ParseMVD2TextDemo(demofile string) (DemoFile, error) {
	demo := DemoFile{}
	data, err := os.ReadFile(demofile)
	if err != nil {
		return demo, err
	}
	if string(data[0:4]) != "TXT2" {
		return demo, errors.New("invalid textdemo")
	}

	lines := strings.Split(string(data[5:]), "\n")
	for _, line := range lines {
		line := strings.Trim(line, " ")
		t := strings.Split(strings.Trim(line, " "), " ")
		if t[0] == "block" {
			//inBlock = true
			continue
		}
		/*
			if t[0] == "serverdata" {

			}
		*/
	}
	return demo, nil
}
