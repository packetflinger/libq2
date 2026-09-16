package bsp

import (
	"fmt"
	"strconv"
	"strings"
)

// MapFile is the parsed contents of a Radiant-family .map source file: a
// flat list of entities, each optionally containing brushes built from
// arbitrary-plane faces. It's an intermediate representation - unlike
// BSPFile, it doesn't correspond to anything written to disk in binary
// form, and building it from an edited .map is only the first step
// towards recompiling a .bsp (turning this brush-plane soup into an
// actual BSP tree/leaves/visibility is a separate, much larger step).
type MapFile struct {
	Entities []*MapEntity
}

// MapEntity is one brace-delimited entity block: its key/value properties
// plus, for brush entities (including the world entity), the brushes
// defined inside it.
type MapEntity struct {
	Properties map[string]string
	Brushes    []*MapBrush
}

// MapBrush is one brace-delimited brush block: the convex volume formed
// by the intersection of all its faces' half-spaces.
type MapBrush struct {
	Faces []*MapBrushFace
}

// MapBrushFace is a single face of a brush, in Quake2 "Valve" format: 3
// points defining the face's plane, a material (texture) name, explicit
// texture axis vectors + offsets, cosmetic rotation/scale, and Quake 2's
// optional trailing contents/flags/value surface attributes.
type MapBrushFace struct {
	P0, P1, P2       vec3
	Material         string
	SAxis, TAxis     vec3
	SOffset, TOffset float64
	Rotation         float64
	ScaleX, ScaleY   float64

	// HasSurfaceAttribs reports whether the optional trailing "contents
	// flags value" trio was present; if false, Contents/Flags/Value are
	// zero rather than meaningfully absent.
	HasSurfaceAttribs      bool
	Contents, Flags, Value int32
}

// ParseMap parses the text of a Radiant-family .map source file. It
// accepts both the "standard" id Software face format (material xoff
// yoff rotation xscale yscale) and the "Valve"/TrenchBroom format used by
// Decompile (material [ sx sy sz soff ] [ tx ty tz toff ] rotation xscale
// yscale), with or without Quake 2's trailing contents/flags/value.
func ParseMap(data []byte) (*MapFile, error) {
	tok := newMapTokenizer(string(data))
	mf := &MapFile{}
	for {
		t, ok := tok.next()
		if !ok {
			break
		}
		if t != "{" {
			return nil, fmt.Errorf("line %d: expected '{' to start an entity, got %q", tok.line, t)
		}
		ent, err := parseMapEntity(tok)
		if err != nil {
			return nil, err
		}
		mf.Entities = append(mf.Entities, ent)
	}
	return mf, nil
}

func parseMapEntity(tok *mapTokenizer) (*MapEntity, error) {
	ent := &MapEntity{Properties: map[string]string{}}
	for {
		t, ok := tok.next()
		if !ok {
			return nil, fmt.Errorf("line %d: unexpected end of file inside entity", tok.line)
		}
		switch t {
		case "}":
			return ent, nil
		case "{":
			brush, err := parseMapBrush(tok)
			if err != nil {
				return nil, err
			}
			ent.Brushes = append(ent.Brushes, brush)
		default:
			val, ok := tok.next()
			if !ok {
				return nil, fmt.Errorf("line %d: missing value for property %q", tok.line, t)
			}
			ent.Properties[t] = val
		}
	}
}

func parseMapBrush(tok *mapTokenizer) (*MapBrush, error) {
	brush := &MapBrush{}
	for {
		t, ok := tok.peek()
		if !ok {
			return nil, fmt.Errorf("line %d: unexpected end of file inside brush", tok.line)
		}
		if t == "}" {
			tok.next()
			return brush, nil
		}
		face, err := parseMapFace(tok)
		if err != nil {
			return nil, err
		}
		brush.Faces = append(brush.Faces, face)
	}
}

func parseMapFace(tok *mapTokenizer) (*MapBrushFace, error) {
	p0, err := parsePoint(tok)
	if err != nil {
		return nil, err
	}
	p1, err := parsePoint(tok)
	if err != nil {
		return nil, err
	}
	p2, err := parsePoint(tok)
	if err != nil {
		return nil, err
	}

	material, ok := tok.next()
	if !ok {
		return nil, fmt.Errorf("line %d: expected a material name", tok.line)
	}

	face := &MapBrushFace{P0: p0, P1: p1, P2: p2, Material: material, ScaleX: 1, ScaleY: 1}

	next, ok := tok.peek()
	if !ok {
		return nil, fmt.Errorf("line %d: unexpected end of file after material name", tok.line)
	}
	if next == "[" {
		sAxis, sOffset, err := parseAxis(tok)
		if err != nil {
			return nil, err
		}
		tAxis, tOffset, err := parseAxis(tok)
		if err != nil {
			return nil, err
		}
		face.SAxis, face.SOffset = sAxis, sOffset
		face.TAxis, face.TOffset = tAxis, tOffset
	} else {
		xoff, err := parseFloatToken(tok)
		if err != nil {
			return nil, err
		}
		yoff, err := parseFloatToken(tok)
		if err != nil {
			return nil, err
		}
		face.SOffset, face.TOffset = xoff, yoff
	}

	if face.Rotation, err = parseFloatToken(tok); err != nil {
		return nil, err
	}
	if face.ScaleX, err = parseFloatToken(tok); err != nil {
		return nil, err
	}
	if face.ScaleY, err = parseFloatToken(tok); err != nil {
		return nil, err
	}

	// Quake 2's trailing "contents flags value" is optional; only consume
	// it if the next token actually parses as a number, so we don't eat
	// the '(' or '}' that would otherwise start the next face/brush.
	if peeked, ok := tok.peek(); ok {
		if _, err := strconv.ParseFloat(peeked, 64); err == nil {
			c, err := parseFloatToken(tok)
			if err != nil {
				return nil, err
			}
			f, err := parseFloatToken(tok)
			if err != nil {
				return nil, err
			}
			v, err := parseFloatToken(tok)
			if err != nil {
				return nil, err
			}
			face.Contents, face.Flags, face.Value = int32(c), int32(f), int32(v)
			face.HasSurfaceAttribs = true
		}
	}

	return face, nil
}

