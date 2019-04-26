package shm

import (
	"fmt"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/capnp/fixed_arena"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/cdi"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/command"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/job"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/quervo"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/reqrsppool"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/ring_buffer"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/rspreceiver"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/schemas/node/common"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/schemas/rv3io"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

const (
	maxInflightRequests = 16 * 1024
)

type Statistics struct {
	NumCommandsProduced     uint64
	NumCommandsConsumed     uint64
	NumAllocationRequests   uint64
	NumAllocationResponses  uint64
	NumJobRequests          uint64
	NumJobResponses         uint64
	NumDeallocationRequests uint64
}

type Context struct {
	logger                  logger.Logger
	consumerQuervo          *quervo.Quervo
	producerQuervo          *quervo.Quervo
	jobBlockAttachers       []*job.JobBlockAttacher
	cdi                     *cdi.Cdi
	channelInfo             *cdi.ChannelInfo
	responseReceiver        *rspreceiver.ResponseReceiver
	responseItemRingBuffer  *ring_buffer.RingBuffer
	pendingRequestResponses []*v3io.RequestResponse
	requestResponsePool     *reqrsppool.RequestResponsePool
	fixedArena              *fixed_arena.FixedArena
	inputEncoderByJobType   [job.TYPE_MAX]func(interface{}, *job.JobBlock) error
	outputDecoderByJobType  [job.TYPE_MAX]func(*job.JobBlock, *v3io.Response) error
	statistics              Statistics
}

func NewContext(parentLogger logger.Logger, daemonAddr string) (*Context, error) {
	var err error

	responseReceiver, err := rspreceiver.NewResponseReceiver(parentLogger)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create response receiver")
	}

	newContext := &Context{
		logger:                  parentLogger.GetChild("ctx").(logger.Logger),
		consumerQuervo:          quervo.NewQuervo(parentLogger, "consumer"),
		producerQuervo:          quervo.NewQuervo(parentLogger, "producer"),
		responseReceiver:        responseReceiver,
		responseItemRingBuffer:  ring_buffer.NewRingBuffer(maxInflightRequests),
		pendingRequestResponses: make([]*v3io.RequestResponse, maxInflightRequests),
		requestResponsePool:     reqrsppool.NewRequestResponsePool(maxInflightRequests),
	}

	newContext.cdi, err = cdi.NewCdi(newContext.logger, daemonAddr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create CDI")
	}

	// create a channel
	newContext.channelInfo, err = newContext.cdi.CreateChannel()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create channel")
	}

	if err = newContext.attachQuervos(); err != nil {
		return nil, errors.Wrap(err, "Failed to attach quervos")
	}

	if err = newContext.attachHeaps(); err != nil {
		return nil, errors.Wrap(err, "Failed to attach heaps")
	}

	// attach quervo to response receiver. all responses from quervo will be written to newContext.responseItemChan
	newContext.responseReceiver.RegisterQuervo(newContext.consumerQuervo, newContext.responseItemRingBuffer)

	// populate ID and request pool
	if err = newContext.populatePools(); err != nil {
		return nil, errors.Wrap(err, "Failed to populate pools")
	}

	// populate encoders/decoders
	newContext.populateEncoderDecoderLookup()

	// create a capnp arena that doesn't allocate
	newContext.fixedArena, err = fixed_arena.NewFixedArena()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create fixed arena")
	}

	// start the response receiver
	responseReceiver.Start()

	newContext.logger.DebugWith("Created", "channelInfo", newContext.channelInfo)

	return newContext, nil
}

func (c *Context) NewSession(sessionID uint32) (*Session, error) {
	return newSession(c.logger, c, sessionID)
}

func (c *Context) GetNextResponse() *v3io.Response {
	var response *v3io.Response

	// while we haven't received a response
	for response == nil {

		// wait for a response
		responseItem, _ := c.responseItemRingBuffer.Get()
		c.statistics.NumCommandsConsumed++

		// decode the command
		commandType := command.DecodeType(responseItem)

		switch commandType {
		case command.TYPE_ALLOCATION_RESPONSE:
			response = c.handleAllocationResponse(responseItem)

		case command.TYPE_JOB_RESPONSE:
			response = c.handleJobResponse(responseItem)

			//case command.TYPE_ECHO_RESPONSE:
			//	response = c.handleEchoResponse(responseItem)
		}
	}

	return response
}

