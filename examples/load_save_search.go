package main

import (
	"fmt"
	"github.com/vtphan/fmi"
	"github.com/vtphan/fmic"
	"math/rand"
	"os"
	"runtime"
)

//-----------------------------------------------------------------------------
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) != 2 {
		panic("Usage: go run program.go file.fasta")
	}
	fmt.Println("======BUILDING INDEX")
	idx := fmic.CompressedIndex(os.Args[1], true, 1)
	idx.Show()
	idx.Check()

	fmt.Println("======SAVING INDEX")
	idx.SaveCompressedIndex(2)
	fmt.Println(">>>", idx.LENS, idx.GENOME_ID)
	fmt.Println("======RELOADING INDEX")
	saved_idx := fmic.LoadCompressedIndex(os.Args[1] + ".fmi")
	saved_idx.Show()
	saved_idx.Check()

	fmt.Println("======Uncompressed INDEX")
	uncompressed_idx := fmi.New(os.Args[1])
	uncompressed_idx.Show()

	fmt.Println("======TEST SEARCH")
	var x, y, x1, y1 int
	for i := 0; i < 100000; i++ {
		a := rand.Int63n(int64(saved_idx.LEN))
		b := rand.Int63n(int64(saved_idx.LEN))
		if a != b {
			if a > b {
				a, b = b, a
			}
			// fmt.Printf("%d %d %d ", i, a, b)
			seq := fmi.SEQ[a:b]
			x, y = saved_idx.Search(seq)
			x1, y1, _ = uncompressed_idx.Search(seq)
			// fmt.Println(x,y,z, x==x1, y==y1, z==z1)
			if x != x1 || y != y1 {
				fmt.Println("Panic:", i, a, b, "\t", x, saved_idx.SA[x], y, "\t", x1, uncompressed_idx.SA[x1], y1, string(seq))
				panic("Something is wrong")
			}
			if i%10000 == 0 {
				fmt.Println("finish testing", i, "random substring searches.")
			}
		}
	}
}
