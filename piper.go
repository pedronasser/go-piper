package piper

import (
	"errors"
	"math"
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
		mu     sync.Mutex
		pipes  []*Pipe
		closed bool
		in     chan interface{}
		out    chan interface{}
	}

	// Handler manages the piper struct
	Handler interface {
		Close() error
		Input() chan interface{}
		Output() chan interface{}
	}
)

// New creates a new piper instance
func New(pipefn ...*Pipe) (Handler, error) {
	piper, err := newPiper(pipefn)
	return Handler(piper), err
}

// Newpiper creates a new Piper from a channel and []*F
func newPiper(pipefn []*Pipe) (p *Piper, err error) {
	if len(pipefn) == 0 {
		return nil, ErrNoPiperFunc
	}

	p = &Piper{
		pipes:  pipefn,
		closed: false,
	}

	for i := range pipefn {
		p.checkPipe(p.pipes[i])
	}

	p.out = p.pipes[len(p.pipes)-1].out

	p.createInput()
	p.run()

	return
}

// createInput cretes a new input channel based on the type of the first step's argument
func (p *Piper) createInput() (err error) {
	p.in = make(chan interface{})

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

	prev := p.in

	for i := range p.pipes {
		prev = p.runPipe(i, prev)
	}

	return nil
}

// GetOutput gets the pipe's output channel
func (p *Piper) GetOutput(pipe PHandler) interface{} {
	return pipe.Output()
}

// runPipe creates and run the target step's goroutine.
func (p *Piper) runPipe(index int, last chan interface{}) (output chan interface{}) {
	pipe := p.pipes[index]

	output = pipe.Output()

	var w uint
	pipe.kill = make(chan bool, pipe.Workers())
	for w = 0; w < pipe.Workers(); w++ {
		go func() {
			fn := p.pipes[index].Fn()
			for {
				select {
				case v := <-last:
					r := fn(v)
					output <- r
				case <-pipe.kill:
					close(output)
					return
				}
			}
		}()
	}

	return
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

	for _, pipe := range p.pipes {
		pipe.KillChannel()
	}

	return nil
}

// Input returns the piper's input channel
func (p *Piper) Input() chan interface{} {
	return p.in
}

// Output returns the piper's output channel
func (p *Piper) Output() chan interface{} {
	return p.out
}

// P creates a new P (pipe)
func P(w uint, fn func(interface{}) interface{}) (pipe *Pipe) {
	pipe = &Pipe{}
	pipe.workerNumber = w
	pipe.pfunc = fn
	pipe.createOutput()

	return
}

type (
	// Pipe is the main struct for the pipe
	Pipe struct {
		workerNumber uint
		out          chan interface{}
		pfunc        func(interface{}) interface{}
		closed       bool
		kill         chan bool
	}

	// PHandler is the interface that manages the pipe
	PHandler interface {
		Output() chan interface{}
		Fn() func(interface{}) interface{}
		Workers() uint
		SetWorkers(uint)
		KillChannel()
		Close()
		Closed() bool
	}
)

// createOutput creates a new channel for output for this pipe
func (p *Pipe) createOutput() (err error) {
	p.out = make(chan interface{})

	return
}

// Output returns the pipe's output channel
func (p *Pipe) Output() chan interface{} {
	return p.out
}

// Fn returns the pipe's function
func (p *Pipe) Fn() func(interface{}) interface{} {
	return p.pfunc
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

// KillChannel kill all working goroutines
func (p *Pipe) KillChannel() {
	workers := int(p.Workers())
	for w := 0; w < workers; w++ {
		p.kill <- true
	}
}

// Close sets pipe to closed
func (p *Pipe) Close() {
	p.closed = true
}

// Closed checks if pipe is closed
func (p *Pipe) Closed() bool {
	return p.closed
}
