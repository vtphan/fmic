package main

import (
	"fmt"
	"github.com/vtphan/fmi"
	"github.com/vtphan/fmic"
	"os"
	"runtime"
	"math/rand"
)

//-----------------------------------------------------------------------------
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if len(os.Args) != 2 {
		panic("Usage: go run program.go file.fasta")
	}
	fmt.Println("======BUILDING INDEX")
	idx := fmic.CompressedIndex(os.Args[1], 10)
	idx.Show()
	idx.Check()

	fmt.Println("======SAVING INDEX")
	idx.SaveCompressedIndex(0)

	fmt.Println("======RELOADING INDEX")
	saved_idx := fmic.LoadCompressedIndex(os.Args[1] + ".fmi")
	saved_idx.Show()
	saved_idx.Check()

	fmt.Println("======TEST SEARCH")
	uncompressed_idx := fmi.New(os.Args[1])
	var x,y,z,x1,y1,z1 int
	for i:=0; i<100000; i++ {
		a := rand.Int63n(int64(saved_idx.LEN))
		b := rand.Int63n(int64(saved_idx.LEN))
		if a!=b {
			if a > b {
				a, b = b, a
			}
			// fmt.Printf("%d %d %d ", i, a, b)
			seq := fmi.SEQ[a:b]
			x,y,z = saved_idx.Search(seq)
			x1,y1,z1 = uncompressed_idx.Search(seq)
			// fmt.Println(x,y,z, x==x1, y==y1, z==z1)
			if x!=x1 || y!=y1 || z!=z1 {
				fmt.Println("Panic:", i, a, b, x,y,z, x1,y1,z1)
				panic("Something is wrong")
			}
			if i%10000 == 0 {
				fmt.Println("finish testing", i, "random substring searches.")
			}
		}
	}
}
