// Package bloomfilter is face-meltingly fast, thread-safe,
// marshalable, unionable, probability- and
// optimal-size-calculating Bloom filter in go
//
// https://github.com/steakknife/bloomfilter
//
// Copyright © 2014, 2015, 2018 Barry Allard
//
// MIT license
//
package bloomfilter

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

func TestWriteRead(t *testing.T) {
	// minimal filter
	f, _ := New(2, 1)
	v := hashableUint64(0)
	f.Add(v)

	t.Run("binary", func(t *testing.T) {
		var b bytes.Buffer
		_, err := f.WriteTo(&b)
		if err != nil {
			t.Fatal(err)
		}
		var f2 *Filter
		if f2, _, err = ReadFrom(&b); err != nil {
			t.Fatal(err)
		}
		if !f2.Contains(v) {
			t.Error("Filters not equal")
		}
	})
	t.Run("text", func(t *testing.T) {
		text, err := f.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%v\n", string(text))
		var f2 *Filter
		// Test create a new filter
		if f2, err = UnmarshalText(text); err != nil {
			t.Fatal(err)
		}
		if !f2.Contains(v) {
			t.Error("Filters not equal")
		}
		// Test overwrite a filter
		f3, _ := New(8, 8)
		if err = f3.UnmarshalText(text); err != nil {
			t.Fatal(err)
		}
		if !f3.Contains(v) {
			t.Error("Filters not equal")
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

func totAllocMb() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return bToMb(m.TotalAlloc)
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
	f.MarshallToWriter(h1)
	// Marshall as a blob
	data, _ := f.MarshalBinary()
	h2.Write(data)

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
