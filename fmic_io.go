/*
   Copyright 2015 Vinhthuy Phan
	Compressed FM index.
*/
package fmic

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type Symb_OCC struct {
	Symb int
	OCC  []int64
}

//-----------------------------------------------------------------------------
func check_for_error(e error) {
	if e != nil {
		panic(e)
	}
}
//-----------------------------------------------------------------------------
func ReadFasta(file string) []byte {
	f, err := os.Open(file)
	check_for_error(err)
	defer f.Close()

	if file[len(file)-6:] != ".fasta" {
		panic("ReadFasta:" + file + "is not a fasta file.")
	}

	scanner := bufio.NewScanner(f)
	byte_array := make([]byte, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			if line[0] != '>' {
				byte_array = append(byte_array, bytes.Trim(line, "\n\r ")...)
			} else if len(byte_array) > 0 {
				byte_array = append(byte_array, byte('|'))
			}
		}
	}
	return append(byte_array, byte('$'))
}

//-----------------------------------------------------------------------------
// Save the index to directory.

func _save_slice(s []int64, filename string) {
	f, err := os.Create(filename)
	check_for_error(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	err = binary.Write(w, binary.LittleEndian, s)
	check_for_error(err)
	w.Flush()
}

// ------------------------------------------------------------------
// save_option:
// 	0 - do not save suffix array and seq
//		1 - save suffix array, but not seq
//		2 - save both suffix array and seq
// ------------------------------------------------------------------
func (I *IndexC) SaveCompressedIndex(save_option int) {
	dir := I.input_file + ".fmi"
	os.Mkdir(dir, 0777)

	var wg sync.WaitGroup
	wg.Add(len(I.SYMBOLS) + 3)

	go func() {
		defer wg.Done()
		err := ioutil.WriteFile(path.Join(dir, "bwt"), I.BWT, 0666)
		check_for_error(err)
	}()

	go func() {
		defer wg.Done()
		if save_option == 1 || save_option == 2 {
			_save_slice(I.SA, path.Join(dir, "sa"))
		}
	}()

	go func() {
		defer wg.Done()
		if save_option == 2 {
			err := ioutil.WriteFile(path.Join(dir, "seq"), I.SEQ, 0666)
			check_for_error(err)
		}
	}()

	for symb := range I.OCC {
		go func(symb byte) {
			defer wg.Done()
			_save_slice(I.OCC[symb], path.Join(dir, "occ."+string(symb)))
		}(symb)
	}

	f, err := os.Create(path.Join(dir, "others"))
	check_for_error(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "%d %d %d %d %d\n", I.LEN, I.OCC_SIZE, I.END_POS, I.M, save_option)
	for i := 0; i < len(I.SYMBOLS); i++ {
		symb := byte(I.SYMBOLS[i])
		fmt.Fprintf(w, "%s %d %d %d\n", string(symb), I.Freq[symb], I.C[symb], I.EP[symb])
	}
	w.Flush()
	wg.Wait()
}

// ------------------------------------------------------------------
// save_option:
// 	0 - suffix array and seq were not saved
//		1 - suffix array was saved; seq was not
//		2 - both suffix array and seq were saved
// ------------------------------------------------------------------
func LoadCompressedIndex(dir string) *IndexC {
	I := new(IndexC)

	// First, load "others"
	f, err := os.Open(path.Join(dir, "others"))
	check_for_error(err)
	defer f.Close()

	var symb byte
	var freq, c, ep int64
	var save_option int
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	fmt.Sscanf(scanner.Text(), "%d%d%d%d%d\n", &I.LEN, &I.OCC_SIZE, &I.END_POS, &I.M, &save_option)

	I.Freq = make(map[byte]int64)
	I.C = make(map[byte]int64)
	I.EP = make(map[byte]int64)
	for scanner.Scan() {
		fmt.Sscanf(scanner.Text(), "%c%d%d%d", &symb, &freq, &c, &ep)
		I.SYMBOLS = append(I.SYMBOLS, int(symb))
		I.Freq[symb], I.C[symb], I.EP[symb] = freq, c, ep
	}

	// Second, load Suffix array, BWT and OCC
	I.OCC = make(map[byte][]int64)
	var wg sync.WaitGroup
	wg.Add(len(I.SYMBOLS) + 3)

	go func() {
		defer wg.Done()
		I.BWT, err = ioutil.ReadFile(path.Join(dir, "bwt"))
		check_for_error(err)
	}()

	go func() {
		defer wg.Done()
		if save_option == 1 || save_option == 2 {
			I.SA = _load_slice(path.Join(dir, "sa"), I.LEN)
		}
	}()
	go func() {
		defer wg.Done()
		if save_option == 2 {
			I.SEQ, err = ioutil.ReadFile(path.Join(dir, "seq"))
			check_for_error(err)
		}
	}()

	Symb_OCC_chan := make(chan Symb_OCC)
	for _, symb := range I.SYMBOLS {
		go func(symb int) {
			defer wg.Done()
			Symb_OCC_chan <- Symb_OCC{symb, _load_slice(path.Join(dir, "occ."+string(symb)), I.OCC_SIZE)}
		}(symb)
	}
	go func() {
		wg.Wait()
		close(Symb_OCC_chan)
	}()

	for symb_occ := range Symb_OCC_chan {
		I.OCC[byte(symb_occ.Symb)] = symb_occ.OCC
	}
	return I
}

//-----------------------------------------------------------------------------
func _load_slice(filename string, length int64) []int64 {
	f, err := os.Open(filename)
	check_for_error(err)
	defer f.Close()

	v := make([]int64, length)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanBytes)
	for i := 0; scanner.Scan(); i++ {
		// convert 8 consecutive bytes to a int64 number
		v[i] = int64(scanner.Bytes()[0])
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 8
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 16
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 24
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 32
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 40
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 48
		scanner.Scan()
		v[i] += int64(scanner.Bytes()[0]) << 56
	}
	// r := bufio.NewReader(f)
	// binary.Read(r, binary.LittleEndian, v)
	return v
}
//-----------------------------------------------------------------------------
