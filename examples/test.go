package main

import (
	"fmt"
	"github.com/vtphan/fmic"
	// "math/rand"
	"runtime"
)

//-----------------------------------------------------------------------------
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("======BUILDING INDEX")
	idx := fmic.CompressedIndex("seq0.fasta", true, 1)
	idx.Show()
	idx.Check()

	// fmt.Println("======SAVING INDEX")
	// idx.SaveCompressedIndex(2)
	// fmt.Println("======RELOADING INDEX")
	// saved_idx := fmic.LoadCompressedIndex(os.Args[1] + ".fmi")
	// saved_idx.Show()
	// saved_idx.Check()
	// for i := 0; i < len(saved_idx.GENOME_ID); i++ {
	// 	fmt.Println(i, saved_idx.LENS[i], saved_idx.GENOME_ID[i])
	// }
}
