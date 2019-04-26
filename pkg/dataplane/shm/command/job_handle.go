package command

func GetHeapIdxFromJobHandle(jobHandle uint64) uint64 {
	return (jobHandle & 0x2F000000000) >> 36
}

func GetOffsetWordsFromJobHandle(jobHandle uint64) uint64 {
	return jobHandle & 0xFFFFFFFFF
}
