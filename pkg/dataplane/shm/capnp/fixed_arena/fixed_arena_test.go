package fixed_arena

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/v3io/v3io-go/internal/schemas/rv3io"

	"github.com/stretchr/testify/suite"
	"github.com/iguazio/go-capnproto2"
)

type CapnTestSuite struct {
	suite.Suite
}

func (suite *CapnTestSuite) TestSingleSegment() {
	return

	msg, segment, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		panic("1")
	}

	s, _ := msg.Segment(0)
	fmt.Println(hex.Dump(s.Data()))

	rv3ioRequest, err := rv3io_capnp.NewRootRv3ioRequest(segment)
	if err != nil {
		panic("2")
	}

	fmt.Println(hex.Dump(s.Data()))

	rv3ioRequest.SetContainerHandle(0xCCBB)

	fmt.Println(hex.Dump(s.Data()))
}

func (suite *CapnTestSuite) TestFixed() {
	arena, err := NewFixedArena()
	if err != nil {
		panic("1")
	}

	buf := make([]byte, 1024)

	s, err := arena.SetEncodeBuffer(buf)
	if err != nil {
		panic("2")
	}

	rv3ioRequest, err := rv3io_capnp.NewRootRv3ioRequest(s)
	if err != nil {
		panic("2")
	}

	fmt.Println(hex.Dump(s.Data()))

	rv3ioRequest.SetContainerHandle(0xCCBB)

	fmt.Println(hex.Dump(s.Data()))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
//func TestAdapterTestSuite(t *testing.T) {
//	suite.Run(t, new(CapnTestSuite))
//}

func BenchmarkFileRead(b *testing.B) {

	arena, err := NewFixedArena()
	if err != nil {
		panic("1")
	}

	buf := make([]byte, 1024)

	for n := 0; n < b.N; n++ {

		s, err := arena.SetEncodeBuffer(buf[0:1024:1024])
		if err != nil {
			panic("2")
		}

		rv3ioRequest, err := rv3io_capnp.NewRootRv3ioRequest(s)
		if err != nil {
			panic("2")
		}

		rv3ioRequest.SetContainerHandle(0xAAAAAAAAAAAAAAAA)

		var auth rv3io_capnp.Rv3ioAuthorization
		var session rv3io_capnp.Rv3ioAuthSession

		auth, err = rv3ioRequest.NewAuthorization()
		if err != nil {
			return
		}

		session, err = auth.NewSession()
		if err != nil {
			return
		}

		session.SetSessionId(0xBBBBBBBB)

		fileRead, err := rv3ioRequest.NewFileRead()
		if err != nil {
			panic("4")
		}

		fileRead.SetFileHandle(0xCCCCCCCCCCCCCCCC)
		fileRead.SetOffset(0xDDDDDDDDDDDDDDDD)
		fileRead.SetBytesCount(0xEEEEEEEEEEEEEEEE)

		// marshal the request (just updates stuff in the header)
		arena.MarshalEncodeBuffer()
	}
}

//func BenchmarkFileRead(b *testing.B) {
//	_buf := make([]byte, 1024)
//
//	for n := 0; n < b.N; n++ {
//
//		msg, s, err := capnp.NewMessage(capnp.SingleSegment(nil))
//		if err != nil {
//			panic("1")
//		}
//
//		rv3ioRequest, err := rv3io_capnp.NewRootRv3ioRequest(s)
//		if err != nil {
//			panic("2")
//		}
//
//		rv3ioRequest.SetContainerHandle(0xCCBB)
//
//		var auth rv3io_capnp.Rv3ioAuthorization
//		var session rv3io_capnp.Rv3ioAuthSession
//
//		auth, err = rv3ioRequest.NewAuthorization()
//		if err != nil {
//			return
//		}
//
//		session, err = auth.NewSession()
//		if err != nil {
//			return
//		}
//
//		session.SetSessionId(4)
//
//		fileRead, err := rv3ioRequest.NewFileRead()
//		if err != nil {
//			panic("4")
//		}
//
//		fileRead.SetFileHandle(0x1234)
//		fileRead.SetOffset(0x123)
//		fileRead.SetBytesCount(0x132)
//
//		buf := bytes.NewBuffer(_buf[0:1024:1024])
//		capnp.NewEncoder(buf).Encode(msg)
//	}
//}
