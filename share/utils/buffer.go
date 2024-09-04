// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

// Simple byte buffer for marshaling data.

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

// A StreamBuffer is a variable-sized buffer of bytes with [StreamBuffer.Read] and [StreamBuffer.Write] methods.
// The zero value for StreamBuffer is an empty buffer ready to use.
type StreamBuffer struct {
	buf []byte     // contents are the bytes buf[off : len(buf)]
	off int        // read at &buf[off], write at &buf[len(buf)]
	mut sync.Mutex // makes the StreamBuffer thread secure
	End atomic.Bool
}

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("bytes.Buffer: too large")

const maxInt = int(^uint(0) >> 1)

// empty reports whether the unread portion of the buffer is empty.
func (b *StreamBuffer) empty() bool { return len(b.buf) <= b.off }

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *StreamBuffer) Len() int { return len(b.buf) - b.off }

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as [StreamBuffer.Truncate](0).
func (b *StreamBuffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

// tryGrowByReslice is an inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *StreamBuffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *StreamBuffer) grow(n int) int {
	m := b.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Add b.off to account for b.buf[:b.off] being sliced off the front.
		b.buf = growSlice(b.buf[b.off:], b.off+n)
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with [ErrTooLarge].
func (b *StreamBuffer) Write(p []byte) (n int, err error) {
	b.mut.Lock()
	defer b.mut.Unlock()
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *StreamBuffer) Read(p []byte) (n int, err error) {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		// if len(p) == 0 {
		// 	return 0, nil
		// }
		if b.End.Load() {
			return 0, io.EOF
		}
		return 0, nil
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}
