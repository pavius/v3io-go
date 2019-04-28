package main

import (
	"github.com/nuclio/zap"
	"runtime"
	"sync"
	"bytes"
	"fmt"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane/shm"
	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/logger"
	"github.com/nuclio/errors"
)

type echoer struct {
	logger       logger.Logger
	dispatcher   *shm.Dispatcher
	context      v3io.Context
	ioDepth      uint64
	requestIndex int
}

func NewEchoer(parentLogger logger.Logger,
	dispatcher *shm.Dispatcher,
	context v3io.Context,
	ioDepth uint64) (*echoer, error) {

	newEchoer := &echoer{
		logger:     parentLogger.GetChild("echoer").(logger.Logger),
		ioDepth:    ioDepth,
		dispatcher: dispatcher,
		context:    context,
	}

	newEchoer.logger.DebugWith("Created", "ioDepth", ioDepth)

	return newEchoer, nil
}

func (e *echoer) Echo() error {
	e.logger.InfoWith("Starting to send echoes")

	if err := e.sendNextEcho(); err != nil {
		return errors.Wrap(err, "Failed to send next echo")
	}

	return nil
}

func (e *echoer) onEchoResponse(response *v3io.Response) error {
	if response.Err != nil {
		panic("Failed to echo")
	}

	// verify the response
	echoOutput := response.Output.(v3io.EchoOutput)
	got := echoOutput.DataBytes
	expected := e.getCurrentPayload()

	if bytes.Compare(got, expected) != 0 {
		panic(fmt.Sprintf("Echo mismatch. Got (%s), expected (%s)", got, expected))
	}

	// release the response
	response.Release()

	// send the next echo
	e.sendNextEcho()

	return nil
}

func (e *echoer) sendNextEcho() error {

	// increment request index
	e.requestIndex++

	echoInput := v3io.EchoInput{
		DataBytes: e.getCurrentPayload(),
	}

	// submit a request
	echoRequest, err := e.context.Echo(&echoInput, e)
	if err != nil {
		return errors.Wrap(err, "Failed submit file write")
	}

	e.dispatcher.RegisterResponseHandler(echoRequest.ID, e.onEchoResponse)

	return nil
}

func (e *echoer) getCurrentPayload() []byte {
	return []byte(fmt.Sprintf("Echoing: %d", e.requestIndex))
}

//
// cmd
//

func pollStatistics(loggerInstance logger.Logger, context v3io.Context) {

	prevStatistics := *context.GetStatistics()

	for {
		statistics := context.GetStatistics()

		loggerInstance.DebugWith("Gathered stats",
			"jobrsp", statistics.NumJobResponses-prevStatistics.NumJobResponses,
		)

		prevStatistics = *statistics

		time.Sleep(1 * time.Second)
	}
}

func run(ioDepth int) error {
	var err error

	loggerInstance, err := nucliozap.NewNuclioZapTest("test")
	if err != nil {
		loggerInstance.ErrorWith("Failed to create logger", "err", err)
	}

	context, err := shm.NewContext(loggerInstance, "127.0.0.1:1967")
	if err != nil {
		return errors.Wrap(err, "Failed to create context")
	}

	dispatcher, err := shm.NewDispatcher(loggerInstance, context, ioDepth)
	if err != nil {
		return errors.Wrap(err, "Failed to create dispatcher")
	}

	// poll stats in the background
	go pollStatistics(loggerInstance, context)

	echoer, err := NewEchoer(loggerInstance, dispatcher, context, uint64(ioDepth))
	if err != nil {
		loggerInstance.ErrorWith("Failed to create echoer", "err", err)
	}

	if err := echoer.Echo(); err != nil {
		return errors.Wrap(err, "Failed to echo")
	}

	if err := dispatcher.Dispatch(); err != nil {
		return err
	}

	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		err := run(32)

		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	wg.Wait()
}
