
[![GoDoc](https://godoc.org/github.com/holiman/bloomfilter?status.png)](https://godoc.org/github.com/holiman/bloomfilter)
[![CircleCI](https://circleci.com/gh/holiman/bloomfilter.svg?style=svg)](https://app.circleci.com/pipelines/github/holiman/bloomfilter)
[![codecov](https://codecov.io/gh/holiman/bloomfilter/branch/master/graph/badge.svg?token=O48l6LbHkL)](https://codecov.io/gh/holiman/bloomfilter)
[![DeepSource](https://deepsource.io/gh/holiman/bloomfilter.svg/?label=active+issues&show_trend=true)](https://deepsource.io/gh/holiman/bloomfilter/?ref=repository-badge)

# History

This bloom filter implementation is a fork from [steakknife/bloomfilter](https://github.com/steakknife/bloomfilter) by Barry Allard. 
The upstream project is now archived, so this fork exists to fix some bugs and also
make a few improvements. Below is the original description. 

The original implemenation is Copyright © 2014-2016,2018 Barry Allard
[MIT license](MIT-LICENSE.txt)

All recent changes are copyright © 2019-2020 Martin Holst Swende. 

## Installation 

```
$ go get github.com/holiman/bloomfilter
```

## Face-meltingly fast, thread-safe, marshalable, unionable, probability- and optimal-size-calculating Bloom filter in go

### WTF is a bloom filter

**TL;DR:** Probabilistic, extra lookup table to track a set of elements kept elsewhere to reduce expensive, unnecessary set element retrieval and/or iterator operations **when an element is not present in the set.** It's a classic time-storage tradeoff algorithm.

### Properties

#### [See wikipedia](https://en.wikipedia.org/wiki/Bloom_filter) for algorithm details

|Impact|What|Description|
|---|---|---|
|Good|No false negatives|know for certain if a given element is definitely NOT in the set|
|Bad|False positives|uncertain if a given element is in the set|
|Bad|Theoretical potential for hash collisions|in very large systems and/or badly hash.Hash64-conforming implementations|
|Bad|Add only|Cannot remove an element, it would destroy information about other elements|
|Good|Constant storage|uses only a fixed amount of memory|

## Naming conventions

(Similar to algorithm)

|Variable/function|Description|Range|
|---|---|---|
|m/M()|number of bits in the bloom filter (memory representation is about m/8 bytes in size)|>=2|
|n/N()|number of elements present|>=0|
|k/K()|number of keys to use (keys are kept private to user code but are de/serialized to Marshal and file I/O)|>=0|
|maxN|maximum capacity of intended structure|>0|
|p|maximum allowed probability of collision (for computing m and k for optimal sizing)|>0..<1|

- Memory representation should be exactly `24 + 8*(k + (m+63)/64) + unsafe.Sizeof(RWMutex)` bytes.
- Serialized (`BinaryMarshaler`) representation should be exactly `72 + 8*(k + (m+63)/64)` bytes. (Disk format is less due to compression.)

## Binary serialization format

All values in Little-endian format

|Offset|Offset (Hex)|Length (bytes)|Name|Type|
|---|---|---|---|---|
|0|00|12|magic + version number|`\0\0\0\0\0\0\0\0v02\n`|
|12|0c|8|k|`uint64`|
|20|14|8|n|`uint64`|
|28|1c|8|m|`uint64`|
|36|24|k|(keys)|`[k]uint64`|
|36+8*k|...|(m+63)/64|(bloom filter)|`[(m+63)/64]uint64`|
|36+8\*k+8\*((m+63)/64)|...|48|(SHA384 of all previous fields, hashed in order)|`[48]byte`|

- `bloomfilter.Filter` conforms to `encoding.BinaryMarshaler` and `encoding.BinaryUnmarshaler'

## Usage

```go

import "github.com/holiman/bloomfilter"

const (
  maxElements = 100000
  probCollide = 0.0000001
)

bf, err := bloomfilter.NewOptimal(maxElements, probCollide)
if err != nil {
  panic(err)
}

someValue := ... // must conform to hash.Hash64

bf.Add(someValue)
if bf.Contains(someValue) { // probably true, could be false
  // whatever
}

anotherValue := ... // must also conform to hash.Hash64

if bf.Contains(anotherValue) {
  panic("This should never happen")
}

err := bf.WriteFile("1.bf.gz")  // saves this BF to a file
if err != nil {
  panic(err)
}

bf2, err := bloomfilter.ReadFile("1.bf.gz") // read the BF to another var
if err != nil {
  panic(err)
}
```


## Design

Where possible, branch-free operations are used to avoid deep pipeline / execution unit stalls on branch-misses.

## Contact

- [Issues](https://github.com/holiman/bloomfilter/issues)

## License

[MIT license](MIT-LICENSE.txt)

Copyright © 2014-2016 Barry Allard
Copyright © 2019-2020 Martin Holst Swende

