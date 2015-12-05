package main

import (
	"github.com/vtphan/fmic"
	"fmt"
	"os"
)

//-----------------------------------------------------------------------------
func main() {
	if len(os.Args) != 2 {
		panic("Usage: go run program.go file.fasta")
	}
	idx := fmic.CompressedIndex(os.Args[1], 10)
	idx.Show()
	queries := []string {
		"thisisthe",
		"sequence",
		"firstsequence",
		"secondsequence",
		"thirdsequence",
		"forthsequence",
		"dynamicprogramming",
		"recursion",
		"iterative",
		"ohlala",
		"partialmatchfo",
	}
	fmt.Println("region\tmatches\ti\tquery")
	for _, q := range(queries) {
		region, count, j := idx.SearchRegion([]byte(q))
		fmt.Println(region,"\t",count,"\t",j,"\t",q)
	}
	fmt.Println()
}