Uncompressed FM index.

## Create FM index from sequence

```
	idx := fmic.CompressedIndex(sequence_of_bytes, 10)
```

Create an FM index with compression ratio 10.  Larger the compression ratios result in linearly smaller indexes and linearly longer search.

## Save the index

```
	idx.SaveCompressedIndex()
```

The index is stored in a directory named "input_sequence_file.fmi", where "input_sequence_file" is the name of the input sequence file.


## Load an index that was previously saved

```
	saved_idx := fmic.LoadCompressedIndex(index_directory)
```