package memory_accessor

import (
	"unsafe"

	"github.com/nuclio/logger"
)

type MemoryAccessor struct {
	logger logger.Logger
	data   []byte
}

func NewMemoryAccessor(parentLogger logger.Logger, data []byte) *MemoryAccessor {
	newMemoryAccessor := MemoryAccessor{
		logger: parentLogger.GetChild("MemoryAccessor").(logger.Logger),
		data:   data,
	}

	newMemoryAccessor.logger.DebugWith("Created", "dataLen", len(data))

	return &newMemoryAccessor
}

func (ma *MemoryAccessor) ReadInt8(offset int) int8 {
	return *ma.AtInt8(offset)
}

func (ma *MemoryAccessor) WriteInt8(offset int, value int8) {
	*ma.AtInt8(offset) = value
}

func (ma *MemoryAccessor) AtInt8(offset int) *int8 {
	return (*int8)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadInt16(offset int) int16 {
	return *ma.AtInt16(offset)
}

func (ma *MemoryAccessor) WriteInt16(offset int, value int16) {
	*ma.AtInt16(offset) = value
}

func (ma *MemoryAccessor) AtInt16(offset int) *int16 {
	return (*int16)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadInt32(offset int) int32 {
	return *ma.AtInt32(offset)
}

func (ma *MemoryAccessor) WriteInt32(offset int, value int32) {
	*ma.AtInt32(offset) = value
}

func (ma *MemoryAccessor) AtInt32(offset int) *int32 {
	return (*int32)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadInt64(offset int) int64 {
	return *ma.AtInt64(offset)
}

func (ma *MemoryAccessor) WriteInt64(offset int, value int64) {
	*ma.AtInt64(offset) = value
}

func (ma *MemoryAccessor) AtInt64(offset int) *int64 {
	return (*int64)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadUint8(offset int) uint8 {
	return *ma.AtUint8(offset)
}

func (ma *MemoryAccessor) WriteUint8(offset int, value uint8) {
	*ma.AtUint8(offset) = value
}

func (ma *MemoryAccessor) AtUint8(offset int) *uint8 {
	return (*uint8)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadUint16(offset int) uint16 {
	return *ma.AtUint16(offset)
}

func (ma *MemoryAccessor) WriteUint16(offset int, value uint16) {
	*ma.AtUint16(offset) = value
}

func (ma *MemoryAccessor) AtUint16(offset int) *uint16 {
	return (*uint16)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadUint32(offset int) uint32 {
	return *ma.AtUint32(offset)
}

func (ma *MemoryAccessor) WriteUint32(offset int, value uint32) {
	*ma.AtUint32(offset) = value
}

func (ma *MemoryAccessor) AtUint32(offset int) *uint32 {
	return (*uint32)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) ReadUint64(offset int) uint64 {
	return *ma.AtUint64(offset)
}

func (ma *MemoryAccessor) WriteUint64(offset int, value uint64) {
	*ma.AtUint64(offset) = value
}

func (ma *MemoryAccessor) AtUint64(offset int) *uint64 {
	return (*uint64)(unsafe.Pointer(&ma.data[offset]))
}

func (ma *MemoryAccessor) AtBytes(offset int, len int) []byte {
	return ma.data[offset : offset+len]
}