func (c *Context) GetStatistics() *Statistics {
	return &c.statistics
}

func (c *Context) SessionAcquire(input *v3io.SessionAcquireInput, cookie interface{}) (*v3io.Request, error) {
	return c.createAndSubmitJob(input, cookie, job.TYPE_SESSION_ACQUIRE, 0, nil)
}

func (sc *Context) SessionAcquireSync(input *v3io.SessionAcquireInput) (*Session, error) {
	//if sc.context.numInflightRequests > 0 {
	//	return errors.New("Can't submit synchronous response while requests are in flight")
	//}

	response, err := sc.waitForSyncResponse(sc.SessionAcquire(input, nil))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to acquire session (sync)")
	}

	defer response.Release()

	// create session
	return sc.NewSession(response.Output.(*v3io.SessionAcquireOutput).SessionID)
}

func (c *Context) Echo(input *v3io.EchoInput, cookie interface{}) (*v3io.Request, error) {
	return c.createAndSubmitJob(input, cookie, job.TYPE_ECHO, 0, nil)
}

func (c *Context) EchoCommand(input *v3io.EchoInput, cookie interface{}) (*v3io.Request, error) {
	return c.createAndSubmitJob(input, cookie, job.TYPE_ECHO, 0, nil)
}

func (c *Context) encodeSessionAcquireInput(input interface{}, jobBlock *job.JobBlock) error {
	sessionAcquireInput := input.(*v3io.SessionAcquireInput)

	rv3ioRequest, err := c.createRv3ioRequest(jobBlock, nil, nil)
	if err != nil {
		return err
	}

	sessionAcquire, err := rv3ioRequest.NewSessionAcquire()
	if err != nil {
		return err
	}

	sessionAcquire.SetInterfaceType(node_common_capnp.InterfaceType(sessionAcquireInput.InterfaceType))

	if err = sessionAcquire.SetLabel(sessionAcquireInput.Label); err != nil {
		return err
	}

	if err = sessionAcquire.SetUserName(sessionAcquireInput.Username); err != nil {
		return err
	}

	if err = sessionAcquire.SetPassword(sessionAcquireInput.Password); err != nil {
		return err
	}

	// marshal the request
	c.marshalRv3ioRequest(jobBlock)

	return nil
}

func (c *Context) decodeSessionAcquireOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
	sessionAcquireOutput := v3io.SessionAcquireOutput{}

	// set the response header has the arena buffer
	message, err := c.fixedArena.SetDecodeBuffer(jobBlock.GetResponseHeaderBuffer())
	if err != nil {
		return err
	}

	// Extract the root struct from the message.
	rv3ioResponse, err := rv3io_capnp.ReadRootRv3ioResponse(message)
	if err != nil {
		return err
	}

	sessionAcquire, err := rv3ioResponse.SessionAcquire()
	if err != nil {
		return err
	}

	authSession, err := sessionAcquire.SessionId()
	if err != nil {
		return err
	}

	sessionAcquireOutput.SessionID = authSession.SessionId()

	// set the output (object gets allocated from heap)
	response.Output = &sessionAcquireOutput

	return nil
}

func (c *Context) encodeEchoInput(input interface{}, jobBlock *job.JobBlock) error {
	echoInput := input.(*v3io.EchoInput)

	// copy the request header
	copy(jobBlock.GetRequestHeaderBuffer(), echoInput.DataBytes)

	// update the size
	jobBlock.Header.RequestHeaderSectionSizeBytes = uint32(len(echoInput.DataBytes))

	return nil
}

func (c *Context) decodeEchoOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
	responseHeader := jobBlock.GetResponseHeaderBuffer()

	echoOutput := v3io.EchoOutput{
		DataBytes: make([]byte, len(responseHeader)),
	}

	// copy from payload to output
	copy(echoOutput.DataBytes, responseHeader)

	// set the output
	response.Output = echoOutput

	return nil
}

func (c *Context) createAndSubmitJob(input interface{},
	cookie interface{},
	jobType job.Type,
	payloadSizeBytes int,
	writer func(interface{}, *job.JobBlock)) (*v3io.Request, error) {

	// allocate request
	requestResponse := c.requestResponsePool.Get()

	// populate fields (input will escape meaning it will be allocated on heap)
	requestResponse.Request.JobType = jobType
	requestResponse.Request.Cookie = cookie
	requestResponse.Request.Input = input
	requestResponse.Request.Writer = writer
	requestResponse.Request.PayloadSizeWords = uint64(payloadSizeBytes / 8)

	// if user asks for a partial word, give him whole word (e.g. 10 bytes = two 8 byte words)
	if payloadSizeBytes&0x7 != 0 {
		requestResponse.Request.PayloadSizeWords++
	}

	// submit the request
	return c.submitRequestViaJob(&requestResponse.Request)
}

