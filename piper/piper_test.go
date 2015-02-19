package piper

import (
    "testing"
)

// Build a default int-typed Piper
func BuildTestPipe(t *testing.T) (piper PipeHandler) {
    piper, err := NewPiper(
        make(chan int),
        &F{
            make(chan int),
            func (d int) int { return d+1 },
        },
    )

    if err != nil {
        t.Fatalf(err.Error())
    }

    return
}

func TestNewPipe(t *testing.T) {
    BuildTestPipe(t);
}

func TestNewPipeInputOutput(t *testing.T) {
    piper := BuildTestPipe(t);

    var result int
    var expect int = 2
    piper.In(1, &result)

    if (result != expect) {
        t.Errorf("Expected `result` to be %d (got %d)", expect, result)
    }

    err := piper.In("", &result)
    if err != ErrInvalidInput {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrInvalidInput, err)
    }

    err = piper.In(1, result)
    if err != ErrInvalidOutput {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrInvalidOutput, err)
    }

    piper = &Piper{closed: false}
    err = piper.In(1, result)
    if err != ErrNoPiperFunc {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrNoPiperFunc, err)
    }
}

func TestNewPipeRun(t *testing.T) {
    piper := &Piper{closed: false}
    err := piper.run()
    if err != ErrNoPiperFunc {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrNoPiperFunc, err)
    }
}

func TestClosePipe(t *testing.T) {
    piper := BuildTestPipe(t);
    piper.Close()

    var result int
    err := piper.In(1, &result)

    if (err != ErrPiperClosed) {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrPiperClosed, err)
    }

    err = piper.Close()
    if (err != ErrPiperClosed) {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrPiperClosed, err)
    }
}

func TestNewPipeNoFuncs(t *testing.T) {
    _, err := newPiper(
        make(chan int),
        []*F{},
    )

    if err != ErrNoPiperFunc {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrNoPiperFunc, err)
    }
}

func TestNewPipeInvalidOutput(t *testing.T) {
    _, err := newPiper(
        1,
        []*F{
            &F{
                make(chan int),
                func (d int) int { return d+1 },
            },
        },
    )

    if err != ErrInvalidOutputChannel {
        t.Errorf("Expected `err` to be '%v' (got '%v')", ErrInvalidOutputChannel, err)
    }
}

// Benchmarks

func BenchmarkInputOutput(b *testing.B) {
    piper,_ := NewPiper(
        make(chan int),
        &F{
            make(chan int),
            func (d int) int { return d },
        },
    )

    var result int = 0
    for i := 0; i < b.N; i++ {
        piper.In(i, result)
	}
}

func BenchmarkThousandsPipes(b *testing.B) {
    var pipes []*F = make([]*F, 10000)

    for i:=0; i<10000; i++ {
        pipes[i] = &F{
            make(chan int),
            func (d int) int { return d },
        }
    }

    piper,_ := newPiper(
        make(chan int),
        pipes,
    )

    var result int = 0
    for i := 0; i < b.N; i++ {
        piper.In(i, result)
	}
}