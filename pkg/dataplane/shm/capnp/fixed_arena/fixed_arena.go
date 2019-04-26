package fixed_arena

import (
	"fmt"
	"unsafe"

	capnp "github.com/iguazio/go-capnproto2"
)

const (
	wordSize = 8 // from capan mem.go

	// Header size, fixed since we have one segment
	hdrSize = wordSize
)

// FixedArena implement Arena interface with fixed underlying buffer
// The FixedArena is meant to be reusable
type FixedArena struct {
	buf []byte
	seg *capnp.Segment
	msg *capnp.Message
}

// NewFixedArena create a new FixedArena using buf as underlying buffer
func NewFixedArena() (*FixedArena, error) {
	fa := &FixedArena{}
	fa.msg = &capnp.Message{Arena: fa}
	fa.seg = fa.msg.InitializeFixed()

	return fa, nil
}

// Buffer return the underlying buffer
func (fa *FixedArena) Buffer() []byte {
	return fa.buf
}

// SetBuffer sets the underlying buffer
func (fa *FixedArena) SetEncodeBuffer(buf []byte) (*capnp.Segment, error) {
	if len(buf) < hdrSize {
		return nil, fmt.Errorf("buffer too small (%d)", len(buf))
	}

	fa.buf = buf

	fa.seg.SetData(fa.buf[hdrSize : hdrSize*2])
	return fa.seg, nil
}

// Marshal return byte representation of the structure
func (fa *FixedArena) MarshalEncodeBuffer() []byte {
	size := len(fa.seg.Data())
	*(*uint64)(unsafe.Pointer(&fa.buf[0])) = 0

	p := (*uint32)(unsafe.Pointer(&fa.buf[4]))
	*p = uint32(size / wordSize)
	return fa.buf[:size+hdrSize]
}

// SetBuffer sets the underlying buffer
func (fa *FixedArena) SetDecodeBuffer(buf []byte) (*capnp.Message, error) {
	if len(buf) < hdrSize {
		return nil, fmt.Errorf("buffer too small (%d)", len(buf))
	}

	fa.buf = buf

	fa.seg.SetData(fa.buf[hdrSize:])
	return fa.msg, nil
}

// Segment returns the segment
func (fa *FixedArena) Segment() *capnp.Segment {
	return fa.seg
}

// Message return the message
func (fa *FixedArena) Message() *capnp.Message {
	return fa.msg
}

// Arena interface

// NumSegments returns the number of segments in the arena.
// This must not be larger than 1<<32.
func (fa *FixedArena) NumSegments() int64 {
	return 1
}

// Data loads the data for the segment with the given ID.
func (fa *FixedArena) Data(id capnp.SegmentID) ([]byte, error) {
	if id != 0 {
		return nil, fmt.Errorf("unknown segment - %d", id)
	}
	if fa.buf == nil {
		return nil, fmt.Errorf("empty buffer")
	}
	return fa.buf[hdrSize:hdrSize], nil
}

// Allocate allocates a byte slice such that cap(data) - len(data) >= minsz.
// segs is a map of already loaded segments keyed by ID.  The arena may
// return an existing segment's ID, in which case the arena is responsible
// for copying the existing data to the returned byte slice.  Allocate must
// not modify the segments passed into it.
func (fa *FixedArena) Allocate(minsz capnp.Size, segs map[capnp.SegmentID]*capnp.Segment) (capnp.SegmentID, []byte, error) {

	return 0, nil, fmt.Errorf("buffer too small (%d < %d)", cap(fa.buf), len(fa.buf)+int(minsz))
}
