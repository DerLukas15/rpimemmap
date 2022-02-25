package rpimemmap

import (
	"fmt"
	"unsafe"

	"github.com/pkg/errors"
)

//PeripheralMap is the representation of memory to a peripheral.
type PeripheralMap struct {
	size     uint32         // size of allocated memory
	busAddr  uint32         // bus address
	virtAddr unsafe.Pointer // virtual address
	physAddr uint32         // physical address
}

//NewPeripheral creates a new PeripheralMap by supplying the desired size. The size is rounded up to the nearest pageSize.
func NewPeripheral(sizeInBytes uint32) *PeripheralMap {
	return &PeripheralMap{
		size: pageRoundUp(sizeInBytes),
	}
}

//String implements Stringer interface
func (m *PeripheralMap) String() string {
	var res string
	res += fmt.Sprintln("Mapping to peripheral")
	res += fmt.Sprintf("Size %d bytes", m.size)
	res += fmt.Sprintf("PhysAddr %x\n", m.physAddr)
	res += fmt.Sprintf("BusAddr %x\n", m.busAddr)
	res += fmt.Sprintf("Virt Addr %x\n", m.virtAddr)
	return res
}

//Map maps a peripheral to virtual memory by physical address.
//memDev can be MemDevDefault (requires root) or MemDevGPIO (only for GPIOs)
func (m *PeripheralMap) Map(physAddr uint32, memDev string, unused uint32) error {
	var err error
	m.physAddr = physAddr
	m.busAddr = physAddr + busRegisterBase
	m.virtAddr, err = MapSegment(m, memDev)
	if err != nil {
		return err
	}
	return nil
}

//Unmap unmaps a virtual address
func (m *PeripheralMap) Unmap() error {
	if m.virtAddr == nil {
		//Not mapped
		return nil
	}
	err := UnmapSegment(m)
	if err != nil {
		return errors.Wrap(err, "unmap peripheral")
	}
	m.virtAddr = nil
	return nil
}

//BusAddr returns the current bus address if available
func (m *PeripheralMap) BusAddr() uint32 {
	return m.busAddr
}

//PhysAddr returns the current physical address if available
func (m *PeripheralMap) PhysAddr() uint32 {
	return m.physAddr
}

//VirtAddr returns the current virtual address if available
func (m *PeripheralMap) VirtAddr() unsafe.Pointer {
	return m.virtAddr
}

//Size returns the current size
func (m *PeripheralMap) Size() uint32 {
	return m.size
}
