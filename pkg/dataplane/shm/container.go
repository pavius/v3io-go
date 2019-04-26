package shm

import (
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/capnp/inplace"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/job"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/schemas/rv3io"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type Container struct {
	logger            logger.Logger
	containerHandle   uint64
	session           *Session
	context           *Context
	fileReadRequest   *inplace.FileReadRequest
	fileReadResponse  *inplace.FileReadResponse
	fileWriteRequest  *inplace.FileWriteRequest
	fileWriteResponse *inplace.FileWriteResponse
}

func newContainer(parentLogger logger.Logger, session *Session, containerHandle uint64) (*Container, error) {
	log := parentLogger.GetChild("container").(logger.Logger)

	log.DebugWith("Container opened", "handle", containerHandle)

	newContainer := &Container{
		logger:          log,
		containerHandle: containerHandle,
		session:         session,
		context:         session.context,
	}

	// populate encoders/decoders
	newContainer.populateEncoderDecoderLookup()

	// create inplace encoder / decoder
	if err := newContainer.createInplaceEncoderDecoders(); err != nil {
		return nil, errors.Wrap(err, "Failed to create inplace encoder/decoder")
	}

	return newContainer, nil
}

func (c *Container) GetSession() *Session {
	return c.session
}

func (c *Container) FileOpen(input *v3io.FileOpenInput, cookie interface{}) (*v3io.Request, error) {
	return c.context.createAndSubmitJob(input, cookie, job.TYPE_FILE_OPEN, 0, nil)
}

func (c *Container) FileOpenSync(input *v3io.FileOpenInput) (*v3io.FileOpenOutput, error) {
	//if c.context.numInflightRequests > 0 {
	//	return errors.New("Can't submit synchronous response while requests are in flight")
	//}

	response, err := c.session.context.waitForSyncResponse(c.FileOpen(input, nil))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open file (sync)")
	}

	defer response.Release()

	// return the response
	return response.Output.(*v3io.FileOpenOutput), nil
}

func (c *Container) FileClose(input *v3io.FileCloseInput, cookie interface{}) (*v3io.Request, error) {
	return c.context.createAndSubmitJob(input, cookie, job.TYPE_FILE_CLOSE, 0, nil)
}

func (c *Container) FileCloseSync(input *v3io.FileCloseInput) error {
	//if c.context.numInflightRequests > 0 {
	//	return errors.New("Can't submit synchronous response while requests are in flight")
	//}

	response, err := c.session.context.waitForSyncResponse(c.FileClose(input, nil))
	if err != nil {
		return errors.Wrap(err, "Failed to close file (sync)")
	}

	response.Release()

	return nil
}

func (c *Container) FileWrite(input *v3io.FileWriteInput, cookie interface{}) (*v3io.Request, error) {
	return c.context.createAndSubmitJob(input, cookie, job.TYPE_FILE_WRITE, 16*1024, input.Writer)
}

func (c *Container) FileWriteSync(input *v3io.FileWriteInput) (*v3io.FileWriteOutput, error) {
	//if c.context.numInflightRequests > 0 {
	//	return errors.New("Can't submit synchronous response while requests are in flight")
	//}

	response, err := c.session.context.waitForSyncResponse(c.FileWrite(input, nil))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open file (sync)")
	}

	defer response.Release()

	// return the response
	return response.Output.(*v3io.FileWriteOutput), nil
}

func (c *Container) FileRead(input *v3io.FileReadInput, cookie interface{}) (*v3io.Request, error) {
	return c.context.createAndSubmitJob(input, cookie, job.TYPE_FILE_READ, int(input.BytesCount), nil)
}

func (c *Container) FileReadSync(input *v3io.FileReadInput) (*v3io.FileReadOutput, error) {
	//if c.context.numInflightRequests > 0 {
	//	return errors.New("Can't submit synchronous response while requests are in flight")
	//}

	response, err := c.session.context.waitForSyncResponse(c.FileRead(input, nil))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open file (sync)")
	}

	defer response.Release()

	// return the response
	return response.Output.(*v3io.FileReadOutput), nil
}

func (c *Container) encodeFileOpenInput(input interface{}, jobBlock *job.JobBlock) error {
	fileOpenInput := input.(*v3io.FileOpenInput)

	rv3ioRequest, err := c.createRv3ioRequest(jobBlock)
	if err != nil {
		return err
	}

	fileOpen, err := rv3ioRequest.NewFileOpen()
	if err != nil {
		return err
	}

	fileOpen.SetMode(uint64(fileOpenInput.Mode))
	fileOpen.SetOflags(uint64(fileOpenInput.Flags))

	if err = fileOpen.SetPath(fileOpenInput.FilePath); err != nil {
		return err
	}

	// marshal the request
	c.context.marshalRv3ioRequest(jobBlock)

	return nil
}

func (c *Container) decodeFileOpenOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
	fileOpenOutput := v3io.FileOpenOutput{}

	// set the response header has the arena buffer
	message, err := c.context.fixedArena.SetDecodeBuffer(jobBlock.GetResponseHeaderBuffer())
	if err != nil {
		return err
	}

	// Extract the root struct from the message.
	rv3ioResponse, err := rv3io_capnp.ReadRootRv3ioResponse(message)
	if err != nil {
		return err
	}

	fileOpen, err := rv3ioResponse.FileOpen()
	if err != nil {
		return err
	}

	fileOpenOutput.FileHandle = fileOpen.FileHandle()

	// set the output (object gets allocated from heap)
	response.Output = &fileOpenOutput

	return nil
}