func (c *Context) createRv3ioRequest(jobBlock *job.JobBlock,
	sessionID *uint32,
	containerHandle *uint64) (rv3ioRequest rv3io_capnp.Rv3ioRequest, err error) {

	// assign the request header slot to the fixed arena
	segment, err := c.fixedArena.SetEncodeBuffer(jobBlock.GetRequestHeaderBuffer())
	if err != nil {
		return
	}

	rv3ioRequest, err = rv3io_capnp.NewRootRv3ioRequest(segment)
	if err != nil {
		return
	}

	if sessionID != nil {
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

		session.SetSessionId(*sessionID)
	}

	if containerHandle != nil {
		rv3ioRequest.SetContainerHandle(*containerHandle)
	}

	return
}

func (c *Context) marshalRv3ioRequest(jobBlock *job.JobBlock) {

	// marshal the request (just updates stuff in the header)
	buf := c.fixedArena.MarshalEncodeBuffer()

	// set used size
	jobBlock.Header.RequestHeaderSectionSizeBytes = uint32(len(buf))
}

func (c *Context) attachQuervos() error {

	if err := c.consumerQuervo.Attach(c.channelInfo.ProducerShmPath); err != nil {
		return errors.Wrap(err, "Failed to attach consumer quervo")
	}

	if err := c.producerQuervo.Attach(c.channelInfo.ConsumerShmPath); err != nil {
		return errors.Wrap(err, "Failed to attach producer quervo")
	}

	return nil
}

func (c *Context) attachHeaps() error {

	for _, heapShmPath := range c.channelInfo.HeapShmPaths {

		// create a job block attacher at the shm heap
		jobBlockAttacher, err := job.NewJobBlockAttacher(c.logger, heapShmPath, maxInflightRequests)
		if err != nil {
			return errors.Wrap(err, "Failed to create job block attacher")
		}

		c.jobBlockAttachers = append(c.jobBlockAttachers, jobBlockAttacher)
	}

	return nil
}

func (c *Context) populatePools() error {
	var requestID uint64

	for requestID = maxInflightRequests; requestID != 0; requestID-- {

		// create a request with said ID
		requestResponse := &v3io.RequestResponse{}
		requestResponse.Request.ID = requestID
		requestResponse.Request.RequestResponse = requestResponse
		requestResponse.Response.ID = requestID
		requestResponse.Response.RequestResponse = requestResponse
		requestResponse.Response.Releaser = c.releaseResponse

		// shove to request pool
		c.requestResponsePool.Put(requestResponse)
	}

	return nil
}

func (c *Context) handleAllocationResponse(responseItem uint64) *v3io.Response {
	var allocationResponse command.AllocationResponse
	var jobRequest command.JobRequest

	c.statistics.NumAllocationResponses++

	allocationResponse.Decode(responseItem)

	if allocationResponse.JobHandle == 0 {
		panic("Failed to allocate")
	}

	// get pending request/response object
	requestResponse := c.pendingRequestResponses[allocationResponse.ID]

	if requestResponse == nil {
		panic("Couldn't find requestResponse")
	}

	// get job block information from handle
	heapIdx := command.GetHeapIdxFromJobHandle(allocationResponse.JobHandle)
	offsetWords := command.GetOffsetWordsFromJobHandle(allocationResponse.JobHandle)

	// attach a job block
	jb := c.jobBlockAttachers[heapIdx].AttachJob(int(offsetWords))
	if jb == nil {
		panic("Failed to attach job")
	}

	// set the handle
	jb.Handle = allocationResponse.JobHandle

	// set the request's job block
	requestResponse.Request.JobBlock = jb

	// if the request has a writer, do that now
	if requestResponse.Request.Writer != nil {
		requestResponse.Request.Writer(requestResponse.Request.Input, jb)
	}

	// write the request to the job block
	c.writeRequestToJobBlock(&requestResponse.Request, jb)

	// submit the job
	jobRequest.JobHandle = allocationResponse.JobHandle
	jobRequest.SetID(allocationResponse.GetID())

	// produce
	c.producerQuervo.Produce(jobRequest.Encode())
	c.statistics.NumJobRequests++
	c.statistics.NumCommandsProduced++

	// no response
	return nil
}

