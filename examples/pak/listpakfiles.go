// Example program listing all the files in a pak file
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/packetflinger/libq2/pak"
)

var (
	pakfile = flag.String("pak_file", "", "The PAK archive to list file in")
)

func main() {
	flag.Parse()
	if *pakfile == "" {
		flag.PrintDefaults()
		return
	}
	data, err := os.ReadFile(*pakfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	archive, err := pak.Unmarshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("PAK files:")
	for _, f := range archive.GetFiles() {
		fmt.Println(" ", f.GetName())
	}
}
