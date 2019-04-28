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

type context struct {
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
	statistics              v3io.Statistics
}

func NewContext(parentLogger logger.Logger, daemonAddr string) (*context, error) {
	var err error

	responseReceiver, err := rspreceiver.NewResponseReceiver(parentLogger)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create response receiver")
	}

	newcontext := &context{
		logger:                  parentLogger.GetChild("ctx").(logger.Logger),
		consumerQuervo:          quervo.NewQuervo(parentLogger, "consumer"),
		producerQuervo:          quervo.NewQuervo(parentLogger, "producer"),
		responseReceiver:        responseReceiver,
		responseItemRingBuffer:  ring_buffer.NewRingBuffer(maxInflightRequests),
		pendingRequestResponses: make([]*v3io.RequestResponse, maxInflightRequests),
		requestResponsePool:     reqrsppool.NewRequestResponsePool(maxInflightRequests),
	}

	newcontext.cdi, err = cdi.NewCdi(newcontext.logger, daemonAddr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create CDI")
	}

	// create a channel
	newcontext.channelInfo, err = newcontext.cdi.CreateChannel()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create channel")
	}

	if err = newcontext.attachQuervos(); err != nil {
		return nil, errors.Wrap(err, "Failed to attach quervos")
	}

	if err = newcontext.attachHeaps(); err != nil {
		return nil, errors.Wrap(err, "Failed to attach heaps")
	}

	// attach quervo to response receiver. all responses from quervo will be written to newcontext.responseItemChan
	newcontext.responseReceiver.RegisterQuervo(newcontext.consumerQuervo, newcontext.responseItemRingBuffer)

	// populate ID and request pool
	if err = newcontext.populatePools(); err != nil {
		return nil, errors.Wrap(err, "Failed to populate pools")
	}

	// populate encoders/decoders
	newcontext.populateEncoderDecoderLookup()

	// create a capnp arena that doesn't allocate
	newcontext.fixedArena, err = fixed_arena.NewFixedArena()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create fixed arena")
	}

	// start the response receiver
	responseReceiver.Start()

	newcontext.logger.DebugWith("Created", "channelInfo", newcontext.channelInfo)

	return newcontext, nil
}

func (c *context) NewSession(input *v3io.NewSessionInput) (v3io.Session, error) {

	// TODO: pass username/password/access key
	return newSession(c.logger, c)
}