func parseFloatToken(tok *mapTokenizer) (float64, error) {
	t, ok := tok.next()
	if !ok {
		return 0, fmt.Errorf("line %d: unexpected end of file, expected a number", tok.line)
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, fmt.Errorf("line %d: expected a number, got %q", tok.line, t)
	}
	return f, nil
}

func parsePoint(tok *mapTokenizer) (vec3, error) {
	if t, ok := tok.next(); !ok || t != "(" {
		return vec3{}, fmt.Errorf("line %d: expected '(' to start a point, got %q", tok.line, t)
	}
	x, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, err
	}
	y, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, err
	}
	z, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, err
	}
	if t, ok := tok.next(); !ok || t != ")" {
		return vec3{}, fmt.Errorf("line %d: expected ')' to end a point, got %q", tok.line, t)
	}
	return vec3{x, y, z}, nil
}

func parseAxis(tok *mapTokenizer) (vec3, float64, error) {
	if t, ok := tok.next(); !ok || t != "[" {
		return vec3{}, 0, fmt.Errorf("line %d: expected '[' to start a texture axis, got %q", tok.line, t)
	}
	x, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, 0, err
	}
	y, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, 0, err
	}
	z, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, 0, err
	}
	off, err := parseFloatToken(tok)
	if err != nil {
		return vec3{}, 0, err
	}
	if t, ok := tok.next(); !ok || t != "]" {
		return vec3{}, 0, fmt.Errorf("line %d: expected ']' to end a texture axis, got %q", tok.line, t)
	}
	return vec3{x, y, z}, off, nil
}

// mapTokenizer splits .map source text into the tokens ParseMap consumes:
// the single-character brace/paren/bracket punctuation, double-quoted
// strings (with \" escapes), and otherwise whitespace-delimited bare
// words (numbers, and unquoted material names).
type mapTokenizer struct {
	data []rune
	pos  int
	line int
}

func newMapTokenizer(s string) *mapTokenizer {
	return &mapTokenizer{data: []rune(s), line: 1}
}

func (t *mapTokenizer) advance() rune {
	c := t.data[t.pos]
	t.pos++
	if c == '\n' {
		t.line++
	}
	return c
}

func (t *mapTokenizer) skipWhitespaceAndComments() {
	for t.pos < len(t.data) {
		c := t.data[t.pos]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			t.advance()
			continue
		}
		if c == '/' && t.pos+1 < len(t.data) && t.data[t.pos+1] == '/' {
			for t.pos < len(t.data) && t.data[t.pos] != '\n' {
				t.advance()
			}
			continue
		}
		break
	}
}

func isPunct(c rune) bool {
	switch c {
	case '{', '}', '(', ')', '[', ']':
		return true
	}
	return false
}

func (t *mapTokenizer) next() (string, bool) {
	t.skipWhitespaceAndComments()
	if t.pos >= len(t.data) {
		return "", false
	}

	c := t.data[t.pos]
	if isPunct(c) {
		t.advance()
		return string(c), true
	}

	if c == '"' {
		t.advance()
		var sb strings.Builder
		for t.pos < len(t.data) {
			c := t.data[t.pos]
			if c == '\\' && t.pos+1 < len(t.data) && t.data[t.pos+1] == '"' {
				sb.WriteRune('"')
				t.advance()
				t.advance()
				continue
			}
			if c == '"' {
				t.advance()
				break
			}
			sb.WriteRune(c)
			t.advance()
		}
		return sb.String(), true
	}

	start := t.pos
	for t.pos < len(t.data) {
		c := t.data[t.pos]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '"' || isPunct(c) {
			break
		}
		t.advance()
	}
	return string(t.data[start:t.pos]), true
}

// peek returns the next token without consuming it.
func (t *mapTokenizer) peek() (string, bool) {
	savedPos, savedLine := t.pos, t.line
	tok, ok := t.next()
	t.pos, t.line = savedPos, savedLine
	return tok, ok
}
