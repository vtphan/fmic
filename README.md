Uncompressed FM index.

## Create FM index from sequence

```
	idx := fmic.CompressedIndex(sequence_of_bytes, 10)
```

Create an FM index with compression ratio 10.  The larger the compression ratio, the smaller the index and the longer search result. Search for a pattern is still optimal, with running time c*m, where m is the length of the pattern and c is a constant that is propotionally to the compression ratio.

## Save the index

```
	idx.SaveCompressedIndex()
```

The index is stored in a directory named "input_sequence_file.fmi", where "input_sequence_file" is the name of the input sequence file.


## Load an index that was previously saved

```
	saved_idx := fmic.LoadCompressedIndex(index_directory)
```