func (c *Container) encodeFileCloseInput(input interface{}, jobBlock *job.JobBlock) error {
	fileCloseInput := input.(*v3io.FileCloseInput)

	rv3ioRequest, err := c.createRv3ioRequest(jobBlock)
	if err != nil {
		return err
	}

	fileClose, err := rv3ioRequest.NewFileClose()
	if err != nil {
		return err
	}

	fileClose.SetFileHandle(fileCloseInput.FileHandle)

	// marshal the request
	c.context.marshalRv3ioRequest(jobBlock)

	return nil
}

func (c *Container) encodeFileWriteInput(input interface{}, jobBlock *job.JobBlock) error {
	fileWriteInput := input.(*v3io.FileWriteInput)

	// use inplace
	c.fileWriteRequest.SetBuffer(jobBlock.GetRequestHeaderBuffer())

	*c.fileWriteRequest.SessionID = c.session.sessionID
	*c.fileWriteRequest.ContainerHandle = c.containerHandle
	*c.fileWriteRequest.FileHandle = fileWriteInput.FileHandle
	*c.fileWriteRequest.Offset = fileWriteInput.Offset

	jobBlock.Header.RequestHeaderSectionSizeBytes = uint32(c.fileWriteRequest.Len())

	if fileWriteInput.Data != nil {
		*c.fileWriteRequest.BytesCount = uint64(len(fileWriteInput.Data))

		// write the data
		payloadBuffer := jobBlock.GetPayloadBufferForWrite()
		copy(payloadBuffer, fileWriteInput.Data)

		// update the size
		jobBlock.Header.PayloadSectionSizeBytes = uint64(len(fileWriteInput.Data))
	}

	return nil
}

func (c *Container) decodeFileWriteOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
	fileWriteOutput := v3io.FileWriteOutput{}

	// use inplace
	c.fileWriteResponse.SetBuffer(jobBlock.GetResponseHeaderBuffer())

	fileWriteOutput.BytesWritten = *c.fileWriteResponse.BytesWritten

	// set the output (object gets allocated from heap)
	response.Output = &fileWriteOutput

	return nil
}

func (c *Container) encodeFileReadInput(input interface{}, jobBlock *job.JobBlock) error {
	fileReadInput := input.(*v3io.FileReadInput)

	// use inplace
	c.fileReadRequest.SetBuffer(jobBlock.GetRequestHeaderBuffer())

	*c.fileReadRequest.SessionID = c.session.sessionID
	*c.fileReadRequest.ContainerHandle = c.containerHandle
	*c.fileReadRequest.FileHandle = fileReadInput.FileHandle
	*c.fileReadRequest.Offset = fileReadInput.Offset
	*c.fileReadRequest.BytesCount = fileReadInput.BytesCount

	jobBlock.Header.RequestHeaderSectionSizeBytes = uint32(c.fileReadRequest.Len())

	return nil
}

func (c *Container) decodeFileReadOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
	fileReadOutput := v3io.FileReadOutput{}

	// use inplace
	c.fileReadResponse.SetBuffer(jobBlock.GetResponseHeaderBuffer())

	// point to the data
	fileReadOutput.Data = jobBlock.GetPayloadBufferForRead()

	// set the output (object gets allocated from heap)
	response.Output = &fileReadOutput

	return nil
}

func (c *Container) createRv3ioRequest(jobBlock *job.JobBlock) (rv3ioRequest rv3io_capnp.Rv3ioRequest, err error) {

	return c.session.createRv3ioRequest(jobBlock, &c.containerHandle)
}

func (c *Container) populateEncoderDecoderLookup() {

	codecs := map[job.Type]struct {
		encoder func(interface{}, *job.JobBlock) error
		decoder func(*job.JobBlock, *v3io.Response) error
	}{
		job.TYPE_FILE_OPEN:  {c.encodeFileOpenInput, c.decodeFileOpenOutput},
		job.TYPE_FILE_CLOSE: {c.encodeFileCloseInput, nil},
		job.TYPE_FILE_WRITE: {c.encodeFileWriteInput, c.decodeFileWriteOutput},
		job.TYPE_FILE_READ:  {c.encodeFileReadInput, c.decodeFileReadOutput},
	}

	for jobType, codecs := range codecs {
		c.context.inputEncoderByJobType[jobType] = codecs.encoder
		c.context.outputDecoderByJobType[jobType] = codecs.decoder
	}
}

func (c *Container) createInplaceEncoderDecoders() error {
	var err error

	c.fileReadRequest, err = inplace.NewFileReadRequest()
	if err != nil {
		return errors.Wrap(err, "Failed to create file read request")
	}

	c.fileReadResponse, err = inplace.NewFileReadResponse()
	if err != nil {
		return errors.Wrap(err, "Failed to create file read response")
	}

	c.fileWriteRequest, err = inplace.NewFileWriteRequest()
	if err != nil {
		return errors.Wrap(err, "Failed to create file write request")
	}

	c.fileWriteResponse, err = inplace.NewFileWriteResponse()
	if err != nil {
		return errors.Wrap(err, "Failed to create file write response")
	}

	return nil
}
