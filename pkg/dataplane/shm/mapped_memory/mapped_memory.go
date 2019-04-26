package mapped_memory

import (
	"fmt"
	"os"
	"syscall"

	"github.com/fabiokung/shm"
)

type MappedMemory struct {
	data *[]byte
	path string
}

func NewMappedMemory(path string) (*MappedMemory, error) {
	newMappedMemory := MappedMemory{
		path: path,
	}

	if err := newMappedMemory.init(); err != nil {
		return nil, err
	}

	return &newMappedMemory, nil
}

func (mm *MappedMemory) Close() {
	if mm.data != nil {
		syscall.Munmap(*mm.data)
		mm.data = nil
	}
}

func (mm *MappedMemory) GetData() []byte {
	return *mm.data
}

func (mm *MappedMemory) init() error {

	// open the file and mmap it
	file, err := shm.Open(mm.path, os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		panic(err)
	}

	defer file.Close()
	fileInfo, err := file.Stat()

	if err != nil {
		return err
	}

	size := fileInfo.Size()
	if size <= 0 {
		return fmt.Errorf("mmap: file %q has negative size", mm.path)
	}

	if size != int64(int(size)) {
		return fmt.Errorf("mmap: file %q is too large", mm.path)
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}

	// unmap when done
	// runtime.SetFinalizer(mm, (*MappedMemory).Close)

	// initialize members now that all is well
	mm.data = &data

	return nil
}
