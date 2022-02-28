package rpimemmap

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/DerLukas15/rpihardware"
	"github.com/dswarbrick/smart/ioctl"
	"github.com/pkg/errors"
)

//Memory flags for uncached memory
const (
	UncachedMemFlagDiscardable            = 1 << 0                                          // can be resized to 0 at any time. Use for cached data
	UncachedMemFlagNormal                 = 0 << 2                                          // normal allocating alias. Don't use from ARM
	UncachedMemFlagDirect          uint32 = 1 << 2                                          // 0xC alias uncached
	UncachedMemFlagCoherent               = 2 << 2                                          // 0x8 alias. Non-allocating in L2 but coherent
	UncachedMemFlagZero                   = 1 << 4                                          // initialise buffer to all zeros
	UncachedMemFlagNoInit                 = 1 << 5                                          // don't initialise (default is initialise to all ones)
	UncachedMemFlagHintPermaLock          = 1 << 6                                          // Likely to be locked for long periods of time
	UncachedMemFlagL1Nonallocation        = UncachedMemFlagDirect | UncachedMemFlagCoherent // Allocating in L2
)

//UncachedMap is the representation of uncached memory via videocore.
type UncachedMap struct {
	handle   int            // handle to videocore
	size     uint32         // size of allocated memory
	busAddr  uint32         // bus address
	virtAddr unsafe.Pointer // virtual address
	physAddr uint32         // physical address
	memRef   uint32         // memory reference from videocore
	mmapRef  []byte         // from MapSegment
}

type vcMsgStruct struct {
	len   uint32         // Overall length (bytes)
	req   uint32         // Zero for request, 1<<31 for response
	tag   uint32         // Command number
	blen  uint32         // Buffer length (bytes)
	dlen  uint32         // Data length (bytes)
	uints [32 - 5]uint32 // Data (108 bytes maximum)
}

//NewUncached creates a new UncachedMap by supplying the desired size.
func NewUncached(sizeInBytes uint32) *UncachedMap {
	return &UncachedMap{
		size: sizeInBytes,
	}
}

//String implements Stringer interface.
func (m *UncachedMap) String() string {
	var res string
	res += fmt.Sprintln("Uncached memory")
	res += fmt.Sprintf("Size %d bytes\n", m.size)
	res += fmt.Sprintf("PhysAddr %x\n", m.physAddr)
	res += fmt.Sprintf("BusAddr %x\n", m.busAddr)
	res += fmt.Sprintf("Virt Addr %x\n", m.virtAddr)
	return res
}

//Map maps a uncached memory to virtual memory.
func (m *UncachedMap) Map(unused uint32, unused2 string, uncachedMemFlags uint32) error {
	var err error
	if curHardware == nil {
		curHardware, err = rpihardware.Check()
		if err != nil {
			return err
		}
	}
	handle, err := vcOpen(m.size)
	if err != nil {
		return errors.Wrap(err, "uncached map open")
	}
	m.handle = handle
	err = m.allocate(uint32(os.Getpagesize()), uncachedMemFlags)
	if err != nil {
		return errors.Wrap(err, "uncached map allocate")
	}
	err = m.lock()
	if err != nil {
		return errors.Wrap(err, "uncached map allocate")
	}
	return nil
}

//Unmap unmaps a virtual address.
func (m *UncachedMap) Unmap() error {
	var err error
	if m.busAddr != 0 {
		err = m.unlock()
		if err != nil {
			return err
		}
	}
	if m.memRef != 0 {
		err = m.free()
		if err != nil {
			return err
		}
	}
	return nil
}

