/*
   Copyright 2015 Vinhthuy Phan
	Compressed FM index.
*/
package fmic

import (
	"fmt"
	"sort"
	"math"
	"bytes"
	"os"
	"bufio"
)

//-----------------------------------------------------------------------------
// Global variables: sequence (SEQ), suffix array (SA), BWT, FM index (C, OCC)
//-----------------------------------------------------------------------------

type IndexC struct {
	SEQ []byte
	BWT []byte
	SA  []indexType          // suffix array
	C   map[byte]indexType   // count table
	OCC map[byte][]indexType // occurence table
	REGION []regionType		 // REGION[i] stores the region contains the suffix at SA[i]
	tmpREGION []regionType

	END_POS indexType          // position of "$" in the text
	SYMBOLS []int          // sorted symbols
	EP      map[byte]indexType // ending row/position of each symbol

	LEN  indexType
	OCC_SIZE indexType
	Freq map[byte]indexType // Frequency of each symbol
	M int             // Compression ratio
	input_file string
}

//-----------------------------------------------------------------------------
// Build FM index given the file storing the text.
func CompressedIndex(file string, compression_ratio int) *IndexC {
	I := new(IndexC)
	I.input_file = file
	I.M = compression_ratio

	// GET THE SEQUENCE
	I.ReadFasta(file)

	// BUILD SUFFIX ARRAY
	I.LEN = indexType(len(I.SEQ))
	I.OCC_SIZE = indexType(math.Ceil(float64(I.LEN/indexType(I.M))))+1
	I.SA = make([]indexType, I.LEN)
	I.REGION = make([]regionType, I.LEN)
	I.tmpREGION = make([]regionType, I.LEN)
	SA := make([]int, I.LEN)
	ws := &WorkSpace{}
	ws.ComputeSuffixArray(I.SEQ, SA)
	region := regionType(0)
	for i := range SA {
		I.SA[i] = indexType(SA[i])
		I.tmpREGION[i] = region
		if I.SEQ[i] == '|' {
			region++
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
		I.REGION[i] = I.tmpREGION[I.SA[i]]
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
		if j % I.M == 0 {
			for symbol := range I.OCC {
				I.OCC[symbol][int(j/I.M)] = count[symbol]
			}
		}
	}

	return I
}

//-----------------------------------------------------------------------------
func (I *IndexC) Occurence(c byte, pos indexType) indexType {
	i := indexType(pos/indexType(I.M))
	count := I.OCC[c][i]
	for j:=i*indexType(I.M)+1; j<=pos; j++ {
		if I.BWT[j]==c {
			count += 1
		}
	}
	return count
}

//-----------------------------------------------------------------------------
// Returns starting, ending positions (sp, ep) and last-matched position (i)
func (I *IndexC) Search(pattern []byte) (int, int, int) {
	var offset indexType
	var i int
	start_pos := len(pattern) - 1
	c := pattern[start_pos]
	sp := I.C[c]
	ep := I.EP[c]
	for i = int(start_pos - 1); sp <= ep && i >= 0; i-- {
		c = pattern[i]
		offset = I.C[c]
		sp = offset + I.Occurence(c,sp-1)
		ep = offset + I.Occurence(c,ep) - 1
	}
	return int(sp), int(ep), i + 1
}

//-----------------------------------------------------------------------------
func (I *IndexC) SearchRegion(pattern []byte) (int, int, int) {
	var offset indexType
	var i int
	start_pos := len(pattern) - 1
	c := pattern[start_pos]
	sp, ok := I.C[c]
	if !ok {
		panic("Unknown character: " + string(c))
	}
	ep := I.EP[c]
	for i = int(start_pos - 1); sp < ep && i >= 0; i-- {
		c = pattern[i]
		offset, ok = I.C[c]
		if !ok {
			panic("Unknown character: " + string(c))
		}
		sp = offset + I.Occurence(c,sp-1)
		ep = offset + I.Occurence(c,ep) - 1
	}
	// fmt.Println(sp, ep, i)
	if sp <= ep {
		for j:=sp+1; j<=ep; j++ {
			if I.REGION[j] != I.REGION[sp] {
				return -1, int(ep-sp+1), i
			}
		}
		return int(I.REGION[sp]), int(ep-sp+1), i
	} else {
		return -1, int(ep-sp+1), i
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
				fmt.Println(">", len(byte_array))
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
	fmt.Printf("\ntmpREGION ")
	for i:=0 ; i<len(I.tmpREGION); i++ {
		fmt.Printf("%d ", I.tmpREGION[i])
	}
	fmt.Println()
	fmt.Printf("\nREGION ")
	for i:=0 ; i<len(I.REGION); i++ {
		fmt.Printf("%d ", I.REGION[i])
	}
	fmt.Println()
	fmt.Println("SEQ", string(I.SEQ))
}
//-----------------------------------------------------------------------------
func (I *IndexC) Check() {
	fmt.Println("Checking...")
	for i:=0; i<len(I.SYMBOLS); i++ {
		c := byte(I.SYMBOLS[i])
		fmt.Printf("%c%6d %6d  [", c, I.Freq[c], I.C[c])
		for j:=0; j<int(I.LEN); j++ {
			fmt.Printf("%d ", I.Occurence(c,indexType(j)))
		}
		fmt.Printf("]\n")
	}
	if len(I.SEQ) > 0 {
		a, b, c := I.Search(I.SEQ[0 : len(I.SEQ)-1])
		fmt.Println("Search for SEQ returns", a, b, c)
	}
}

//-----------------------------------------------------------------------------
