package job

import (
	"unsafe"

	"github.com/v3io/v3io-go/pkg/dataplane/shm/mapped_memory"
	"github.com/v3io/v3io-go/pkg/dataplane/shm/memory_accessor"

	"github.com/nuclio/logger"
	"github.com/pkg/errors"
)

const JOB_BLOCK_MAGIC = 0x19677691

// offsets. TODO: verify
const (
	JOB_BLOCK_MAGIC_NUMBER_OFFSET                           = 0
	JOB_BLOCK_STATE_OFFSET                                  = JOB_BLOCK_MAGIC_NUMBER_OFFSET + 4
	JOB_BLOCK_NUM_PIGGYBACKED_ALLOCATIONS_OFFSET            = JOB_BLOCK_STATE_OFFSET + 2
	JOB_BLOCK_HEAP_INDEX_OFFSET                             = JOB_BLOCK_NUM_PIGGYBACKED_ALLOCATIONS_OFFSET + 1
	JOB_BLOCK_JOB_SECTION_OFFSET_WORDS_OFFSET               = JOB_BLOCK_HEAP_INDEX_OFFSET + 1
	JOB_BLOCK_JOB_SECTION_MAX_SIZE_WORDS_OFFSET             = JOB_BLOCK_JOB_SECTION_OFFSET_WORDS_OFFSET + 2
	JOB_BLOCK_JOB_SECTION_SIZE_BYTES_OFFSET                 = JOB_BLOCK_JOB_SECTION_MAX_SIZE_WORDS_OFFSET + 2
	JOB_BLOCK_REQUEST_HEADER_SECTION_OFFSET_WORDS_OFFSET    = JOB_BLOCK_JOB_SECTION_SIZE_BYTES_OFFSET + 4
	JOB_BLOCK_REQUEST_HEADER_SECTION_MAX_SIZE_WORDS_OFFSET  = JOB_BLOCK_REQUEST_HEADER_SECTION_OFFSET_WORDS_OFFSET + 2
	JOB_BLOCK_REQUEST_HEADER_SECTION_SIZE_BYTES_OFFSET      = JOB_BLOCK_REQUEST_HEADER_SECTION_MAX_SIZE_WORDS_OFFSET + 2
	JOB_BLOCK_RESPONSE_HEADER_SECTION_OFFSET_WORDS_OFFSET   = JOB_BLOCK_REQUEST_HEADER_SECTION_SIZE_BYTES_OFFSET + 4
	JOB_BLOCK_RESPONSE_HEADER_SECTION_MAX_SIZE_WORDS_OFFSET = JOB_BLOCK_RESPONSE_HEADER_SECTION_OFFSET_WORDS_OFFSET + 2
	JOB_BLOCK_RESPONSE_HEADER_SECTION_SIZE_BYTES_OFFSET     = JOB_BLOCK_RESPONSE_HEADER_SECTION_MAX_SIZE_WORDS_OFFSET + 2
	JOB_BLOCK_PAYLOAD_SECTION_OFFSET_WORDS_OFFSET           = JOB_BLOCK_RESPONSE_HEADER_SECTION_SIZE_BYTES_OFFSET + 4
	JOB_BLOCK_PAYLOAD_SECTION_MAX_SIZE_WORDS_OFFSET         = JOB_BLOCK_PAYLOAD_SECTION_OFFSET_WORDS_OFFSET + 4
	JOB_BLOCK_PAYLOAD_SECTION_SIZE_BYTES_OFFSET             = JOB_BLOCK_PAYLOAD_SECTION_MAX_SIZE_WORDS_OFFSET + 4
	JOB_BLOCK_JOB_TYPE_OFFSET                               = JOB_BLOCK_PAYLOAD_SECTION_SIZE_BYTES_OFFSET + 8
	JOB_BLOCK_SIZE_WORDS_OFFSET                             = JOB_BLOCK_JOB_TYPE_OFFSET + 4
	JOB_BLOCK_RESULT_OFFSET                                 = JOB_BLOCK_SIZE_WORDS_OFFSET + 4
	JOB_BLOCK_PADDING_2_OFFSET                              = JOB_BLOCK_RESULT_OFFSET + 4
	JOB_BLOCK_PIGGYBACKED_DEALLOCATIONS_OFFSET              = JOB_BLOCK_PADDING_2_OFFSET + 4
	JOB_BLOCK_DATA_OFFSET                                   = JOB_BLOCK_PIGGYBACKED_DEALLOCATIONS_OFFSET + 256
)

type JobBlockSection struct {
	offsetWords  *uint16
	MaxSizeWords *uint16
	SizeBytes    *uint32
}

type JobBlock struct {
	raw    []byte
	Header *JobBlockHeader
	Handle uint64
}

