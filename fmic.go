/*
   Copyright 2015 Vinhthuy Phan
	Compressed FM index.
*/
package fmic

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
)

//-----------------------------------------------------------------------------
// Global variables: sequence (SEQ), suffix array (SA), BWT, FM index (C, OCC)
//-----------------------------------------------------------------------------

type IndexC struct {
	SEQ []byte
	BWT []byte
	SA  []indexType          // suffix array
	SSA []sequenceType       // SSA[i] stores the sequence containing position SA[i]
	C   map[byte]indexType   // count table
	OCC map[byte][]indexType // occurence table

	END_POS indexType          // position of "$" in the text
	SYMBOLS []int              // sorted symbols
	EP      map[byte]indexType // ending row/position of each symbol

	LEN        indexType
	OCC_SIZE   indexType
	Freq       map[byte]indexType // Frequency of each symbol
	M          int                // Compression ratio
	Multiple   bool               // True if the input contains multiple sequences
	input_file string
}

//-----------------------------------------------------------------------------
// Build FM index given the file storing the text.
// multiple is true if the input file contains multiple sequences
// compression ratio >=1
//-----------------------------------------------------------------------------
func CompressedIndex(file string, multiple bool, compression_ratio int) *IndexC {
	I := new(IndexC)
	I.input_file = file
	I.M = compression_ratio
	I.Multiple = multiple

	// GET THE SEQUENCE
	I.ReadFasta(file)

	// BUILD SUFFIX ARRAY
	I.LEN = indexType(len(I.SEQ))
	I.OCC_SIZE = indexType(math.Ceil(float64(I.LEN/indexType(I.M)))) + 1
	I.SA = make([]indexType, I.LEN)
	var SID []sequenceType
	if I.Multiple {
		I.SSA = make([]sequenceType, I.LEN)
		SID = make([]sequenceType, I.LEN)
	}
	SA := make([]int, I.LEN)
	ws := &WorkSpace{}
	ws.ComputeSuffixArray(I.SEQ, SA)
	sid := sequenceType(0)
	for i := range SA {
		I.SA[i] = indexType(SA[i])
		if I.Multiple {
			SID[i] = sid
			if I.SEQ[i] == '|' {
				sid++
			}
		}
	}

	// BUILD BWT
	I.Freq = make(map[byte]indexType)
	I.BWT = make([]byte, I.LEN)
	var i indexType
	for i = 0; i < I.LEN; i++ {
		I.Freq[I.SEQ[i]]++
		if I.SA[i] == 0 {
			I.BWT[i] = I.SEQ[I.LEN-1]
		} else {
			I.BWT[i] = I.SEQ[I.SA[i]-1]
		}
		if I.BWT[i] == '$' {
			I.END_POS = i
		}
		if I.Multiple {
			I.SSA[i] = SID[I.SA[i]]
		}
	}

	// BUILD COUNT AND OCCURENCE TABLE
	I.C = make(map[byte]indexType)
	I.OCC = make(map[byte][]indexType)
	for c := range I.Freq {
		I.SYMBOLS = append(I.SYMBOLS, int(c))
		I.OCC[c] = make([]indexType, I.OCC_SIZE)
		I.C[c] = 0
	}
	sort.Ints(I.SYMBOLS)
	I.EP = make(map[byte]indexType)
	count := make(map[byte]indexType)

	for j := 1; j < len(I.SYMBOLS); j++ {
		curr_c, prev_c := byte(I.SYMBOLS[j]), byte(I.SYMBOLS[j-1])
		I.C[curr_c] = I.C[prev_c] + I.Freq[prev_c]
		I.EP[curr_c] = I.C[curr_c] + I.Freq[curr_c] - 1
		count[curr_c] = 0
	}

	for j := 0; j < len(I.BWT); j++ {
		count[I.BWT[j]] += 1
		if j%I.M == 0 {
			for symbol := range I.OCC {
				I.OCC[symbol][int(j/I.M)] = count[symbol]
			}
		}
	}

	return I
}

//-----------------------------------------------------------------------------
func (I *IndexC) Occurence(c byte, pos indexType) indexType {
	i := indexType(pos / indexType(I.M))
	count := I.OCC[c][i]
	for j := i*indexType(I.M) + 1; j <= pos; j++ {
		if I.BWT[j] == c {
			count += 1
		}
	}
	return count
}

// -----------------------------------------------------------------------------
// Returns starting, ending positions (sp, ep) and last-matched position (i)

func (I *IndexC) Search(query []byte) (int, int) {
	var offset indexType
	var i int
	start_pos := len(query) - 1
	c := query[start_pos]
	sp, ok := I.C[c]
	if !ok {
		panic("Unknown character: " + string(c))
	}
	ep := I.EP[c]
	// fmt.Println(ep-sp+1, "\t", i, string(c), len(query))
	for i = int(start_pos - 1); sp <= ep && i >= 0; i-- {
		c = query[i]
		offset, ok = I.C[c]
		if !ok {
			panic("Unknown character: " + string(c))
		}
		sp = offset + I.Occurence(c, sp-1)
		ep = offset + I.Occurence(c, ep) - 1
		// fmt.Println(ep-sp+1, "\t", i, string(c), len(query))
	}
	return int(sp), int(ep)
}

