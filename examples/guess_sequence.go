package main

import (
	"fmt"
	"github.com/vtphan/fmic"
	"math/rand"
	"time"
)

//-----------------------------------------------------------------------------
func main() {
	rand.Seed(time.Now().UnixNano())

	idx := fmic.CompressedIndex("seq1.fasta", true, 10)
	// idx.Show()
	fmt.Println("======SAVING INDEX (sa and seq are not saved)")
	idx.SaveCompressedIndex(0)

	fmt.Println("======RELOADING INDEX")
	saved_idx := fmic.LoadCompressedIndex("seq1.fasta.fmi")
	// saved_idx.Show()

	queries := []string{
		"thisisthe",
		"sequence",
		"firstsequence",
		"secondsequence",
		"thirdsequence",
		"forthsequence",
		"recursion",
		"iterative",
		"dynamicprogramming",
		"ohlala",
		"partialmatchfo",
	}
	fmt.Println("seqId\tmatches\tquery")
	for _, q := range queries {
		seq, count := saved_idx.Guess([]byte(q), 2)
		fmt.Println(seq, "\t", count, "\t", q)
		// seq, count, j = idx.Guess([]byte(q))
		// fmt.Println(seq,"\t",count,"\t",j,"\t",q)
	}
	fmt.Println()
}
