
// indexType is the type of the indices of the BWT
// if indexType is uint32, the length of sequence must be smaller than 2^32
// for really long sequences, set indexType to int64
type indexType int64

// NUM_BYTES is the number of bytes of indexType
var NUM_BYTES = uint(8)