//-----------------------------------------------------------------------------
// Guess which sequence contains the query.
// If randomized_round is 0, there is no randomization. The search begins at the.
//-----------------------------------------------------------------------------

func (I *IndexC) Guess(query []byte, randomized_round int) (int, int) {
	var seq, count int
	// var start_pos, end_pos int
	if randomized_round == 0 {
		seq, count, _ = I._guess(query, len(query)-1)
		return seq, count
	} else {
		for i := 0; i < randomized_round; i++ {
			// start_pos = rand.Intn(len(query))
			// seq, count, end_pos = I._guess(query, start_pos)
			// fmt.Println(end_pos, start_pos, "<")
			seq, count, _ = I._guess(query, rand.Intn(len(query)))
			if seq >= 0 {
				return seq, count
			}
		}
		return -1, 0
	}
}

//-----------------------------------------------------------------------------

func (I *IndexC) GuessPair(query1 []byte, query2 []byte, randomized_round int) int {
	var seq1, seq2 int
	if randomized_round == 0 {
		seq1, _, _ = I._guess(query1, len(query1)-1)
		seq2, _, _ = I._guess(query2, len(query2)-1)
		if seq1 == seq2 {
			return seq1
		} else {
			return -1
		}
	} else {
		for i := 0; i < randomized_round; i++ {
			seq1, _, _ = I._guess(query1, rand.Intn(len(query1)))
			seq2, _, _ = I._guess(query2, rand.Intn(len(query2)))
			if seq1 >= 0 && seq1 == seq2 {
				return seq1
			}
		}
		if seq1 == -1 {
			return seq2
		} else if seq2 == -1 {
			return seq1
		}
		return -1
	}
}

//-----------------------------------------------------------------------------
func (I *IndexC) _guess(query []byte, start_pos int) (int, int, int) {
	if !I.Multiple {
		return 0, -1, -1
	}
	var offset indexType
	var i int
	c := query[start_pos]
	sp, ok := I.C[c]
	if !ok {
		return -2, 0, 0
	}
	ep := I.EP[c]
	// fmt.Println(ep-sp+1, "\t", i, string(c), len(query))
	for i = int(start_pos - 1); sp < ep && i >= 0; i-- {
		c = query[i]
		offset, ok = I.C[c]
		if !ok {
			return -2, 0, 0
		}
		sp = offset + I.Occurence(c, sp-1)
		ep = offset + I.Occurence(c, ep) - 1
		// fmt.Println(ep-sp+1, "\t", i, string(c), len(query))
	}
	if sp <= ep {
		for j := sp + 1; j <= ep; j++ {
			if I.SSA[j] != I.SSA[sp] {
				return -1, int(ep - sp + 1), i
			}
		}
		return int(I.SSA[sp]), int(ep - sp + 1), i
	} else {
		return -1, int(ep - sp + 1), i
	}
}

//-----------------------------------------------------------------------------
func (I *IndexC) ReadFasta(file string) {
	f, err := os.Open(file)
	check_for_error(err)
	defer f.Close()

	if file[len(file)-6:] != ".fasta" {
		panic("ReadFasta:" + file + "is not a fasta file.")
	}

	scanner := bufio.NewScanner(f)
	byte_array := make([]byte, 0)
	i := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			if line[0] != '>' {
				byte_array = append(byte_array, bytes.Trim(line, "\n\r ")...)
			} else if len(byte_array) > 0 {
				byte_array = append(byte_array, byte('|'))
			}
			i++
		}
	}
	I.SEQ = append(byte_array, byte('$'))
}

//-----------------------------------------------------------------------------
func (I *IndexC) Show() {
	fmt.Printf(" %6s %6s  OCC\n", "Freq", "C")
	for i := 0; i < len(I.SYMBOLS); i++ {
		c := byte(I.SYMBOLS[i])
		fmt.Printf("%c%6d %6d  %d\n", c, I.Freq[c], I.C[c], I.OCC[c])
	}
	fmt.Printf("SA ")
	for i := 0; i < len(I.SA); i++ {
		fmt.Print(I.SA[i], " ")
	}
	fmt.Printf("\nBWT ")
	for i := 0; i < len(I.BWT); i++ {
		fmt.Print(string(I.BWT[i]))
	}
	fmt.Println()
	fmt.Printf("\nSSA ")
	for i := 0; i < len(I.SSA); i++ {
		fmt.Printf("%d ", I.SSA[i])
	}
	fmt.Println()
	fmt.Println("SEQ", string(I.SEQ))
}

//-----------------------------------------------------------------------------
func (I *IndexC) Check() {
	fmt.Println("Checking...")
	for i := 0; i < len(I.SYMBOLS); i++ {
		c := byte(I.SYMBOLS[i])
		fmt.Printf("%c%6d %6d  [", c, I.Freq[c], I.C[c])
		for j := 0; j < int(I.LEN); j++ {
			fmt.Printf("%d ", I.Occurence(c, indexType(j)))
		}
		fmt.Printf("]\n")
	}
	if len(I.SEQ) > 0 {
		seq, count := I.Search(I.SEQ[0 : len(I.SEQ)-1])
		fmt.Println("Search for SEQ returns", seq, count)
	}
}

//-----------------------------------------------------------------------------
