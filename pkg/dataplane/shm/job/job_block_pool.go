package job

type jobBlockPool struct {
	jobBlocks       []*JobBlock
	currentIndex    int
	maxNumJobBlocks int
}

func newJobBlockPool(maxNumJobBlocks int) *jobBlockPool {

	// create an empty jobBlock pool
	return &jobBlockPool{
		jobBlocks:       make([]*JobBlock, maxNumJobBlocks, maxNumJobBlocks),
		currentIndex:    0,
		maxNumJobBlocks: maxNumJobBlocks,
	}
}

func (rp *jobBlockPool) get() *JobBlock {
	if rp.currentIndex == 0 {
		return nil
	}

	// don't use defer here
	jobBlock := rp.jobBlocks[rp.currentIndex-1]
	rp.currentIndex--

	return jobBlock
}

func (rp *jobBlockPool) put(jobBlock *JobBlock) {
	if rp.currentIndex >= rp.maxNumJobBlocks {
		panic("JobBlock pool is full, can't receive another jobBlock")
	}

	rp.jobBlocks[rp.currentIndex] = jobBlock
	rp.currentIndex++
}

func (rp *jobBlockPool) len() int {
	return rp.currentIndex
}
