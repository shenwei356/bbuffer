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

package bbuffer

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

var alphabet = "12345678ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz.-_;,"
var texts [][]byte // strings of different scales
var points [][]int // points for fragmenting the text
var scales []int

func init() {
	rand.Seed(1)

	scales = []int{1e3, 1e6, 1e8}
	texts = make([][]byte, len(scales))
	points = make([][]int, len(scales))

	for i, s := range scales {
		fmt.Printf("generate test data for test. scale %d/%d: %d\n", i+1, len(scales), s)
		text := make([]byte, (rand.Intn(10)+1)*s)
		for j := range text {
			text[j] = alphabet[rand.Intn(len(alphabet))]
		}
		texts[i] = text

		s, r := 0, 0
		for s < len(text) {
			r = rand.Intn(len(text)/100) + 1 // about 100 fragments
			s += r
			if s < len(text) {
				points[i] = append(points[i], s)
			}
		}
	}
}

func TestBuffer(t *testing.T) {
	for i, text := range texts {
		var b Buffer
		for j, p := range points[i] {
			if j == 0 {
				b.Write(text[0:p])
			} else if j == len(points[i])-1 {
				b.Write(text[points[i][j-1]:])
			} else {
				b.Write(text[points[i][j-1]:p])
			}
		}

		if bytes.Compare(b.Bytes(), text) != 0 {
			t.Errorf("Unequal data, true:%s, false:%s\n", string(text), string(b.Bytes()))
		}
	}
}

func BenchmarkBBufer(t *testing.B) {
	for i, text := range texts {
		var b Buffer

		for j, p := range points[i] {
			if j == 0 {
				b.Write(text[0:p])
			} else if j == len(points[i])-1 {
				b.Write(text[points[i][j-1]:])
			} else {
				b.Write(text[points[i][j-1]:p])
			}
		}

		b.Bytes()
		b.Reset()
	}
}

func BenchmarkBytesBufer(t *testing.B) {
	for i, text := range texts {
		var b bytes.Buffer

		for j, p := range points[i] {
			if j == 0 {
				b.Write(text[0:p])
			} else if j == len(points[i])-1 {
				b.Write(text[points[i][j-1]:])
			} else {
				b.Write(text[points[i][j-1]:p])
			}
		}

		b.Bytes()
		b.Reset()
	}
}

func TestMethodSlice(t *testing.T) {
	var b Buffer
	b.Write([]byte("012345"))
	b.Write([]byte("67890a"))
	b.Write([]byte("bcdef"))
	str := b.Bytes()

	rs := [][]int{
		// range in one subslice
		[]int{0, 0},
		[]int{0, 3},
		[]int{0, 5},
		[]int{2, 4},

		// range in another subslice
		[]int{6, 10},
		[]int{7, 9},
		[]int{13, 14},
		[]int{len(str) - 3, len(str)},

		// range accross multiple subslices
		[]int{3, 7},
		[]int{3, 15},
		[]int{6, 15},
		[]int{0, len(str)},
		[]int{1, len(str)},
	}
	for _, r := range rs {
		s, err := b.Slice(r[0], r[1])
		if err != nil {
			t.Errorf("unwanted err")
			return
		}
		if bytes.Compare(str[r[0]:r[1]], s) != 0 {
			t.Errorf("wrong Slice result of %d:%d, true:%s, false:%s",
				r[0], r[1], string(str[r[0]:r[1]]), string(s))
			return
		}
	}
}
