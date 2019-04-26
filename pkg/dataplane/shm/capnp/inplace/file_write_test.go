package inplace

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/v3io/v3io-go/internal/schemas/rv3io"

	"github.com/stretchr/testify/suite"
	"github.com/iguazio/go-capnproto2"
)

type FileWriteTestSuite struct {
	suite.Suite
}

func (suite *FileWriteTestSuite) TestRequestEncode() {

	//
	// Encode with inplace
	//

	newFileWrite, err := NewFileWriteRequest()
	suite.NoError(err)

	fileWriteInplaceBuffer := make([]byte, 1024)

	newFileWrite.SetBuffer(fileWriteInplaceBuffer)

	*newFileWrite.ContainerHandle = 0x1111111111111111
	*newFileWrite.SessionID = 0x22222222
	*newFileWrite.FileHandle = 0x3333333333333333
	*newFileWrite.Offset = 0x4444444444444444
	*newFileWrite.BytesCount = 0x5555555555555555

	//
	// Encode with standard
	//
	fileWriteStandardBuffer := make([]byte, 1024)

	message, segment, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	rv3ioRequest, _ := rv3io_capnp.NewRootRv3ioRequest(segment)

	rv3ioRequest.SetContainerHandle(0x1111111111111111)

	var auth rv3io_capnp.Rv3ioAuthorization
	var session rv3io_capnp.Rv3ioAuthSession

	auth, _ = rv3ioRequest.NewAuthorization()
	session, _ = auth.NewSession()
	session.SetSessionId(0x22222222)

	fileWrite, _ := rv3ioRequest.NewFileWrite()
	fileWrite.SetFileHandle(0x3333333333333333)
	fileWrite.SetOffset(0x4444444444444444)
	fileWrite.SetBytesCount(0x5555555555555555)

	buf := bytes.NewBuffer(fileWriteStandardBuffer[0:0:1024])
	capnp.NewEncoder(buf).Encode(message)

	//
	// Compare
	//

	fmt.Println(hex.Dump(fileWriteInplaceBuffer))

	suite.Equal(fileWriteInplaceBuffer, fileWriteStandardBuffer)
}

func (suite *FileWriteTestSuite) TestRequestDecode() {

	//
	// Encode with standard
	//
	fileWriteStandardBuffer := make([]byte, 1024)

	message, segment, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	rv3ioResponse, _ := rv3io_capnp.NewRootRv3ioResponse(segment)

	fileWrite, _ := rv3ioResponse.NewFileWrite()
	fileWrite.SetBytesWritten(0x1111111111111111)

	buf := bytes.NewBuffer(fileWriteStandardBuffer[0:0:1024])
	capnp.NewEncoder(buf).Encode(message)

	//
	// Decode with inplace
	//

	newFileWrite, err := NewFileWriteResponse()
	suite.NoError(err)

	newFileWrite.SetBuffer(fileWriteStandardBuffer)

	suite.Equal(uint64(0x1111111111111111), *newFileWrite.BytesWritten)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFileWriteTestSuite(t *testing.T) {
	suite.Run(t, new(FileWriteTestSuite))
}

func BenchmarkFileWrite(b *testing.B) {
	newFileWrite, _ := NewFileWriteRequest()
	fileWriteBuffer := make([]byte, 1024)

	for n := 0; n < b.N; n++ {
		newFileWrite.SetBuffer(fileWriteBuffer[0:1024:1024])

		*newFileWrite.ContainerHandle = 0x1111111111111111
		*newFileWrite.SessionID = 0x22222222
		*newFileWrite.FileHandle = 0x3333333333333333
		*newFileWrite.Offset = 0x4444444444444444
		*newFileWrite.BytesCount = 0x5555555555555555
	}
}
