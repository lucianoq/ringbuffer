# ringbuffer

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/lucianoq/ringbuffer?status.svg)](https://godoc.org/github.com/lucianoq/ringbuffer)
[![Go Report Card](https://goreportcard.com/badge/github.com/lucianoq/ringbuffer)](https://goreportcard.com/report/github.com/lucianoq/ringbuffer)
 
This repository provides the `ringbuffer` package.

The package provides a `RingBuffer` object that behaves like a 
[Dynamic Array](https://en.wikipedia.org/wiki/Dynamic_array) until the maximum 
size is reached, and, after that, it behaves like a 
[Circular Buffer](https://en.wikipedia.org/wiki/Circular_buffer) always 
overwriting the oldest content without using new memory.

The buffer implements the `io.Writer`, `io.Closer` and `fmt.Stringer`
interfaces.

---

The existing ringbuffers written in Go like:
- [github.com/armon/circbuf](https://github.com/armon/circbuf)
- [github.com/glycerine/rbuf](https://github.com/glycerine/rbuf)
- [github.com/smallnest/ringbuffer](https://github.com/smallnest/ringbuffer)

are all fixed-size buffered, that means you are forced to allocate all the 
memory. 

If in the most frequent case you are not expecting to write data or you need 
just few bytes, but you still want the safety to cap your buffer and not be 
killed for OOM, this package could be convenient.  


# Documentation

Full documentation can be found on [Godoc](http://godoc.org/github.com/lucianoq/ringbuffer)

# Usage


```go
package main

import (
    "fmt"

    "github.com/lucianoq/ringbuffer"
)

func main() {
    // create an empty buffer that can't exceed 1024 bytes
    buf := ringbuffer.NewRingBuffer(0, 1024)
    
    // it implements io.Writer, so it can be used in Fprint()
    fmt.Fprint(buf, "hello world")
    
    // prints: hello world
    fmt.Println(buf)
}
```

### Quick methods overview

```go
// Empty buffer
buf := NewRingBuffer(0, 1024)

// If I'm expecting data, I can pre-allocate the buffer for better performance
buf.Grow(512)

// Returns the allocated size of the byte buffer
n := buf.Cap()

// Returns the counter of written data so far
n := buf.Written()

// Writes on the Buffer
buf.Write([]byte{'a', 'b'})

// Returns the content in a byte slice
slice := buf.Bytes()

// Returns the content in a string
str := buf.String()

// Clean the buffer
buf.Reset()

// Close the buffer. No more operation allowed on it.
// It can be garbage collected.
buf.Close()
```

### Example

```go
package main

import (
    "fmt"

    "github.com/lucianoq/ringbuffer"
)

func main() {
	r := ringbuffer.NewRingBuffer(0, 21)

	for i := 0; i < 20; i++ {
		fmt.Printf("memory: %d\twritten: %d\tcontent: %s\n", r.Cap(), r.Written(), r)
		fmt.Fprintf(r, "%d", i)
	}
}

// memory: 0	written: 0	content: 
// memory: 1	written: 1	content: 0
// memory: 2	written: 2	content: 01
// memory: 4	written: 3	content: 012
// memory: 4	written: 4	content: 0123
// memory: 8	written: 5	content: 01234
// memory: 8	written: 6	content: 012345
// memory: 8	written: 7	content: 0123456
// memory: 8	written: 8	content: 01234567
// memory: 16	written: 9	content: 012345678
// memory: 16	written: 10	content: 0123456789
// memory: 16	written: 12	content: 012345678910
// memory: 16	written: 14	content: 01234567891011
// memory: 16	written: 16	content: 0123456789101112
// memory: 21	written: 18	content: 012345678910111213
// memory: 21	written: 20	content: 01234567891011121314
// memory: 21	written: 22	content: 123456789101112131415
// memory: 21	written: 24	content: 345678910111213141516
// memory: 21	written: 26	content: 567891011121314151617
// memory: 21	written: 28	content: 789101112131415161718
```