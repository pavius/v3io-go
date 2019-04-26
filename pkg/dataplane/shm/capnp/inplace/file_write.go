package inplace

import (
	"unsafe"

	"github.com/v3io/v3io-go/pkg/dataplane/shm/capnp/fixed_arena"
)

//
// Request
//

type FileWriteRequest struct {
	template          []byte
	templateSizeBytes int
	buffer            []byte
	SessionID         *uint32
	ContainerHandle   *uint64
	FileHandle        *uint64
	Offset            *uint64
	BytesCount        *uint64
}

func NewFileWriteRequest() (*FileWriteRequest, error) {
	newFileWriteRequest := &FileWriteRequest{
		template: make([]byte, 2048),
	}
	templateLen := len(newFileWriteRequest.template)

	arena, err := fixed_arena.NewFixedArena()
	if err != nil {
		return nil, err
	}

	segment, err := arena.SetEncodeBuffer(newFileWriteRequest.template[0:templateLen:templateLen])
	if err != nil {
		return nil, err
	}

	rv3ioRequest, err := rv3io_capnp.NewRootRv3ioRequest(segment)
	if err != nil {
		return nil, err
	}

	rv3ioRequest.SetContainerHandle(0xAAAAAAAAAAAAAAAA)

	var auth rv3io_capnp.Rv3ioAuthorization
	var session rv3io_capnp.Rv3ioAuthSession

	auth, err = rv3ioRequest.NewAuthorization()
	if err != nil {
		return nil, err
	}

	session, err = auth.NewSession()
	if err != nil {
		return nil, err
	}

	session.SetSessionId(0xBBBBBBBB)

	fileWrite, err := rv3ioRequest.NewFileWrite()
	if err != nil {
		return nil, err
	}

	fileWrite.SetFileHandle(0xCCCCCCCCCCCCCCCC)
	fileWrite.SetOffset(0xDDDDDDDDDDDDDDDD)
	fileWrite.SetBytesCount(0xEEEEEEEEEEEEEEEE)

	// marshal the request (just updates stuff in the header)
	arena.MarshalEncodeBuffer()

	// set length of template (+8 for header)
	newFileWriteRequest.templateSizeBytes = len(segment.Data()) + 8

	return newFileWriteRequest, nil
}

func (fr *FileWriteRequest) SetBuffer(buffer []byte) {
	fr.buffer = buffer

	// copy the template over
	copy(buffer, fr.template[:fr.templateSizeBytes])

	// update pointers
	fr.ContainerHandle = (*uint64)(unsafe.Pointer(&fr.buffer[16]))
	fr.SessionID = (*uint32)(unsafe.Pointer(&fr.buffer[80]))
	fr.FileHandle = (*uint64)(unsafe.Pointer(&fr.buffer[88]))
	fr.Offset = (*uint64)(unsafe.Pointer(&fr.buffer[96]))
	fr.BytesCount = (*uint64)(unsafe.Pointer(&fr.buffer[112]))
}

func (fr *FileWriteRequest) Len() int {
	return fr.templateSizeBytes
}

//
// Response
//

type FileWriteResponse struct {
	buffer       []byte
	BytesWritten *uint64
}

func NewFileWriteResponse() (*FileWriteResponse, error) {
	return &FileWriteResponse{}, nil
}

func (fr *FileWriteResponse) SetBuffer(buffer []byte) {
	fr.buffer = buffer

	fr.BytesWritten = (*uint64)(unsafe.Pointer(&buffer[64]))
}
