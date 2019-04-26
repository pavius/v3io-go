package quervo

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/v3io/v3io-go/pkg/dataplane/shm/mapped_memory"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/memory_accessor"

	"github.com/nuclio/logger"
	"github.com/pkg/errors"
)

const (
	fifoPathOffset         = 0
	fifoPathLengthOffset   = 248
	itemsOffset            = 2048
	producerNumItemsOffset = 256
	producerMaskOffset     = 260
	producerHeadOffset     = 264
	producerTailOffset     = 268
	consumerNumItemsOffset = 320
	consumerMaskOffset     = 324
	consumerHeadOffset     = 328
	consumerTailOffset     = 332
	consumerStateOffset    = 336
)

type Quervo struct {
	shmPath      string
	shmMmap      *io.ReaderAt
	fifoPath     string
	mappedMemory *mapped_memory.MappedMemory
	logger       logger.Logger
	items        *memory_accessor.MemoryAccessor

	producer struct {
		numItems *uint32
		mask     *uint32
		head     *uint32
		tail     *uint32
	}

	consumer struct {
		numItems *uint32
		mask     *uint32
		head     *uint32
		tail     *uint32
		state    *uint64
	}
}

func NewQuervo(parentLogger logger.Logger, name string) *Quervo {
	return &Quervo{
		logger: parentLogger.GetChild(fmt.Sprintf("quervo-%s", name)).(logger.Logger),
	}
}

func (q *Quervo) Attach(shmPath string) error {
	var err error

	// try to mmap the file
	q.mappedMemory, err = mapped_memory.NewMappedMemory(shmPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to open shared memory @ %s", shmPath)
	}

	// initialize the members after having initialized the memory accessor
	q.initializeMembers()

	return nil
}

func (q *Quervo) Produce(item uint64) bool {
	var producerHead, consumerTail uint32
	var producerNext, freeEntries uint32
	var mask uint32

	mask = *q.producer.mask

	producerHead = atomic.LoadUint32(q.producer.head)
	consumerTail = atomic.LoadUint32(q.consumer.tail)

	// get the number of free entries
	freeEntries = mask + (consumerTail - producerHead)
	if freeEntries == 0 {
		return false
	}

	// calculate and set producer head
	producerNext = producerHead + 1
	atomic.StoreUint32(q.producer.head, producerNext)

	// write the entry in the ring @ producer_head & mask
	q.items.WriteUint64(int((producerHead&mask)*8), item)

	// set the tail of the producer to next
	atomic.StoreUint32(q.producer.tail, producerNext)

	// success
	return true
}

func (q *Quervo) Consume() (uint64, bool) {
	var consumerHead, producerTail uint32
	var consumerNext, entries uint32
	var mask uint32
	var item uint64

	mask = *q.producer.mask

	consumerHead = atomic.LoadUint32(q.consumer.head)
	producerTail = atomic.LoadUint32(q.producer.tail)

	// check how much stuff is pending in the queue
	entries = producerTail - consumerHead
	if entries == 0 {
		return 0, false
	}

	consumerNext = consumerHead + 1
	atomic.StoreUint32(q.consumer.head, consumerNext)

	// dequeue item
	item = q.items.ReadUint64(int((consumerHead & mask) * 8))

	atomic.StoreUint32(q.consumer.tail, consumerNext)

	// success
	return item, true
}

func (q *Quervo) initializeMembers() {

	// create an accessor on the mmapped data
	memoryAccessor := memory_accessor.NewMemoryAccessor(q.logger, q.mappedMemory.GetData())

	// initialize pointers to the data we'll be reading/writing
	q.producer.numItems = memoryAccessor.AtUint32(producerNumItemsOffset)
	q.producer.mask = memoryAccessor.AtUint32(producerMaskOffset)
	q.producer.head = memoryAccessor.AtUint32(producerHeadOffset)
	q.producer.tail = memoryAccessor.AtUint32(producerTailOffset)
	q.consumer.numItems = memoryAccessor.AtUint32(consumerNumItemsOffset)
	q.consumer.mask = memoryAccessor.AtUint32(consumerMaskOffset)
	q.consumer.head = memoryAccessor.AtUint32(consumerHeadOffset)
	q.consumer.tail = memoryAccessor.AtUint32(consumerTailOffset)
	q.consumer.state = memoryAccessor.AtUint64(consumerStateOffset)

	// read the FIFO path length
	fifoPathLength := memoryAccessor.ReadUint64(fifoPathLengthOffset)

	// read the FIFO path
	q.fifoPath = string(memoryAccessor.AtBytes(fifoPathOffset, int(fifoPathLength)))

	// create an items memory accessor with which we can read items
	q.items = memory_accessor.NewMemoryAccessor(q.logger, q.mappedMemory.GetData()[itemsOffset:])
}
