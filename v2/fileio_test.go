// Package bloomfilter is face-meltingly fast, thread-safe,
// marshalable, unionable, probability- and
// optimal-size-calculating Bloom filter in go
//
// https://github.com/holiman/bloomfilter
//
// Original source:
// https://github.com/steakknife/bloomfilter
//
// Copyright © 2014, 2015, 2018 Barry Allard
// Copyright © 2020 Martin Holst Swende
//
// MIT license
//

package v2

import (
	"bytes"
	"crypto/sha512"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type devnull struct{}

func (d devnull) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestWriteRead(t *testing.T) {
	// minimal filter
	f, _ := New(8*1024*100, 5)
	// Add some content
	var tests = make([]hashableUint64, 20)
	for i := 0; i < 20; i++ {
		tests[i] = hashableUint64(rand.Uint64())
		f.Add(tests[i])
	}
	verify := func(t *testing.T, f *Filter) {
		for i, v := range tests {
			if !f.Contains(v) {
				t.Errorf("missing item %d", i)
			}
		}
	}

	t.Run("binary", func(t *testing.T) {
		var b bytes.Buffer
		_, err := f.WriteTo(&b)
		if err != nil {
			t.Fatal(err)
		}
		cpy := append([]byte{}, b.Bytes()...)
		var f2 *Filter
		if f2, _, err = ReadFrom(&b); err != nil {
			t.Fatal(err)
		}
		verify(t, f2)
		// test overwrite
		f3, _ := New(8*5, 3)
		if _, err = f3.ReadFrom(bytes.NewReader(cpy)); err != nil {
			t.Fatal(err)
		}
		verify(t, f3)
	})
	t.Run("gob", func(t *testing.T) {
		var buffer bytes.Buffer
		err := gob.NewEncoder(&buffer).Encode(f)
		if err != nil {
			t.Fatal(err)
		}
		var f2 Filter
		err = gob.NewDecoder(&buffer).Decode(&f2)
		if err != nil {
			t.Fatal(err)
		}
		verify(t, &f2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := json.Marshal(f)
		if err != nil {
			t.Fatal(err)
		}
		var f2 Filter
		if err = json.Unmarshal(data, &f2); err != nil {
			t.Fatal(err)
		}
		verify(t, &f2)
	})
	t.Run("file", func(t *testing.T) {
		fName := filepath.Join(os.TempDir(), "temp.deleteme.gz")
		if _, err := f.WriteFile(fName); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(fName)
		if f2, _, err := ReadFile(fName); err != nil {
			t.Fatal(err)
		} else {
			verify(t, f2)
		}
	})
}

func TestCorruption(t *testing.T) {
	// minimal filter
	f, _ := New(8*32, 5)
	// Add some content
	var tests = make([]hashableUint64, 20)
	for i := 0; i < 20; i++ {
		tests[i] = hashableUint64(rand.Uint64())
		f.Add(tests[i])
	}
	t.Run("binary", func(t *testing.T) {
		var b bytes.Buffer
		_, err := f.WriteTo(&b)
		if err != nil {
			t.Fatal(err)
		}
		buf := b.Bytes()
		buf[len(buf)/2] ^= 1
		if _, _, err := ReadFrom(&b); err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("gob", func(t *testing.T) {
		var buffer bytes.Buffer
		err := gob.NewEncoder(&buffer).Encode(f)
		if err != nil {
			t.Fatal(err)
		}
		data := buffer.Bytes()
		// Flip a bit
		data[len(data)/2] ^= 1
		var f2 Filter
		err = gob.NewDecoder(&buffer).Decode(&f2)
		if err == nil {
			t.Errorf("expected error")
		}
	})

}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func TestWrite(t *testing.T) {
	// 1Mb
	f, _ := New(4*8*1024*1024, 1)
	fmt.Printf("Allocated 1mb filter\n")
	PrintMemUsage()
	_, _ = f.WriteTo(devnull{})
	fmt.Printf("Wrote filter to devnull\n")
	PrintMemUsage()
}

// fillRandom fills the filter with N random values, where N is roughly half
// the size of the number of uint64's in the filter
func fillRandom(f *Filter) {
	num := len(f.bits) * 4
	for i := 0; i < num; i++ {
		f.AddHash(uint64(rand.Int63()))
	}
}

// TestMarshaller tests that it writes outputs correctly.
func TestMarshaller(t *testing.T) {

	h1 := sha512.New384()
	h2 := sha512.New384()

	f, _ := New(1*8*1024*1024, 1)
	fillRandom(f)
	// Marshall using writer
	_, _, _ = f.MarshallToWriter(h1)
	// Marshall as a blob
	data, _ := f.MarshalBinary()
	_, _ = h2.Write(data)

	if have, want := h1.Sum(nil), h2.Sum(nil); !bytes.Equal(have, want) {
		t.Errorf("Marshalling error, have %x want %x", have, want)
	}
}

func BenchmarkWrite1Mb(b *testing.B) {

	// 1Mb
	f, _ := New(1*8*1024*1024, 1)
	f.Add(hashableUint64(0))
	f.Add(hashableUint64(1))
	f.Add(hashableUint64(1 << 3))
	f.Add(hashableUint64(1 << 40))
	f.Add(hashableUint64(1 << 23))
	f.Add(hashableUint64(1 << 16))
	f.Add(hashableUint64(1 << 28))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = f.WriteTo(devnull{})
	}
}
