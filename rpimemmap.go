//Package rpimemmap is used for allocating uncached memory and physical devices on Raspberry Pi's using plain GO.
package rpimemmap

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/DerLukas15/rpihardware"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

//Valid MemDev to use for mapping of peripherals.
const (
	MemDevDefault = "/dev/mem"
	MemDevGPIO    = "/dev/gpiomem"
)

const (
	busRegisterBase = 0x7e000000
)

var curHardware *rpihardware.Hardware // Set during initialize

//MemMap defines a memory device.
type MemMap interface {
	fmt.Stringer
	Map(physAddr uint32, memDev string, flags uint32) error
	Unmap() error
	PhysAddr() uint32
	BusAddr() uint32
	VirtAddr() unsafe.Pointer
	Size() uint32
}

//MapSegment maps a MemMap to virtual memory from physical memory address.
//See const MemDev for possible memDev.
func MapSegment(curMap MemMap, memDev string) ([]byte, error) {
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
	return mem, nil
}

//UnmapSegment unmaps a MemMap from virtual memory
func UnmapSegment(ref []byte) error {
	err := unix.Munmap(ref)
	if err != nil {
		return errors.Wrap(err, "UnmapSegment")
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

//Reg32 returns a pointer to 4 bytes (32 bits as uint32) of memory from m which starts with an offset by memOffsetToBase bytes.
//You should be aware that the order of the bits does not have to be the actual order in memory as uint32 does not have to store the bits in the correct order (endian). Use multiple Reg8 calls to get the same memory region in the correct bit order.
/*
Example:

The virtual memory address of the MemMap 'm' is 0xa6c1e000.

`pointerToMemory := Reg32(m, 0)` will be 32 bits which start at 0xa6c1e000

`pointerToMemory := Reg32(m, 1)` will be 32 bits which start at 0xa6c1e001 as 1 byte was added

`pointerToMemory := Reg32(m, 8)` will be 32 bits which start at 0xa6c1e008 as 8 bytes were added
*/
func Reg32(m MemMap, memOffsetToBase uint32) *uint32 {
	if m.VirtAddr() == nil {
		return nil
	}
	return (*uint32)(unsafe.Add(m.VirtAddr(), memOffsetToBase))
}

//Reg8 returns a pointer to 1 byte of memory from m which starts with an offset by memOffsetToBase bytes.
//The order of bits is unaltered. See Reg32 for detailed description of usage.
func Reg8(m MemMap, memOffsetToBase uint32) *byte {
	if m.VirtAddr() == nil {
		return nil
	}
	return (*byte)(unsafe.Add(m.VirtAddr(), memOffsetToBase))
}

//Dump will output the content of the memory. If size == 0, the size of the allocated memory will be used.
func Dump(m MemMap, size uint32) string {
	var res string
	entryPerLine := uint32(8)
	if size == 0 {
		size = m.Size()
	}
	res += fmt.Sprintf("memdump %d bytes starting at 0x%X\n", size, m.VirtAddr())
	for i := uint32(1); i <= size; i++ {
		if (i-1)%entryPerLine == 0 {
			res += fmt.Sprintf("0x%X: ", unsafe.Add(m.VirtAddr(), (i-1)))
		}
		res += fmt.Sprintf("%08b ", *(*byte)(unsafe.Add(m.VirtAddr(), (i - 1))))
		if i%entryPerLine == 0 {
			res += fmt.Sprintf("(%d)\n", i)
		}
	}
	res += fmt.Sprintf("\n")
	return res
}
