package rspreceiver

import (
	"runtime"

	"github.com/v3io/v3io-go/pkg/dataplane/shm/quervo"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/ring_buffer"

	"github.com/nuclio/logger"
)

// number of iterations of empty quervos to do other things
const MAX_CONSECUTIVE_EMPTY_QUERVOS = 100000

type controlRequestKind int

const (
	CONTROL_REQUEST_KIND_REGISTER_QUERVO = iota
	CONTROL_REQUEST_KIND_UNREGISTER_QUERVO
)

type controlRequest struct {
	kind controlRequestKind
	info registeredQuervo // TODO: inherit
}

type registeredQuervo struct {
	quervo         *quervo.Quervo
	itemRingBuffer *ring_buffer.RingBuffer
}

type ResponseReceiver struct {
	logger             logger.Logger
	registeredQuervos  []registeredQuervo
	controlRequestChan chan *controlRequest
	done               bool
}

func NewResponseReceiver(parentLogger logger.Logger) (*ResponseReceiver, error) {
	return &ResponseReceiver{
		logger:             parentLogger.GetChild("response_receiver").(logger.Logger),
		registeredQuervos:  []registeredQuervo{},
		controlRequestChan: make(chan *controlRequest),
		done:               false,
	}, nil
}

func (rr *ResponseReceiver) Start() {
	go rr.receive()
}

func (rr *ResponseReceiver) RegisterQuervo(quervo *quervo.Quervo,
	itemRingBuffer *ring_buffer.RingBuffer) {
	rr.controlRequestChan <- &controlRequest{
		kind: CONTROL_REQUEST_KIND_REGISTER_QUERVO,
		info: registeredQuervo{quervo, itemRingBuffer},
	}
}

func (rr *ResponseReceiver) receive() {
	rr.logger.Debug("Starting to receive")

	// lock to a thread - we don't want to be disturbed
	runtime.LockOSThread()

	// while we're not done
	for !rr.done {

		// see if there's any request to handle
		rr.handleControlRequests()

		// see if there are any responses from quervo
		rr.handleQuervoResponses()
	}

	rr.logger.Debug("Done receiving")
}

func (rr *ResponseReceiver) handleControlRequests() {
	controlRequestChanEmpty := false

	// handle requests until channel is empty
	for !controlRequestChanEmpty {

		select {
		case controlRequest := <-rr.controlRequestChan:
			rr.handleControlRequest(controlRequest)
		default:
			controlRequestChanEmpty = true
		}
	}
}

func (rr *ResponseReceiver) handleControlRequest(request *controlRequest) {

	switch request.kind {
	case CONTROL_REQUEST_KIND_REGISTER_QUERVO:
		rr.registeredQuervos = append(rr.registeredQuervos, request.info)
	}
}

func (rr *ResponseReceiver) handleQuervoResponses() {
	consecutiveEmptyQuervos := MAX_CONSECUTIVE_EMPTY_QUERVOS

	// cache length locally, it won't change between invocations
	numRegisteredQuervos := len(rr.registeredQuervos)

	// don't bother if there are no quervos
	if numRegisteredQuervos == 0 {
		return
	}

	// if consecutiveEmptyQuervos is zero, it means we didn't read from all the quervos in
	// MAX_CONSECUTIVE_EMPTY_QUERVOS iterations
	for consecutiveEmptyQuervos != 0 {

		// iterate over all quervos
		for registeredQuervoIdx := 0; registeredQuervoIdx < numRegisteredQuervos; registeredQuervoIdx++ {
			registeredQuervo := rr.registeredQuervos[registeredQuervoIdx]

			// don't read too deeply into a single quervo
			for quervoItemIdx := 0; quervoItemIdx < 128; quervoItemIdx++ {

				// try to consume an item
				item, received := registeredQuervo.quervo.Consume()

				// if there's nothing to read, stop looking @ this quervo
				if !received {
					consecutiveEmptyQuervos--
					break
				}

				// we got an item so we need to reset consecutive empty quervos
				consecutiveEmptyQuervos = MAX_CONSECUTIVE_EMPTY_QUERVOS

				// post the item to the appropriate channel
				registeredQuervo.itemRingBuffer.Offer(item)
			}
		}
	}
}
