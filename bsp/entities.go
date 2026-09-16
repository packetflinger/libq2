package bsp

import (
	"bytes"
	"strings"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// parseEntities decodes the entity lump, a single NUL-terminated text blob
// containing brace-delimited blocks of quoted key/value pairs, ex:
//
//	{
//	"classname" "info_player_start"
//	"origin" "64 312 408"
//	}
func parseEntities(data []byte) ([]*bpb.BSPEntity, error) {
	if i := bytes.IndexByte(data, 0); i >= 0 {
		data = data[:i]
	}

	var ents []*bpb.BSPEntity
	var props map[string]string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "{":
			props = map[string]string{}
		case "}":
			if props != nil {
				ents = append(ents, &bpb.BSPEntity{
					ClassName:  props["classname"],
					Properties: props,
				})
				props = nil
			}
		default:
			if props == nil {
				continue
			}
			if key, val, ok := parseEntityLine(line); ok {
				props[key] = val
			}
		}
	}
	return ents, nil
}

// parseEntityLine splits a `"key" "value"` line into its two quoted tokens.
func parseEntityLine(line string) (key, val string, ok bool) {
	tokens := strings.SplitN(line, " ", 2)
	if len(tokens) < 2 {
		return "", "", false
	}
	k, v := tokens[0], tokens[1]
	if len(k) < 2 || k[0] != '"' || k[len(k)-1] != '"' {
		return "", "", false
	}
	if len(v) < 2 || v[0] != '"' || v[len(v)-1] != '"' {
		return "", "", false
	}
	return k[1 : len(k)-1], v[1 : len(v)-1], true
}

// marshalEntities encodes entities back into the brace-delimited text blob
// format, terminated with a trailing NUL byte. Property order within a
// block isn't preserved (they're stored in a map), but classname is always
// written first for readability.
func marshalEntities(ents []*bpb.BSPEntity) []byte {
	var buf bytes.Buffer
	for _, ent := range ents {
		buf.WriteString("{\n")
		if cn, ok := ent.GetProperties()["classname"]; ok {
			buf.WriteString("\"classname\" \"" + cn + "\"\n")
		}
		for k, v := range ent.GetProperties() {
			if k == "classname" {
				continue
			}
			buf.WriteString("\"" + k + "\" \"" + v + "\"\n")
		}
		buf.WriteString("}\n")
	}
	buf.WriteByte(0)
	return buf.Bytes()
}
