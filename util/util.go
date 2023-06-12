package util

import (
	"bufio"
	"os"
	"strings"
)

func VectorCompare(v1 [3]int16, v2 [3]int16) bool {
	return (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2])
}

func VectorCompare8(v1 [3]int8, v2 [3]int8) bool {
	return (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2])
}

func Vector4Compare8(v1 [4]uint8, v2 [4]uint8) bool {
	return (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2]) && (v1[3] == v2[3])
}

func FileExists(f string) bool {
	info, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Remove any duplipcates
func Deduplicate(in []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range in {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func SplitLines(str string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(str))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

// Make sure val is in between lower and upper
func Clamp(val int, lower int, upper int) int {
	if val < lower {
		return lower
	}
	if val > upper {
		return upper
	}
	return val
}

// Change consolechars back to normal text
func ConvertHighChars(in string) string {
	runes := []rune{}
	for _, chr := range in {
		runes = append(runes, chr&0x7f)
	}
	return string(runes)
}

// Change normal text to console chars
func ConvertLowChars(in string) string {
	runes := []rune{}
	for _, chr := range in {
		runes = append(runes, chr^0x80)
	}
	return string(runes)
}
