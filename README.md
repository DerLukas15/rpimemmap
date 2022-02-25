# rpimemmap [![GoDoc](https://godoc.org/github.com/DerLukas15/rpimemmap?status.svg)](http://godoc.org/github.com/DerLukas15/rpimemmap)
Package for allocating uncached memory and physical devices on Raspberry Pi's using plain GO.

## About

This package provides access to uncached memory and physical devices on any Raspberry Pi. Uncached memory is done with the help of videocore. All access is directly written to memory.

## Installation

```sh
go get github.com/DerLukas15/rpimemmap
```

## Usage

## Open
You need to supply the desired size during the creation of the actual map. However, the actual allocation is done during the map call.

The supplyed size is rounded up to the nearest pageSize.

### Physical Devices
Map any physical device to virtual memory for access by supplying the physical address of the device:
```go
deviceMap := rpimemmap.NewPeripheral(size)
err := deviceMap.Map(physicalAddress, memoryDevice, 0)
```

### Uncached Memory
Map uncached memory to virtual memory for access:
```go
uncachedMap := rpimemmap.NewUncached(size)
err := uncachedMap.Map(0, "", UncachedMemFlags)
```

## Close
Close any MemMap with
```go
err := curMap.Unmap()
```

## License

MIT, see [LICENSE](LICENSE)
