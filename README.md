Relatively efficient implementation of uncompressed FM index.  Search is optimally linear in query length.  Index size can be compressed; how much compression is specified by users.

## Create FM index from sequence

```
	idx := fmic.CompressedIndex(sequence_of_bytes, 10)
```

Create an FM index with compression ratio 10.  Larger compression ratios result in linearly smaller indexes and linearly longer search.

## Save the index

```
	idx.SaveCompressedIndex()
```

The index is stored in a directory named "input_sequence_file.fmi", where "input_sequence_file" is the name of the input sequence file.


## Load an index that was previously saved

```
	saved_idx := fmic.LoadCompressedIndex(index_directory)
```

## Features

- User-definable compresion ratio as a trade off between size of index and search time.
- Multiple goroutines to save/load index quickly.
- Should work with sequences longer than 4GB.
- Suffix array is built quickly using the SAIS algorithm.
