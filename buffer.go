// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

//Package bbufer is a more efficient bytes buffer both in time and memory
// compared to the standard library `bytes.Buffer`.
package bbuffer

import (
	"errors"
	"runtime"
	"sync"
)

// Buffer struct
type Buffer struct {
	bufs [][]byte // storing the wirtten data
	acci []int    // accumulated size
}

// NewBuffer creates a new buffer
func NewBuffer() *Buffer {
	return new(Buffer)
}

// Writes appends byte slice to the slice of byte slice
func (b *Buffer) Write(p []byte) {
	if b.bufs == nil {
		b.bufs = make([][]byte, 0)
		b.acci = make([]int, 0)
	}

	b.bufs = append(b.bufs, p)
	if len(b.acci) == 0 {
		b.acci = append(b.acci, len(p))
	} else {
		b.acci = append(b.acci, b.acci[len(b.acci)-1]+len(p))
	}
}

// Len returns the length of all bytes
func (b *Buffer) Len() int {
	if b.acci == nil {
		return 0
	}
	return b.acci[len(b.acci)-1]
}

// ConcurrencyNum is Concurrency Number for parallelly copying slice
var ConcurrencyNum = runtime.NumCPU()

// Bytes returns the long bytes
func (b *Buffer) Bytes() []byte {
	result := make([]byte, b.Len())
	i := 0
	var wg sync.WaitGroup
	tokens := make(chan int, ConcurrencyNum)
	for _, buf := range b.bufs {
		wg.Add(1)
		tokens <- 1
		go func(i int, buf []byte) {
			copy(result[i:i+len(buf)], buf)
			wg.Done()
			<-tokens
		}(i, buf)
		i += len(buf)
	}
	wg.Wait()
	return result
}

// ErrInvalidSliceRange is the error when using invalid range for Slice() method
var ErrInvalidSliceRange = errors.New("bbufer: invalid slice range")

// Slice is used to slicing the byte slice
func (b *Buffer) Slice(start, end int) ([]byte, error) {
	if start < 0 || end < 0 {
		return nil, ErrInvalidSliceRange
	}
	if end > b.Len() {
		end = b.Len()
	}

	result := make([]byte, end-start)
	for i, acci := range b.acci {
		if start > acci {
			continue
		}
		if end <= acci { // right here
			if i > 0 {
				if start >= b.acci[i-1] {
					copy(result, b.bufs[i][start-b.acci[i-1]:end-b.acci[i-1]])
				} else {
					copy(result[b.acci[i-1]-start:], b.bufs[i][0:end-b.acci[i-1]])
				}
			} else {
				copy(result, b.bufs[i][start:end])
			}
			return result, nil
		}
		// not reach the end
		if i > 0 {
			copy(result[b.acci[i-1]-start:b.acci[i-1]-start+len(b.bufs[i])], b.bufs[i])
		} else {
			copy(result[0:b.acci[i]-start], b.bufs[i][start:])
		}
	}

	return result, nil
}
