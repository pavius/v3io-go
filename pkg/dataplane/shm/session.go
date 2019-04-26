package shm

import (
	"strconv"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/job"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/schemas/rv3io"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type Session struct {
	logger    logger.Logger
	sessionID uint32
	context   *Context
}

func newSession(parentLogger logger.Logger, context *Context, sessionID uint32) (*Session, error) {
	log := parentLogger.GetChild("session").(logger.Logger)

	log.DebugWith("Session acquired", "sessionID", sessionID)

	newSession := &Session{
		logger:    log,
		sessionID: sessionID,
		context:   context,
	}

	// populate encoders/decoders
	newSession.populateEncoderDecoderLookup()

	return newSession, nil
}

func (c *Container) GetContext() *Context {
	return c.context
}

func (s *Session) NewContainer(containerHandler uint64) (*Container, error) {
	return newContainer(s.logger, s, containerHandler)
}

func (s *Session) ContainerOpen(input *v3io.ContainerOpenInput, cookie interface{}) (*v3io.Request, error) {
	return s.context.createAndSubmitJob(input, cookie, job.TYPE_CONTAINER_OPEN, 0, nil)
}

func (s *Session) ContainerOpenSync(input *v3io.ContainerOpenInput) (*Container, error) {
	//if sc.context.numInflightRequests > 0 {
	//	return errors.New("Can't submit synchronous response while requests are in flight")
	//}

	response, err := s.context.waitForSyncResponse(s.ContainerOpen(input, nil))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open container (sync)")
	}

	defer response.Release()

	// create session
	return s.NewContainer(response.Output.(*v3io.ContainerOpenOutput).ContainerHandle)
}

func (s *Session) encodeContainerOpenInput(input interface{}, jobBlock *job.JobBlock) error {
	containerOpenInput := input.(*v3io.ContainerOpenInput)

	rv3ioRequest, err := s.createRv3ioRequest(jobBlock, nil)
	if err != nil {
		return err
	}

	containerOpen, err := rv3ioRequest.NewContainerOpen()
	if err != nil {
		return err
	}

	if err = containerOpen.SetContainerId(strconv.Itoa(int(containerOpenInput.ContainerID))); err != nil {
		return err
	}

	// marshal the request
	s.context.marshalRv3ioRequest(jobBlock)

	return nil
}

func (s *Session) decodeContainerOpenOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
	containerOpenOutput := v3io.ContainerOpenOutput{}

	// set the response header has the arena buffer
	message, err := s.context.fixedArena.SetDecodeBuffer(jobBlock.GetResponseHeaderBuffer())
	if err != nil {
		return err
	}

	// Extract the root struct from the message.
	rv3ioResponse, err := rv3io_capnp.ReadRootRv3ioResponse(message)
	if err != nil {
		return err
	}

	containerOpen, err := rv3ioResponse.ContainerOpen()
	if err != nil {
		return err
	}

	containerOpenOutput.ContainerHandle = containerOpen.ContainerHandle()

	// set the output (object gets allocated from heap)
	response.Output = &containerOpenOutput

	return nil
}

func (s *Session) createRv3ioRequest(jobBlock *job.JobBlock,
	containerHandle *uint64) (rv3ioRequest rv3io_capnp.Rv3ioRequest, err error) {

	return s.context.createRv3ioRequest(jobBlock, &s.sessionID, containerHandle)
}

func (s *Session) populateEncoderDecoderLookup() {

	codecs := map[job.Type]struct {
		encoder func(interface{}, *job.JobBlock) error
		decoder func(*job.JobBlock, *v3io.Response) error
	}{
		job.TYPE_CONTAINER_OPEN: {s.encodeContainerOpenInput, s.decodeContainerOpenOutput},
	}

	for jobType, codecs := range codecs {
		s.context.inputEncoderByJobType[jobType] = codecs.encoder
		s.context.outputDecoderByJobType[jobType] = codecs.decoder
	}
}