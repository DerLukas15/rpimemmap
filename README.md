# rpigpio [![GoDoc](https://godoc.org/github.com/DerLukas15/rpigpio?status.svg)](http://godoc.org/github.com/DerLukas15/rpigpio)
Package for controlling GPIOs on the Raspberry Pi using plain GO.

## About

This package provides access to the GPIO pins on any Raspberry Pi. The GPIOs are switch and controlled by directly writing to the corresponding registers.

The used Raspberry Pi hardware is automatically detected to accommodate different hardware addresses. For this to work, a curated list is included via the package [github.com/DerLukas15/rpihardware](https://github.com/DerLukas15/rpihardware).

## Installation

```sh
go get github.com/DerLukas15/rpigpio
```

## Usage

First initialize the package. This will determine the hardware. (You can also define your pins and initialize later. However the initialization has to be done before any methods are called.)
```go
err := rpigpio.Initialize()
```

Now create the pin object. Use the GPIO numbers which you can find online and not the physical pin number.
```go
pin, err := rpigpio.NewPin(17)
```
### Typical uses
#### Output
Set the 'direction' of the pin:
```go
err := pin.Mode(rpigpio.ModeOutput)
```

Set the output high:
```go
err := pin.Set(1)
```
Set the output low:
```go
err := pin.Set(0)
```

If the direction of the pin is an input, the `Set` method won't do anything.

#### Input
Set the 'direction' of the pin:
```go
err := pin.Mode(rpigpio.ModeInput)
```

Get the current value:
Now set the 'direction' of the pin:
```go
value, err := pin.Get()
```

If the direction of the pin is an output, the 'Get' method will always return 0.

### Extended uses
#### Alternate modes
To activate alternate modes for a pin use (e.g. for alternate mode 1)
```go
err := pin.Mode(rpigpio.ModeAlternate1)
```
#### Pull-Up / Pull-Down
To activate the build in Pull-Up or Pull-Down resistors use:

For Pull-Up
```go
err := pin.Pull(rpigpio.PullUp)
```

For Pull-Down
```go
err := pin.Pull(rpigpio.PullDown)
```

To disable Pull-Up or Pull-Down
```go
err := pin.Pull(rpigpio.PullOff)
```

The current status of the Pull-Up / Pull-Down resistors can not be read due to hardware restrictions.

#### Event Detects
There are different detect methods for events described below. To get the result of any event use:
```go
hasEvent, err := pin.Event()
```

Reading an event will clear the event to be ready for the next one. To prevent the automatic event clearing set
```go
rpigpio.NoEventClearing = true
```
Don't forget to clear the events by yourself
```go
err := pin.ClearEvent()
```

If an event is still valid during an attempt to clear it, the clear won't do anything.

##### Rising Edge Detect
Enable rising edge detection with
```go
err := pin.RisingEdgeDetect(true)
```
Any transition from 0 -> 1 will trigger the event.

Can be used together with 'Falling Edge Detect'

An asynchronous version of this method is available. This will sample outside of the system clock making it perfect for use with very short edges. Use `pin.ARisingEdgeDetect(true)`

##### Falling Edge Detect
Enable rising edge detection with
```go
err := pin.FallingEdgeDetect(true)
```
Any transition from 1 -> 0 will trigger the event.

Can be used together with 'Rising Edge Detect'

An asynchronous version of this method is available. This will sample outside of the system clock making it perfect for use with very short edges. Use `pin.FallingEdgeDetect(true)`

##### High Detect
Will set the event once the pin is high:
```go
err := pin.HighDetect(true)
```

##### Low Detect
Will set the event once the pin is low:
```go
err := pin.LowDetect(true)
```


## License

MIT, see [LICENSE](LICENSE)
