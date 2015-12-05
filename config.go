package fmic

// indexType is the type of the indices of the sequence to be indexed.
// If indexType is uint32, the length of sequence must be smaller than 2^32
// For really long sequences, set indexType to int64
type indexType int64

// NUM_BYTES is the number of bytes of indexType
const NUM_BYTES = uint(8)
