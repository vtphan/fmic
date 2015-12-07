package main

import (
	"github.com/vtphan/fmic"
	"fmt"
)

//-----------------------------------------------------------------------------
func main() {
	idx := fmic.CompressedIndex("seq1.fasta", true, 10)
	idx.Show()
	fmt.Println("======SAVING INDEX (sa and seq are not saved)")
	idx.SaveCompressedIndex(0)

	fmt.Println("======RELOADING INDEX")
	saved_idx := fmic.LoadCompressedIndex("seq1.fasta.fmi")
	saved_idx.Show()

	queries := []string {
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
	fmt.Println("seq_id\tmatches\ti\tquery")
	for _, q := range(queries) {
		seq, count, j := saved_idx.GuessSequence([]byte(q))
		fmt.Println(seq,"\t",count,"\t",j,"\t",q)
		// seq, count, j = idx.GuessSequence([]byte(q))
		// fmt.Println(seq,"\t",count,"\t",j,"\t",q)
	}
	fmt.Println()
}