func (c *context) GetNextResponse() *v3io.Response {
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

func (c *context) GetStatistics() *v3io.Statistics {
	return &c.statistics
}

func (c *context) SessionAcquire(input *v3io.SessionAcquireInput, cookie interface{}) (*v3io.Request, error) {
	return c.createAndSubmitJob(input, cookie, job.TYPE_SESSION_ACQUIRE, 0, nil)
}

func (sc *context) SessionAcquireSync(input *v3io.SessionAcquireInput) (v3io.Session, error) {
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

func (c *context) Echo(input *v3io.EchoInput, cookie interface{}) (*v3io.Request, error) {
	return c.createAndSubmitJob(input, cookie, job.TYPE_ECHO, 0, nil)
}

func (c *context) EchoCommand(input *v3io.EchoInput, cookie interface{}) (*v3io.Request, error) {
	return c.createAndSubmitJob(input, cookie, job.TYPE_ECHO, 0, nil)
}

// GetContainers
func (c *context) GetContainers(*v3io.GetContainersInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// GetContainersSync
func (c *context) GetContainersSync(*v3io.GetContainersInput) (*v3io.Response, error) {
	return nil, nil
}

// GetContainers
func (c *context) GetContainerContents(*v3io.GetContainerContentsInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// GetContainerContentsSync
func (c *context) GetContainerContentsSync(*v3io.GetContainerContentsInput) (*v3io.Response, error) {
	return nil, nil
}

// GetObject
func (c *context) GetObject(*v3io.GetObjectInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// GetObjectSync
func (c *context) GetObjectSync(*v3io.GetObjectInput) (*v3io.Response, error) {
	return nil, nil
}

// PutObject
func (c *context) PutObject(*v3io.PutObjectInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// PutObjectSync
func (c *context) PutObjectSync(*v3io.PutObjectInput) error {
	return nil
}

// DeleteObject
func (c *context) DeleteObject(*v3io.DeleteObjectInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// DeleteObjectSync
func (c *context) DeleteObjectSync(*v3io.DeleteObjectInput) error {
	return nil
}

// GetItem
func (c *context) GetItem(*v3io.GetItemInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// GetItemSync
func (c *context) GetItemSync(*v3io.GetItemInput) (*v3io.Response, error) {
	return nil, nil
}

// GetItems
func (c *context) GetItems(*v3io.GetItemsInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// GetItemSync
func (c *context) GetItemsSync(*v3io.GetItemsInput) (*v3io.Response, error) {
	return nil, nil
}

// PutItem
func (c *context) PutItem(*v3io.PutItemInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// PutItemSync
func (c *context) PutItemSync(*v3io.PutItemInput) error {
	return nil
}

// PutItems
func (c *context) PutItems(*v3io.PutItemsInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// PutItemsSync
func (c *context) PutItemsSync(*v3io.PutItemsInput) (*v3io.Response, error) {
	return nil, nil
}

// UpdateItem
func (c *context) UpdateItem(*v3io.UpdateItemInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// UpdateItemSync
func (c *context) UpdateItemSync(*v3io.UpdateItemInput) error {
	return nil
}

// CreateStream
func (c *context) CreateStream(*v3io.CreateStreamInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// CreateStreamSync
func (c *context) CreateStreamSync(*v3io.CreateStreamInput) error {
	return nil
}

// DeleteStream
func (c *context) DeleteStream(*v3io.DeleteStreamInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// DeleteStreamSync
func (c *context) DeleteStreamSync(*v3io.DeleteStreamInput) error {
	return nil
}

// SeekShard
func (c *context) SeekShard(*v3io.SeekShardInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// SeekShardSync
func (c *context) SeekShardSync(*v3io.SeekShardInput) (*v3io.Response, error) {
	return nil, nil
}

// PutRecords
func (c *context) PutRecords(*v3io.PutRecordsInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// PutRecordsSync
func (c *context) PutRecordsSync(*v3io.PutRecordsInput) (*v3io.Response, error) {
	return nil, nil
}

// GetRecords
func (c *context) GetRecords(*v3io.GetRecordsInput, interface{}, chan *v3io.Response) (*v3io.Request, error) {
	return nil, nil
}

// GetRecordsSync
func (c *context) GetRecordsSync(*v3io.GetRecordsInput) (*v3io.Response, error) {
	return nil, nil
}


func (c *context) encodeSessionAcquireInput(input interface{}, jobBlock *job.JobBlock) error {
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

func (c *context) decodeSessionAcquireOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
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

func (c *context) encodeEchoInput(input interface{}, jobBlock *job.JobBlock) error {
	echoInput := input.(*v3io.EchoInput)

	// copy the request header
	copy(jobBlock.GetRequestHeaderBuffer(), echoInput.DataBytes)

	// update the size
	jobBlock.Header.RequestHeaderSectionSizeBytes = uint32(len(echoInput.DataBytes))

	return nil
}

func (c *context) decodeEchoOutput(jobBlock *job.JobBlock, response *v3io.Response) error {
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

func (c *context) createAndSubmitJob(input interface{},
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

func (c *context) createRv3ioRequest(jobBlock *job.JobBlock,
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

func (c *context) marshalRv3ioRequest(jobBlock *job.JobBlock) {

	// marshal the request (just updates stuff in the header)
	buf := c.fixedArena.MarshalEncodeBuffer()

	// set used size
	jobBlock.Header.RequestHeaderSectionSizeBytes = uint32(len(buf))
}

func (c *context) attachQuervos() error {

	if err := c.consumerQuervo.Attach(c.channelInfo.ProducerShmPath); err != nil {
		return errors.Wrap(err, "Failed to attach consumer quervo")
	}

	if err := c.producerQuervo.Attach(c.channelInfo.ConsumerShmPath); err != nil {
		return errors.Wrap(err, "Failed to attach producer quervo")
	}

	return nil
}

func (c *context) attachHeaps() error {

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

func (c *context) populatePools() error {
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

func (c *context) handleAllocationResponse(responseItem uint64) *v3io.Response {
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

func (c *context) handleJobResponse(responseItem uint64) *v3io.Response {
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

func (c *context) submitRequestViaJob(request *v3io.Request) (*v3io.Request, error) {

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

func (c *context) writeRequestToJobBlock(request *v3io.Request, jobBlock *job.JobBlock) error {

	// call the encoder
	return c.inputEncoderByJobType[request.JobType](request.Input, jobBlock)
}

func (c *context) readResponseFromJobBlock(response *v3io.Response, jobBlock *job.JobBlock) error {

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

func (c *context) releaseResponse(response *v3io.Response) {

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

func (c *context) populateEncoderDecoderLookup() {

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

func (c *context) waitForSyncResponse(request *v3io.Request, submitError error) (*v3io.Response, error) {
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
