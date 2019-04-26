package cdi

import (
	"bufio"
	"bytes"
	"fmt"
	"net"

	"github.com/nuclio/logger"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/schemas/rv3io"

	"github.com/pkg/errors"
	"github.com/iguazio/go-capnproto2"
)

type ChannelInfo struct {
	ChannelID       uint64
	WorkerID        uint64
	ConsumerShmPath string
	ProducerShmPath string
	HeapShmPaths    []string
}

type Cdi struct {
	logger           logger.Logger
	conn             net.Conn
	daemonListenAddr string
}

func NewCdi(parentLogger logger.Logger, daemonListenAddr string) (*Cdi, error) {
	var err error

	newCdi := Cdi{
		logger:           parentLogger.GetChild("cdi").(logger.Logger),
		daemonListenAddr: daemonListenAddr,
	}

	newCdi.conn, err = net.Dial("udp", daemonListenAddr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed create UDP socket")
	}

	newCdi.logger.DebugWith("Created", "daemonListenAddr", daemonListenAddr)

	return &newCdi, nil
}

func (cdi *Cdi) CreateChannel() (*ChannelInfo, error) {
	var channelInfo ChannelInfo

	populate := func(rv3ioRequest *rv3io_capnp.Rv3ioRequest) error {
		channelCreate, err := rv3ioRequest.NewChannelCreate()
		if err != nil {
			return errors.Wrap(err, "Failed to create channel create in rv3io")
		}

		channelCreate.SetConsumerNumQuervoItems(128 * 1024)
		channelCreate.SetProducerNumQuervoItems(128 * 1024)
		channelCreate.SetWorkerPreference(uint64(0xFFFFFFFFFFFFFFFF))

		return nil
	}

	parse := func(rv3ioResponse *rv3io_capnp.Rv3ioResponse) error {
		channelCreate, err := rv3ioResponse.ChannelCreate()
		if err != nil {
			return errors.Wrap(err, "Failed to get session acquire")
		}

		channelInfo.ChannelID = channelCreate.ChannelId()
		channelInfo.WorkerID = channelCreate.WorkerId()
		channelInfo.ConsumerShmPath, _ = channelCreate.ConsumerShmPath()
		channelInfo.ProducerShmPath, _ = channelCreate.ProducerShmPath()

		// prefix with /dev/shm
		shmPath := ""
		channelInfo.ConsumerShmPath = shmPath + channelInfo.ConsumerShmPath
		channelInfo.ProducerShmPath = shmPath + channelInfo.ProducerShmPath

		heapShmPaths, _ := channelCreate.HeapShmPaths()

		for heapShmIdx := 0; heapShmIdx < heapShmPaths.Len(); heapShmIdx++ {
			heapShmPath, _ := heapShmPaths.At(heapShmIdx)
			channelInfo.HeapShmPaths = append(channelInfo.HeapShmPaths, shmPath+heapShmPath)
		}

		cdi.logger.DebugWith("Channel created",
			"ConsumerShmPath", channelInfo.ConsumerShmPath,
			"ProducerShmPath", channelInfo.ProducerShmPath,
			"heapShmPath", channelInfo.HeapShmPaths,
			"ChannelID", channelInfo.ChannelID,
			"WorkerID", channelInfo.WorkerID)

		return nil
	}

	if err := cdi.sendRequestToDaemon(populate, parse); err != nil {
		return nil, errors.Wrap(err, "Failed to send request to daemon")
	}

	return &channelInfo, nil
}

func (cdi *Cdi) DeleteChannel(ChannelID uint64) error {
	populate := func(rv3ioRequest *rv3io_capnp.Rv3ioRequest) error {
		channelDelete, err := rv3ioRequest.NewChannelDelete()
		if err != nil {
			return errors.Wrap(err, "Failed to create channel create in rv3io")
		}

		channelDelete.SetChannelId(ChannelID)

		return nil
	}

	parse := func(rv3ioResponse *rv3io_capnp.Rv3ioResponse) error {
		return nil
	}

	if err := cdi.sendRequestToDaemon(populate, parse); err != nil {
		return errors.Wrap(err, "Failed to send request to daemon")
	}

	return nil
}

func (cdi *Cdi) sendRequestToDaemon(populateRequest func(rv3ioRequest *rv3io_capnp.Rv3ioRequest) error,
	parseResponse func(rv3ioResponse *rv3io_capnp.Rv3ioResponse) error) error {

	msg, segment, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return errors.Wrap(err, "Failed to create capn message")
	}

	rv3ioRequest, err := rv3io_capnp.NewRootRv3ioRequest(segment)
	if err != nil {
		return errors.Wrap(err, "Failed to create root rv3io request")
	}

	// populate the request
	err = populateRequest(&rv3ioRequest)
	if err != nil {
		return errors.Wrap(err, "Failed to populate request")
	}

	// prepare the request stream
	buf := bytes.NewBuffer([]byte{})
	capnp.NewEncoder(buf).Encode(msg)

	// send outwards to daemon
	written, err := fmt.Fprint(cdi.conn, buf)
	if err != nil || written != len(buf.Bytes()) {
		return errors.Wrap(err, "Failed to write request to daemon")
	}

	// Read the message
	msg, err = capnp.NewDecoder(bufio.NewReader(cdi.conn)).Decode()
	if err != nil {
		return errors.Wrap(err, "Failed to decode response")
	}

	// Extract the root struct from the message.
	rv3ioResponse, err := rv3io_capnp.ReadRootRv3ioResponse(msg)
	if err != nil {
		return errors.Wrap(err, "Failed to read rv3io response")
	}

	// check response code
	if rv3ioResponse.Result() != 0 {
		return errors.Wrapf(err, "CDI request failed with result %d", rv3ioResponse.Result())
	}

	// parse the response
	err = parseResponse(&rv3ioResponse)
	if err != nil {
		return errors.Wrap(err, "Failed to parse response")
	}

	return nil
}
