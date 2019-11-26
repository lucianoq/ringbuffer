// Package ringbuffer provides a byte RingBuffer object that behaves like a
// https://en.wikipedia.org/wiki/Dynamic_array until the maximum size is
// reached, and, after that, like a
// https://en.wikipedia.org/wiki/Circular_buffer always overwriting the oldest
// content without using new memory.
// The buffer implements the io.Writer, io.Closer and fmt.Stringer interfaces.
package ringbuffer

import "errors"

// expansionFactor is the growing factor of the underlying slice
const expansionFactor = 2

// RingBuffer is a variable-sized buffer of bytes with a maximum size.
// After the limit is reached, it behaves like a ring buffer, overwriting
// content if and retaining only the last `size` bytes.
// The buffer implements the io.Writer and io.Closer interface.
type RingBuffer struct {
	buf      []byte
	pos      int
	written  int
	ringMode bool
	maxSize  int
}

// Cap returns the actual size of memory allocated for the underlying buffer.
// This size can't be higher than the maximum size defined in the constructor.
func (r *RingBuffer) Cap() int {
	return cap(r.buf)
}

// Close removes any reference of the underlying slice letting the memory be
// freed.
// Any other method called on this RingBuffer has no meaning and could lead to
// panic.
func (r *RingBuffer) Close() error {
	r.buf = nil
	r.pos = 0
	r.ringMode = false
	r.written = 0
	r.maxSize = 0
	return nil
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (r *RingBuffer) Write(p []byte) (int, error) {
	if r.ringMode {
		return r.writeRing(p)
	}
	return r.write(p)
}

// write copies the input slice p into the internal buffer.
// If the buffer is big enough, it simply copies it.
// If the buffer is smaller than required, it tries to expand it enough to
// contains the input, using the expansionFactor.
// If the maximumSize is reached, it acts like a ring buffer and calls writeRing.
func (r *RingBuffer) write(p []byte) (int, error) {
	pLen := len(p)

	// if necessary, expands r.buf at least size || max
	if len(r.buf) < r.pos+pLen {
		err := r.Grow(r.pos + pLen)
		if err != nil {
			return 0, err
		}
	}

	// If buf can fit the write, do it
	if len(r.buf) >= r.pos+pLen {
		n := copy(r.buf[r.pos:], p)
		r.written += n
		r.pos += n
		if r.pos == r.maxSize {
			r.ringMode = true
			r.pos = 0
		}
		return n, nil
	}

	// buf is full and can't fit this write,
	// it's time to behave like a ring
	r.ringMode = true
	return r.writeRing(p)
}

// writeRing copies the input slice into the internal buffer. If during
// writing the maximum length has been reached, it starts from the beginning
// overriding the oldest content.
func (r *RingBuffer) writeRing(p []byte) (int, error) {
	pLen, bufLen := len(p), len(r.buf)

	// if we are going to write more than the buf size,
	// we just need to keep the last bufLen bytes of the input slice
	// cause the previous will be overwritten
	if pLen > bufLen {
		copy(r.buf, p[pLen-bufLen:])
		r.pos = 0
		r.written += pLen
		return pLen, nil
	}

	// write to the end
	written := copy(r.buf[r.pos:], p)
	r.pos = 0

	// if there is still something to write
	if pLen > written {
		r.pos = copy(r.buf, p[written:])
	}

	r.written += pLen

	return pLen, nil
}

// Grow expands the underlying buffer, in order to be able to contain at least
// size byte.
// If size is greater than the limit defined via constructor, the latter is
// used.
// There is no need to call this method directly. It could be useful, however,
// for performance reasons. If the caller at some point knows the expected
// size, they could pre-expand the buffer in order to avoid multiple expensive
// grow-and-copy on every write.
func (r *RingBuffer) Grow(size int) error {
	newSize := len(r.buf)

	// multiply the buffer size by `expansionFactor`, until it is enough to
	// contain `size`.
	// Special case is buf size = 0, because it can't be multiplied.
	if newSize == 0 {
		newSize = 1
	}
	for newSize < size {
		newSize *= expansionFactor
	}

	// in any case a size bigger than defined cap is not allowed
	if newSize >= r.maxSize {
		newSize = r.maxSize
	}

	// create a new bigger slice and copy all the content from the old buffer
	// to the new
	newBuf, err := makeSlice(newSize)
	if err != nil {
		// grow can fail
		return err
	}
	copy(newBuf, r.buf)

	// new buffer is the new slice
	// the old one is no more referenced, so it could be collected.
	r.buf = newBuf

	return nil
}

// makeSlice allocates a slice of size n.
// If the allocation panics, this function recovers it and returns an error.
func makeSlice(n int) (b []byte, err error) {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			err = errors.New("can't make slice")
		}
	}()
	b = make([]byte, n)
	return
}

// Bytes returns the buffer content in a slice of bytes.
func (r *RingBuffer) Bytes() []byte {
	if r.ringMode {
		return append(r.buf[r.pos:], r.buf[:r.pos]...)
	}

	return r.buf[:r.pos]
}

// String returns the buffer content as a string.
// With this method RingBuffer implements the fmt.Stringer interface.
func (r *RingBuffer) String() string {
	return string(r.Bytes())
}

// Written returns the number of bytes written so far in the buffer.
func (r *RingBuffer) Written() int {
	return r.written
}

// Reset clears the buffer.
// The underlying slice is kept, so it keeps the same size reached so far.
// The `written` counter is reset too.
func (r *RingBuffer) Reset() {
	r.written = 0
	r.ringMode = false
	r.pos = 0
}

// NewRingBuffer creates and initialise a new RingBuffer using
// - initialSize as length of the pre-allocated underlying buffer (can be 0)
// - maxSize as maximum limit this buffer can reach.
//
// If initial is greater than cap, cap is used as size.
func NewRingBuffer(initialSize, maxSize int) *RingBuffer {
	if initialSize > maxSize {
		initialSize = maxSize
	}

	return &RingBuffer{
		buf:      make([]byte, initialSize, initialSize),
		written:  0,
		ringMode: false,
		pos:      0,
		maxSize:  maxSize,
	}
}
