Relatively efficient implementation of uncompressed FM index.  Construction is optimally linear.  Search is optimally linear in query length.  Index size can be compressed; how much compression is specified by users.

## Import

```
import "github.com/vtphan/fmic"
```

## Create FM index from sequence

```
	idx := fmic.CompressedIndex(sequence_of_bytes, true, 10)
```

Arguments:

1. Name of file that stores the sequence(s).
2. true if there are multiple sequences in the file.
3. compression ration. Larger compression ratios result in linearly smaller indexes and linearly longer search.

## Save the index

```
	idx.SaveCompressedIndex(0)
```

The index is stored in a directory named "input_sequence_file.fmi", where "input_sequence_file" is the name of the input sequence file.

SavedCompressedIndex takes as input a save_option, which has value 0, 1, or 2:

- 0: Do not save suffix array and seq.
- 1: Save uffix array was, but do not save seq.
- 2: Save both suffix array and seq.

## Load an index that was previously saved

```
	saved_idx := fmic.LoadCompressedIndex(index_directory)
```

## Query search

API is subject to change.

```
s, e, _ := saved_idx.Search(pattern)
```
**e-s+1** is the number of occurrences of the pattern in the indexed sequence.


## Guess which sequence contains a query

See examples/guess_sequence.go

```
	seq, count, _ := saved_idx.GuessSequence([]byte(q))
```

Return values:
- seq: the id of the sequence most likely contains the query q.
- count: the occurences of the query q in the sequence.

Assumptions:

+ q must occur in one of the sequences.
+ But q might be slightly changed (e.g. due to sequencing error or genetic variation).

## Features

- Should work with sequences with fewer than 2^63 (or ~9223 quadrillion) characters.
- User-definable compresion ratio as a trade off between size of index and search time.
- Multiple goroutines to save/load index quickly.
- Suffix array is built quickly using the SAIS algorithm. (Go's built-in suffix array is slow.)
