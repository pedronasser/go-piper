package piper

import (
	"errors"
	"math"
	"reflect"
	"sync"
)

// Constants
const (
	MaxWorkers = 1000
)

// Piper Errors
var (
	ErrNoPiperFunc          = errors.New("[]Piper.pipes is empty")
	ErrInvalidInput         = errors.New("Input data type invalid")
	ErrInvalidOutput        = errors.New("Output should be a pointer")
	ErrInvalidOutputType    = errors.New("Wrong type for output")
	ErrPiperClosed          = errors.New("Piper is already closed")
	ErrInvalidOutputChannel = errors.New("Output interface should be a channel")
	ErrInvalidInputChannel  = errors.New("Input interface should be a channel")
	ErrOutOfRange           = errors.New("Out of range of pipes")
	ErrInvalidPipeChannel   = errors.New("Pipe's channel must have the same type as the function's return")
)

type (
	// Piper is the main struct on the piper
	Piper struct {
		mu       sync.Mutex
		pipes    []Pipe
		lastPipe reflect.Value
		kill     reflect.Value
		closed   bool
		in       interface{}
		out      interface{}
	}

	// Handler manages the piper struct
	Handler interface {
		Close() error
		Input() interface{}
		Output() interface{}
	}
)

// New creates a new piper instance
func New(pipefn ...Pipe) (Handler, error) {
	piper, err := newPiper(pipefn)
	return Handler(piper), err
}

// Newpiper creates a new Piper from a channel and []*F
func newPiper(pipefn []Pipe) (p *Piper, err error) {
	if len(pipefn) == 0 {
		return nil, ErrNoPiperFunc
	}

	p = &Piper{
		pipes:  pipefn,
		closed: false,
	}

	var pipe Pipe
	for _, pipe = range p.pipes {
	    p.checkPipe(&pipe);
	}
	p.lastPipe = reflect.ValueOf(pipe.out)

	p.createInput();
    p.createOutput();

	p.run()

	return
}

// createInput cretes a new input channel based on the type of the first step's argument
func (p *Piper) createInput() (err error) {
	target := p.pipes[0]
	p.in = reflect.Indirect(
		reflect.MakeChan(
			reflect.ChanOf(reflect.BothDir, reflect.TypeOf(target.Fn().Interface()).In(0)),
			0,
		),
	).Interface()

	return
}

// createOutput cretes a new input channel based on the type of the last step's return
func (p *Piper) createOutput() (err error) {
	target := p.pipes[len(p.pipes)-1]
	p.out = reflect.Indirect(
		reflect.MakeChan(
			reflect.ChanOf(reflect.BothDir, reflect.TypeOf(target.Fn().Interface()).Out(0)),
			0,
		),
	).Interface()

	return
}

// checkPipe check's if the pipe is valid to be used in the piper
func (p *Piper) checkPipe(pipe PHandler) error {
	if pipe.Workers() < 1 {
		pipe.SetWorkers(1)
	}

	return nil
}

// Runs all steps and start monitoring the
func (p *Piper) run() (err error) {
	if len(p.pipes) == 0 {
		return ErrNoPiperFunc
	}

	var prev reflect.Value
	prev = reflect.ValueOf(p.in)
	p.kill = reflect.ValueOf(make(chan bool, len(p.pipes)+1))

	for _, pipe := range p.pipes {
		prev = p.runPipe(&pipe, prev)
	}

	go p.outputMonitor()

	return nil
}

// GetOutput gets the reflect.Value of the pipe's output channel
func (p *Piper) GetOutput(pipe PHandler) reflect.Value {
	return pipe.Output()
}

// runPipe creates and run the target step's goroutine.
func (p *Piper) runPipe(pipe PHandler, prev reflect.Value) (output reflect.Value) {
	output = pipe.Output()
	fn := pipe.Fn()

	var w uint
	for w = 0; w < pipe.Workers(); w++ {
		go func() {
			selects := []reflect.SelectCase{
				reflect.SelectCase{Dir: reflect.SelectRecv, Chan: prev},
				reflect.SelectCase{Dir: reflect.SelectRecv, Chan: p.kill},
			}

			for {
				chosen, recv, ok := reflect.Select(selects)
				switch chosen {
				case 0:
					if ok {
						v := fn.Call([]reflect.Value{recv})
						output.Send(v[0])
					}
				case 1:
					if pipe.Closed() == false {
						output.Close()
						pipe.Close()
					}
					return
				}
			}

		}()
	}

	return
}

// Goroutine the monitors the last step's channel
// waiting for the final response.
func (p *Piper) outputMonitor() {
	for {
		selects := []reflect.SelectCase{
			reflect.SelectCase{Dir: reflect.SelectRecv, Chan: p.lastPipe},
			reflect.SelectCase{Dir: reflect.SelectRecv, Chan: p.kill},
		}

		for {
			chosen, recv, ok := reflect.Select(selects)
			switch chosen {
			case 0:
				if ok {
					reflect.ValueOf(p.out).Send(recv)
				}
			case 1:
				return
			}

		}
	}
}

// This methods clear all information on the *Piper struct
func (p *Piper) clear() {
	p.pipes = nil
}

// Close sends a kill sign to all go routines
// Then it calls p.clear()
func (p *Piper) Close() error {
	if p.closed == true {
		return ErrPiperClosed
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true

	tru := reflect.ValueOf(p.closed)

	var w uint
	for i := 0; i < len(p.pipes); i++ {
		for w = 0; w < p.pipes[i].workerNumber; w++ {
			p.kill.Send(tru)
		}
	}
	p.kill.Send(tru)

	p.kill.Close()
	p.clear()

	return nil
}

// Input returns the piper's input channel
func (p *Piper) Input() interface{} {
	return p.in
}

// Output returns the piper's output channel
func (p *Piper) Output() interface{} {
	return p.out
}

// P creates a new P (pipe)
func P(w uint, fn interface{}) (pipe Pipe) {
	pipe = Pipe{}
	pipe.workerNumber = w
	pipe.pfunc = fn

	pipe.createOutput()

	return
}

type (
	// Pipe is the main struct for the pipe
	Pipe struct {
		workerNumber uint
		out, pfunc   interface{}
		closed       bool
	}

	// PHandler is the interface that manages the pipe
	PHandler interface {
		Output() reflect.Value
		Fn() reflect.Value
		Workers() uint
		SetWorkers(uint)
		Close()
		Closed() bool
	}
)

// createOutput creates a new channel for output for this pipe
func (p *Pipe) createOutput() (err error) {
	p.out = reflect.Indirect(reflect.MakeChan(
		reflect.ChanOf(reflect.BothDir, reflect.TypeOf(p.Fn().Interface()).Out(0)),
		0,
	)).Interface()

	return
}

// Output returns the reflect.Value of the pipe's output channel
func (p *Pipe) Output() reflect.Value {
	return reflect.ValueOf(p.out)
}

// Fn returns the reflect.Value of the pipe's function
func (p *Pipe) Fn() reflect.Value {
	return reflect.ValueOf(p.pfunc)
}

// Workers returns the pipe's number of workers
func (p *Pipe) Workers() uint {
	return p.workerNumber
}

// SetWorkers sets a new pipe's number of workers
func (p *Pipe) SetWorkers(d uint) {
	dd := math.Min(float64(d), float64(MaxWorkers))
	p.workerNumber = uint(dd)
}

// Close sets pipe to closed
func (p *Pipe) Close() {
	p.closed = true
}

// Closed checks if pipe is closed
func (p *Pipe) Closed() bool {
	return p.closed
}
