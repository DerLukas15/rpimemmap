package rpimemmap

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

//Valid memDev to use for mapping of peripherals
const (
	MemDevDefault = "/dev/mem"
	MemDevGPIO    = "/dev/gpiomem"
)

const (
	busRegisterBase = 0x7e000000
)

//MemMap defines a memory device
type MemMap interface {
	fmt.Stringer
	Map(physAddr uint32, memDev string, flags uint32) error
	Unmap() error
	PhysAddr() uint32
	BusAddr() uint32
	VirtAddr() unsafe.Pointer
	Size() uint32
}

// get virtual memory segment for from physical memory address
func mapSegment(curMap MemMap, memDev string) (unsafe.Pointer, error) {
	memFd, err := os.OpenFile(memDev, os.O_RDWR, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "open memdevice")
	}
	flags := unix.MAP_SHARED
	prot := unix.PROT_READ | unix.PROT_WRITE
	mem, err := unix.Mmap(int(memFd.Fd()), int64(curMap.PhysAddr()), int(curMap.Size()), prot, flags)
	if err != nil {
		return nil, errors.Wrap(err, "mmap call")
	}
	memFd.Close()
	return unsafe.Pointer(&mem[0]), nil
}

//unmap virtual memory
func unmapSegment(mem unsafe.Pointer) error {
	err := unix.Munmap(*(*[]byte)(mem))
	if err != nil {
		return err
	}
	return nil
}

func pageRoundUp(size uint32) uint32 {
	pageSize := uint32(os.Getpagesize())
	if size%pageSize == 0 {
		return size
	}
	return (size + pageSize) % (^(pageSize - 1))
}

// convert bus address to phys address
func busToPhys(x uint32) uint32 {
	return x & ^uint32(0xc0000000)
}

//Reg32 returns a pointer to 32 bytes (as uint32) of memory from m which starts with an offset by memOffsetToBase.
func Reg32(m MemMap, memOffsetToBase uint32) *uint32 {
	if m.VirtAddr() == nil {
		return nil
	}
	return (*uint32)(unsafe.Add(m.VirtAddr(), memOffsetToBase))
}

//Reg8 returns a pointer to 1 byte of memory from m which starts with an offset by memOffsetToBase.
func Reg8(m MemMap, memOffsetToBase uint32) *byte {
	if m.VirtAddr() == nil {
		return nil
	}
	return (*byte)(unsafe.Add(m.VirtAddr(), memOffsetToBase))
}
