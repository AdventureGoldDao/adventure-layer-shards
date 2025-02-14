// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

//go:build !wasm
// +build !wasm

package arbcompress

/*
#cgo CFLAGS: -g -Wall -I${SRCDIR}/../target/include/
#cgo LDFLAGS: ${SRCDIR}/../target/lib/libstylus.a -lm
#include "arbitrator.h"
*/
import "C"
import "fmt"

// Type aliases for C types to improve readability in Go code
type u8 = C.uint8_t
type u32 = C.uint32_t
type usize = C.size_t

type brotliBool = uint32
type brotliBuffer = C.BrotliBuffer

// Brotli compression status constants
const (
	brotliFalse brotliBool = iota
	brotliTrue
)

// CompressWell compresses the input data using the 'well' compression level and an empty dictionary.
func CompressWell(input []byte) ([]byte, error) {
	return Compress(input, LEVEL_WELL, EmptyDictionary)
}

// Compress compresses the input data with the specified compression level and dictionary.
func Compress(input []byte, level uint32, dictionary Dictionary) ([]byte, error) {
	// Allocate a buffer for the compressed data and perform the compression
	maxSize := compressedBufferSizeFor(len(input))
	output := make([]byte, maxSize)
	outbuf := sliceToBuffer(output)
	inbuf := sliceToBuffer(input)

	// Perform compression using the Brotli C library
	status := C.brotli_compress(inbuf, outbuf, C.Dictionary(dictionary), u32(level))
	if status != C.BrotliStatus_Success {
		return nil, fmt.Errorf("failed compression: %d", status)
	}

	// Return the compressed data
	output = output[:*outbuf.len]
	return output, nil
}

// Decompress decompresses the input data using an empty dictionary and a maximum output size.
func Decompress(input []byte, maxSize int) ([]byte, error) {
	return DecompressWithDictionary(input, maxSize, EmptyDictionary)
}

// DecompressWithDictionary decompresses the input data using the provided dictionary and a maximum output size.
func DecompressWithDictionary(input []byte, maxSize int, dictionary Dictionary) ([]byte, error) {
	// Allocate a buffer for the decompressed data and perform the decompression
	output := make([]byte, maxSize)
	outbuf := sliceToBuffer(output)
	inbuf := sliceToBuffer(input)

	// Perform decompression using the Brotli C library
	status := C.brotli_decompress(inbuf, outbuf, C.Dictionary(dictionary))
	if status != C.BrotliStatus_Success {
		return nil, fmt.Errorf("failed decompression: %d", status)
	}

	// Ensure decompressed data doesn't exceed maxSize
	if *outbuf.len > usize(maxSize) {
		return nil, fmt.Errorf("failed decompression: result too large: %d", *outbuf.len)
	}

	// Return the decompressed data
	output = output[:*outbuf.len]
	return output, nil
}

// sliceToBuffer converts a Go slice to a BrotliBuffer, which is a C struct used by the Brotli compression functions.
func sliceToBuffer(slice []byte) brotliBuffer {
	count := usize(len(slice))
	if count == 0 {
		slice = []byte{0x00} // ensures pointer is not null (shouldn't be necessary, but brotli docs are picky about NULL)
	}
	return brotliBuffer{
		ptr: (*u8)(&slice[0]),
		len: &count,
	}
}
