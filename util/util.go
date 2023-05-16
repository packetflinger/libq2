package util

func VectorCompare(v1 [3]int16, v2 [3]int16) bool {
	return (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2])
}

func VectorCompare8(v1 [3]int8, v2 [3]int8) bool {
	return (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2])
}

func Vector4Compare8(v1 [4]uint8, v2 [4]uint8) bool {
	return (v1[0] == v2[0]) && (v1[1] == v2[1]) && (v1[2] == v2[2]) && (v1[3] == v2[3])
}
