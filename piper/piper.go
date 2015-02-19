package piper

import (
    "errors"
    "reflect"
    "sync"
)

// Piper Errors
var (
	ErrNoPiperFunc          = errors.New("[]Piper.pipes is empty")
	ErrInvalidInput         = errors.New("Input data type invalid")
	ErrInvalidOutput        = errors.New("Output should be a pointer")
	ErrInvalidOutputType    = errors.New("Wrong type for output")
	ErrPiperClosed          = errors.New("Piper is already closed")
	ErrInvalidOutputChannel = errors.New("Output interface should be a channel")
)

// Piper Types
type (
    Piper struct {
        PipeHandler
        mu          sync.Mutex
        pipes       []*F
        kill        chan bool
        closed      bool
        out         interface{}
    }

    F struct {
        In, Fn      interface{}
    }

    PipeHandler interface {
        In(in interface{}, out interface{}) error
        Close() error
    }
)

// Creates a new Piper from a channel and ...*Fs
func NewPiper(out interface{}, pipefn ...*F) (PipeHandler, error) {
    return newPiper(out, pipefn)
}

// Creates a new Piper from a channel and []*Fs
func newPiper(out interface{}, pipefn []*F) (PipeHandler, error) {
    if (len(pipefn) == 0) {
        return nil, ErrNoPiperFunc
    }

    if (reflect.TypeOf(out).Kind() != reflect.Chan) {
        return nil, ErrInvalidOutputChannel
    }

    pipe := &Piper{
        pipes:  pipefn,
        out:    out,
        closed: false,
    }

    pipe.run();

    return PipeHandler(pipe), nil
}

// This method should create a go routine for each Pipefunc
// Then it creates a kill channel for each go routine
// The go routine will wait for data or a kill sign
func (p *Piper) run() error {
    if len(p.pipes) == 0 {
        return ErrNoPiperFunc
    }

	var kill chan bool
	p.kill = make(chan bool)
	kill = p.kill

	for i:=0; i<len(p.pipes); i++ {
	    next := make(chan bool)

    	go func(i int, die chan bool, next chan bool) {
    	    fn := reflect.ValueOf(p.pipes[i].Fn)
    	    in := reflect.ValueOf(p.pipes[i].In)
    	    var out reflect.Value
            if (i+1 < len(p.pipes)) {
                out = reflect.ValueOf(p.pipes[i+1].In)
            } else {
                out = reflect.ValueOf(p.out)
            }

            dieValue := reflect.ValueOf(die)
            selects := []reflect.SelectCase{
                reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: in },
                reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: dieValue },
            }

    		for {
                chosen, recv, ok := reflect.Select(selects)
                switch chosen {
                    case 0:
            			if (ok) {
            			    out.Send(fn.Call([]reflect.Value{recv})[0])
            			}
                    case 1:
    				    next <- true
    					in.Close()
    					close(die)
    					return
                }

    		}
    	}(i, kill, next)

		kill = next
	}

    return nil
}

// Sends an interface{} to the pipeline
// Then waits for the result
// The result will be addressed to the out interface{} pointer
func (p *Piper) In(in interface{}, out interface{}) error {
    if (p.closed == true) {
        return ErrPiperClosed
    }

    if len(p.pipes) == 0 {
        return ErrNoPiperFunc
    }

    if reflect.ChanOf(reflect.BothDir,reflect.TypeOf(in)) != reflect.TypeOf(p.pipes[0].In) {
        return ErrInvalidInput
    }

    if reflect.TypeOf(out).Kind() != reflect.Ptr {
        return ErrInvalidOutput
    }

    p.mu.Lock()
	defer p.mu.Unlock()

    reflect.ValueOf(p.pipes[0].In).Send(reflect.ValueOf(1))

    if v, ok := reflect.ValueOf(p.out).Recv(); ok {
        reflect.ValueOf(out).Elem().Set(v)
    }

    return nil
}

// This methods clear all information on the *Piper struct
func (p *Piper) clear() {
    p.mu.Lock()
	defer p.mu.Unlock()

    p.closed = true
    p.pipes = nil
}

// Sends a kill sign to all go routines
// Then it calls p.clear()
func (p *Piper) Close() error {
    if (p.closed == true) {
        return ErrPiperClosed
    }

    p.kill <- true
    p.clear()

    return nil
}