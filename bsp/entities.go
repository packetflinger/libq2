package bsp

import (
	"fmt"
	"strings"

	u "github.com/packetflinger/libq2/util"
)

type BSPEntity struct {
	Class  string
	Values map[string]string
}

func (bsp *BSPFile) FetchEntities() []BSPEntity {
	ents := []BSPEntity{}
	ent := BSPEntity{}
	kvmap := make(map[string]string)

	lines := u.SplitLines(bsp.FetchEntityString())
	for _, line := range lines {
		if line == "{" {
			ent = BSPEntity{}
			kvmap = map[string]string{}
			continue
		}
		if line == "}" {
			ent.Values = kvmap
			ents = append(ents, ent)
			continue
		}

		tokens := strings.SplitN(line, " ", 2)
		if len(tokens) < 2 {
			continue
		}
		key := strings.ToLower(tokens[0][1 : len(tokens[0])-1])
		val := strings.ToLower(tokens[1][1 : len(tokens[1])-1])
		kvmap[key] = val
		if key == "classname" {
			ent.Class = val
		}
	}
	return ents
}

func (bsp *BSPFile) FetchEntityString() string {
	return string(bsp.LumpData[EntityLump].Data.Buffer)
}

func (bsp *BSPFile) BuildEntityString() string {
	buf := ""
	for _, ent := range bsp.Ents {
		buf += "{\n"
		for k, v := range ent.Values {
			buf += fmt.Sprintf("\"%s\" \"%s\"\n", k, v)
		}
		buf += "}\n"
	}
	return buf
}