func (c *Context) handleJobResponse(responseItem uint64) *v3io.Response {
	var jobResponse command.JobResponse
	jobResponse.Decode(responseItem)

	c.statistics.NumJobResponses++

	// get pending request response object and nullify
	requestResponse := c.pendingRequestResponses[jobResponse.ID]
	c.pendingRequestResponses[jobResponse.ID] = nil

	if requestResponse.Request.JobBlock == nil {
		panic("Invalid job block")
	}

	// set job block in response
	requestResponse.Response.JobBlock = requestResponse.Request.JobBlock
	requestResponse.Response.JobType = requestResponse.Request.JobType

	// read the response header into the output structure (allocates on infrequent requests)
	if err := c.readResponseFromJobBlock(&requestResponse.Response, requestResponse.Response.JobBlock); err != nil {
		requestResponse.Response.Err = err
	}

	// return the response
	return &requestResponse.Response
}

func (c *Context) submitRequestViaJob(request *v3io.Request) (*v3io.Request, error) {

	// save request in pending requests based on id
	c.pendingRequestResponses[request.ID] = request.RequestResponse

	// initialize allocation command
	allocationRequestCommand := command.AllocationRequest{
		JobType:          request.JobType,
		PayloadSizeWords: request.PayloadSizeWords,
	}

	allocationRequestCommand.SetID(request.ID)

	// allocate a job block
	c.producerQuervo.Produce(allocationRequestCommand.Encode())
	c.statistics.NumCommandsProduced++
	c.statistics.NumAllocationRequests++

	return request, nil
}

func (c *Context) writeRequestToJobBlock(request *v3io.Request, jobBlock *job.JobBlock) error {

	// call the encoder
	return c.inputEncoderByJobType[request.JobType](request.Input, jobBlock)
}

func (c *Context) readResponseFromJobBlock(response *v3io.Response, jobBlock *job.JobBlock) error {

	// if there's an error, create an error for it
	if jobBlock.Header.Result != 0 {
		return fmt.Errorf("v3io returned error: %d", jobBlock.Header.Result)
	}

	// get decoder
	decoder := c.outputDecoderByJobType[response.JobType]

	// call the decoder if one specified (some jobs, e.g. file-close, don't need to decode anything)
	if decoder != nil {
		return decoder(jobBlock, response)
	}

	return nil
}

func (c *Context) releaseResponse(response *v3io.Response) {

	// initialize allocation command
	deallocationRequestCommand := command.DeallocationRequest{
		JobHandle: response.JobBlock.Handle,
	}

	deallocationRequestCommand.SetID(response.ID)

	// allocate a job block
	c.producerQuervo.Produce(deallocationRequestCommand.Encode())
	c.statistics.NumCommandsProduced++
	c.statistics.NumDeallocationRequests++

	// release the request/response
	c.requestResponsePool.Put(response.RequestResponse)

	// release job
	c.jobBlockAttachers[response.JobBlock.Header.HeapIdx].ReleaseJob(response.JobBlock)
}

func (c *Context) populateEncoderDecoderLookup() {

	codecs := map[job.Type]struct {
		encoder func(interface{}, *job.JobBlock) error
		decoder func(*job.JobBlock, *v3io.Response) error
	}{
		job.TYPE_SESSION_ACQUIRE: {c.encodeSessionAcquireInput, c.decodeSessionAcquireOutput},
		job.TYPE_ECHO:            {c.encodeEchoInput, c.decodeEchoOutput},
	}

	for jobType, codecs := range codecs {
		c.inputEncoderByJobType[jobType] = codecs.encoder
		c.outputDecoderByJobType[jobType] = codecs.decoder
	}
}

func (c *Context) waitForSyncResponse(request *v3io.Request, submitError error) (*v3io.Response, error) {
	if submitError != nil {
		return nil, errors.Wrap(submitError, "Failed to submit")
	}

	// get the next response
	response := c.GetNextResponse()

	// verify response is that of the request
	if &response.RequestResponse.Request != request {
		return nil, errors.New("Got unexpected response")
	}

	return response, nil
}