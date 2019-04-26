package inplace

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/iguazio/go-capnproto2"
)

type FileReadTestSuite struct {
	suite.Suite
}

func (suite *FileReadTestSuite) TestRequestEncode() {

	//
	// Encode with inplace
	//

	newFileRead, err := NewFileReadRequest()
	suite.NoError(err)

	fileReadInplaceBuffer := make([]byte, 1024)

	newFileRead.SetBuffer(fileReadInplaceBuffer)

	*newFileRead.ContainerHandle = 0x1111111111111111
	*newFileRead.SessionID = 0x22222222
	*newFileRead.FileHandle = 0x3333333333333333
	*newFileRead.Offset = 0x4444444444444444
	*newFileRead.BytesCount = 0x5555555555555555

	//
	// Encode with standard
	//
	fileReadStandardBuffer := make([]byte, 1024)

	message, segment, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	rv3ioRequest, _ := rv3io_capnp.NewRootRv3ioRequest(segment)

	rv3ioRequest.SetContainerHandle(0x1111111111111111)

	var auth rv3io_capnp.Rv3ioAuthorization
	var session rv3io_capnp.Rv3ioAuthSession

	auth, _ = rv3ioRequest.NewAuthorization()
	session, _ = auth.NewSession()
	session.SetSessionId(0x22222222)

	fileRead, _ := rv3ioRequest.NewFileRead()
	fileRead.SetFileHandle(0x3333333333333333)
	fileRead.SetOffset(0x4444444444444444)
	fileRead.SetBytesCount(0x5555555555555555)

	buf := bytes.NewBuffer(fileReadStandardBuffer[0:0:1024])
	capnp.NewEncoder(buf).Encode(message)

	//
	// Compare
	//

	suite.Equal(fileReadInplaceBuffer, fileReadStandardBuffer)
}

func (suite *FileReadTestSuite) TestRequestDecode() {

	//
	// Encode with standard
	//
	fileReadStandardBuffer := make([]byte, 1024)

	message, segment, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	rv3ioResponse, _ := rv3io_capnp.NewRootRv3ioResponse(segment)

	fileRead, _ := rv3ioResponse.NewFileRead()
	fileRead.SetBytesRead(0x1111111111111111)

	buf := bytes.NewBuffer(fileReadStandardBuffer[0:0:1024])
	capnp.NewEncoder(buf).Encode(message)

	//
	// Decode with inplace
	//

	newFileRead, err := NewFileReadResponse()
	suite.NoError(err)

	newFileRead.SetBuffer(fileReadStandardBuffer)

	suite.Equal(uint64(0x1111111111111111), *newFileRead.BytesRead)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFileReadTestSuite(t *testing.T) {
	suite.Run(t, new(FileReadTestSuite))
}

func BenchmarkFileRead(b *testing.B) {
	newFileRead, _ := NewFileReadRequest()
	fileReadBuffer := make([]byte, 1024)

	for n := 0; n < b.N; n++ {
		newFileRead.SetBuffer(fileReadBuffer[0:1024:1024])

		*newFileRead.ContainerHandle = 0x1111111111111111
		*newFileRead.SessionID = 0x22222222
		*newFileRead.FileHandle = 0x3333333333333333
		*newFileRead.Offset = 0x4444444444444444
		*newFileRead.BytesCount = 0x5555555555555555
	}
}
