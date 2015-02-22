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
		pipes    []P
		lastPipe reflect.Value
		kill     reflect.Value
		closed   bool
		In       interface{}
		Out      interface{}
	}

	// Handler manages the piper struct
	Handler interface {
		Close() error
		Input() interface{}
		Output() interface{}
	}
)

// New creates a new piper instance
func New(in interface{}, out interface{}, pipefn ...P) (Handler, error) {
	piper, err := newPiper(in, out, pipefn)
	return Handler(piper), err
}

// Newpiper creates a new Piper from a channel and []*F
func newPiper(in interface{}, out interface{}, pipefn []P) (*Piper, error) {
	if len(pipefn) == 0 {
		return nil, ErrNoPiperFunc
	}

	if reflect.TypeOf(in).Kind() != reflect.Chan {
		return nil, ErrInvalidInputChannel
	}

	if reflect.TypeOf(out).Kind() != reflect.Chan {
		return nil, ErrInvalidOutputChannel
	}

	p := &Piper{
		pipes:  pipefn,
		In:     in,
		Out:    out,
		closed: false,
	}

	var pipe P
	for _, pipe = range p.pipes {
		err := p.checkPipe(&pipe)
		if err != nil {
			return nil, err
		}
	}
	p.lastPipe = reflect.ValueOf(pipe.Out)

	p.run()

	return p, nil
}

// checkPipe check's if the pipe is valid to be used in the piper
func (p *Piper) checkPipe(pipe PHandler) error {
	if pipe.Workers() < 1 {
		pipe.SetWorkers(1)
	}

	fn := pipe.Fn().Interface()
	out := pipe.Output().Interface()
	fnout := reflect.TypeOf(fn).Out(0).Kind()
	chanout := reflect.TypeOf(out).Elem().Kind()

	if fnout != chanout {
		return ErrInvalidPipeChannel
	}

	return nil
}

// Runs all steps and start monitoring the
func (p *Piper) run() (err error) {
	if len(p.pipes) == 0 {
		return ErrNoPiperFunc
	}

	var prev reflect.Value
	prev = reflect.ValueOf(p.In)
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
				output.Close()
				return
			}

		}
	}()

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
					reflect.ValueOf(p.Out).Send(recv)
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

	for i := 0; i < len(p.pipes); i++ {
		p.kill.Send(tru)
	}
	p.kill.Send(tru)

	p.kill.Close()
	p.clear()

	return nil
}

// Input returns the piper's input channel
func (p *Piper) Input() interface{} {
	return p.In
}

// Output returns the piper's output channel
func (p *Piper) Output() interface{} {
	return p.Out
}

type (
	// P is the main struct for the pipe
	P struct {
		WorkerNumber uint
		Out, Func    interface{}
	}

	// PHandler is the interface that manages the pipe
	PHandler interface {
		Output() reflect.Value
		Fn() reflect.Value
		Workers() uint
		SetWorkers(uint)
	}
)

// Output returns the reflect.Value of the pipe's output channel
func (p *P) Output() reflect.Value {
	return reflect.ValueOf(p.Out)
}

// Fn returns the reflect.Value of the pipe's function
func (p *P) Fn() reflect.Value {
	return reflect.ValueOf(p.Func)
}

// Workers returns the pipe's number of workers
func (p *P) Workers() uint {
	return p.WorkerNumber
}

// SetWorkers sets a new pipe's number of workers
func (p *P) SetWorkers(d uint) {
	dd := math.Min(float64(d), float64(MaxWorkers))
	p.WorkerNumber = uint(dd)
}
