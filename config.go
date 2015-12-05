package fmic

// indexType is the type of the indices of the sequence to be indexed.
// If indexType is uint32, the length of sequence must be smaller than 2^32.
// For example, uint32 can be used for the human genome.  But int64 must
// be used a collective metagenome that is longer than 4Gbp.

type indexType int64

// NUM_BYTES is the number of bytes of indexType
const NUM_BYTES = uint(8)
