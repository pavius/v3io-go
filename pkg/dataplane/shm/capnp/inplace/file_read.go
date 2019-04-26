package inplace

import (
	"unsafe"

	"github.com/v3io/v3io-go/pkg/dataplane/shm/capnp/fixed_arena"
)

//
// Request
//

type FileReadRequest struct {
	template          []byte
	templateSizeBytes int
	buffer            []byte
	SessionID         *uint32
	ContainerHandle   *uint64
	FileHandle        *uint64
	Offset            *uint64
	BytesCount        *uint64
}

func NewFileReadRequest() (*FileReadRequest, error) {
	newFileReadRequest := &FileReadRequest{
		template: make([]byte, 2048),
	}
	templateLen := len(newFileReadRequest.template)

	arena, err := fixed_arena.NewFixedArena()
	if err != nil {
		return nil, err
	}

	segment, err := arena.SetEncodeBuffer(newFileReadRequest.template[0:templateLen:templateLen])
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

	fileRead, err := rv3ioRequest.NewFileRead()
	if err != nil {
		return nil, err
	}

	fileRead.SetFileHandle(0xCCCCCCCCCCCCCCCC)
	fileRead.SetOffset(0xDDDDDDDDDDDDDDDD)
	fileRead.SetBytesCount(0xEEEEEEEEEEEEEEEE)

	// marshal the request (just updates stuff in the header)
	arena.MarshalEncodeBuffer()

	// set length of template (+8 for header)
	newFileReadRequest.templateSizeBytes = len(segment.Data()) + 8

	return newFileReadRequest, nil
}

func (fr *FileReadRequest) SetBuffer(buffer []byte) {
	fr.buffer = buffer

	// copy the template over
	copy(buffer, fr.template[:fr.templateSizeBytes])

	// update pointers
	fr.ContainerHandle = (*uint64)(unsafe.Pointer(&fr.buffer[16]))
	fr.SessionID = (*uint32)(unsafe.Pointer(&fr.buffer[80]))
	fr.FileHandle = (*uint64)(unsafe.Pointer(&fr.buffer[88]))
	fr.Offset = (*uint64)(unsafe.Pointer(&fr.buffer[96]))
	fr.BytesCount = (*uint64)(unsafe.Pointer(&fr.buffer[104]))
}

func (fr *FileReadRequest) Len() int {
	return fr.templateSizeBytes
}

//
// Response
//

type FileReadResponse struct {
	buffer    []byte
	BytesRead *uint64
}

func NewFileReadResponse() (*FileReadResponse, error) {
	return &FileReadResponse{}, nil
}

func (fr *FileReadResponse) SetBuffer(buffer []byte) {
	fr.buffer = buffer

	fr.BytesRead = (*uint64)(unsafe.Pointer(&buffer[64]))
}