type JobBlockHeader struct {
	MagicNumber                       uint32
	State                             uint16
	NumPiggybackedAllocations         uint8
	HeapIdx                           uint8
	JobSectionOffsetWords             uint16
	JobSectionMaxSizeWords            uint16
	JobSectionSizeBytes               uint32
	RequestHeaderSectionOffsetWords   uint16
	RequestHeaderSectionMaxSizeWords  uint16
	RequestHeaderSectionSizeBytes     uint32
	ResponseHeaderSectionOffsetWords  uint16
	ResponseHeaderSectionMaxSizeWords uint16
	ResponseHeaderSectionSizeBytes    uint32
	PayloadSectionOffsetWords         uint32
	PayloadSectionMaxSizeWords        uint32
	PayloadSectionSizeBytes           uint64
	JobType                           uint32
	BlockSizeWords                    uint32
	Result                            uint32
	Padding                           uint32
}

func (jb *JobBlock) attach(memoryAccessor *memory_accessor.MemoryAccessor, offsetBytes int) {

	// cast structure
	jb.Header = (*JobBlockHeader)(unsafe.Pointer(memoryAccessor.AtUint8(offsetBytes)))

	// before anything, verify magic
	if jb.Header.MagicNumber != JOB_BLOCK_MAGIC {
		panic("Got job block with invalid magic")
	}

	// create a slice off of the data
	jb.raw = memoryAccessor.AtBytes(offsetBytes, int(jb.Header.BlockSizeWords)*8)
}

func (jb *JobBlock) GetRequestHeaderBuffer() []byte {
	offsetBytes := (int(jb.Header.RequestHeaderSectionOffsetWords) * 8) + JOB_BLOCK_DATA_OFFSET
	y := offsetBytes + int(jb.Header.RequestHeaderSectionMaxSizeWords)*8

	return jb.raw[offsetBytes:y:y]
}

func (jb *JobBlock) GetRequestHeaderBufferForRead() []byte {
	offsetBytes := (int(jb.Header.RequestHeaderSectionOffsetWords) * 8) + JOB_BLOCK_DATA_OFFSET
	y := offsetBytes + int(jb.Header.RequestHeaderSectionSizeBytes)

	return jb.raw[offsetBytes:y:y]
}

func (jb *JobBlock) GetResponseHeaderBuffer() []byte {
	offsetBytes := (int(jb.Header.ResponseHeaderSectionOffsetWords) * 8) + JOB_BLOCK_DATA_OFFSET

	return jb.raw[offsetBytes : offsetBytes+int(jb.Header.ResponseHeaderSectionSizeBytes) : offsetBytes+int(jb.Header.ResponseHeaderSectionMaxSizeWords)*8]
}

func (jb *JobBlock) GetPayloadBufferForWrite() []byte {
	offsetBytes := (int(jb.Header.PayloadSectionOffsetWords) * 8) + JOB_BLOCK_DATA_OFFSET

	return jb.raw[offsetBytes : offsetBytes+int(jb.Header.PayloadSectionMaxSizeWords)*8 : offsetBytes+int(jb.Header.PayloadSectionMaxSizeWords)*8]
}

func (jb *JobBlock) GetPayloadBufferForRead() []byte {
	offsetBytes := (int(jb.Header.PayloadSectionOffsetWords) * 8) + JOB_BLOCK_DATA_OFFSET

	return jb.raw[offsetBytes : offsetBytes+int(jb.Header.PayloadSectionSizeBytes) : offsetBytes+int(jb.Header.PayloadSectionMaxSizeWords)*8]
}

type JobBlockAttacher struct {
	logger    logger.Logger
	jobBlocks *jobBlockPool
	shmPath   string
	shmData   *memory_accessor.MemoryAccessor
}

func NewJobBlockAttacher(parentLogger logger.Logger, shmPath string, numJobBlocks int) (*JobBlockAttacher, error) {

	newJobBlockAttacher := JobBlockAttacher{
		logger:    parentLogger.GetChild("job_block_attacher").(logger.Logger),
		shmPath:   shmPath,
		shmData:   nil,
		jobBlocks: newJobBlockPool(numJobBlocks),
	}

	for jobBlockIdx := 0; jobBlockIdx < numJobBlocks; jobBlockIdx++ {
		newJobBlockAttacher.jobBlocks.put(&JobBlock{})
	}

	// attach to shm
	mappedBytes, err := mapped_memory.NewMappedMemory(shmPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to map shm")
	}

	// map a memory accessor on the data of the mapped shm
	newJobBlockAttacher.shmData = memory_accessor.NewMemoryAccessor(newJobBlockAttacher.logger, mappedBytes.GetData())

	return &newJobBlockAttacher, nil
}

func (jba *JobBlockAttacher) AttachJob(offsetWords int) *JobBlock {
	jb := jba.jobBlocks.get()
	if jb == nil {
		panic("Out of job blocks")
	}

	jb.attach(jba.shmData, offsetWords*8)

	return jb
}

func (jba *JobBlockAttacher) ReleaseJob(jobBlock *JobBlock) {
	jba.jobBlocks.put(jobBlock)
}