//used to take to videocore
func (m *UncachedMap) vcMsg(buf vcMsgStruct) (vcMsgStruct, error) {
	var retVal = vcMsgStruct{}
	if m.handle != 0 {
		var ui uint32
		for i := buf.dlen / 4; i <= buf.blen/4; i += 4 {
			buf.uints[i+1] = 0
		}
		buf.len = uint32((buf.blen + 6)) * uint32(unsafe.Sizeof(ui))
		buf.req = 0x0
		//fmt.Printf("Msg %v\n", buf)
		bPointer := uintptr(unsafe.Pointer(&buf))
		err := ioctl.Ioctl(uintptr(m.handle), ioctl.Iowr(100, 0, unsafe.Sizeof(bPointer)), bPointer)
		if err != nil {
			return vcMsgStruct{}, errors.Wrap(err, "ioctl uncachedMap msg")
		}
		res := *(*vcMsgStruct)(unsafe.Pointer(&buf))
		//fmt.Printf("Response %v\n", res)
		if res.req == 0x80000001 || (res.req&0x80000000) == 0 {
			//Error during request
			return vcMsgStruct{}, errors.New("Got error response from videocore")
		}
		return res, nil
	}
	return retVal, nil
}

//allocate uncached memory through videocore
func (m *UncachedMap) allocate(align, flags uint32) error {
	if m.memRef != 0 {
		return errors.New("Allread allocated")
	}
	var p = vcMsgStruct{
		tag:  0x3000c,
		blen: 12,
		dlen: 12,
	}
	p.uints[0] = m.size
	p.uints[1] = align
	p.uints[2] = flags
	var err error
	p, err = m.vcMsg(p)
	if err != nil {
		return errors.Wrap(err, "uncachedMap alloc")
	}
	m.memRef = p.uints[0]
	return nil
}

//free allocated uncached memory
func (m *UncachedMap) free() error {
	if m.memRef == 0 {
		return nil
	}
	var p = vcMsgStruct{
		tag:  0x3000f,
		blen: 4,
		dlen: 4,
	}
	p.uints[0] = m.memRef
	var err error
	p, err = m.vcMsg(p)
	if err != errors.Wrap(err, "uncachedMap alloc") {
		return err
	}
	if p.uints[0] != 0 {
		return errors.New("could not free mbox mem")
	}
	m.memRef = 0
	return nil
}

//lock allocated uncached memory and get adresses
func (m *UncachedMap) lock() error {
	if m.busAddr != 0 {
		return errors.New("Already locked")
	}
	var p = vcMsgStruct{
		tag:  0x3000d,
		blen: 4,
		dlen: 4,
	}
	p.uints[0] = m.memRef
	var err error
	p, err = m.vcMsg(p)
	if err != nil {
		return errors.Wrap(err, "uncachedMap lock")
	}
	m.busAddr = p.uints[0]
	m.physAddr = busToPhys(m.busAddr)
	virtAddrs, err := MapSegment(m, MemDevDefault)
	if err != nil {
		return errors.Wrap(err, "uncachedMap lock")
	}
	m.virtAddr = unsafe.Pointer(&virtAddrs[0])
	m.mmapRef = virtAddrs
	return nil
}

//unlock and release uncached memory
func (m *UncachedMap) unlock() error {
	if m.busAddr == 0 {
		return nil
	}
	var p = vcMsgStruct{
		tag:  0x3000e,
		blen: 4,
		dlen: 4,
	}
	p.uints[0] = m.memRef
	var err error
	p, err = m.vcMsg(p)
	if err != nil {
		return errors.Wrap(err, "uncachedMap unlock")
	}
	if p.uints[0] != 0 {
		return errors.New("could not free mbox mem")
	}
	err = UnmapSegment(m.mmapRef)
	if err != nil {
		return errors.Wrap(err, "uncachedMap unlock")
	}
	m.busAddr = 0
	m.physAddr = 0
	m.virtAddr = nil
	m.mmapRef = nil
	return nil
}

//BusAddr returns the current bus address if available.
func (m *UncachedMap) BusAddr() uint32 {
	return m.busAddr
}

//PhysAddr returns the current physical address if available.
func (m *UncachedMap) PhysAddr() uint32 {
	return m.physAddr
}

//VirtAddr returns the current virtual address if available.
func (m *UncachedMap) VirtAddr() unsafe.Pointer {
	return m.virtAddr
}

//Size returns the current size.
func (m *UncachedMap) Size() uint32 {
	return m.size
}
