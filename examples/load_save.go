package main

import (
	"os"
	"github.com/vtphan/fmic"
	"github.com/vtphan/fmi"
	"fmt"
)

//-----------------------------------------------------------------------------
func main() {
	if len(os.Args) != 2 {
		panic("Usage: go run program.go file.fasta")
	}
	idx := fmic.CompressedIndex(os.Args[1], 10)
	idx.Show()
	idx.Check()
	fmt.Println("======SAVING INDEX")
	idx.SaveCompressedIndex(2)

	fmt.Println("======RELOADING INDEX")
	saved_idx := fmic.LoadCompressedIndex(os.Args[1] + ".fmi")
	saved_idx.Show()
	saved_idx.Check()

	fmt.Println("======UNCOMPRESSED INDEX")
	uncompressed_idx := fmi.New(os.Args[1])
	uncompressed_idx.Show()
	uncompressed_idx.Check()
}
