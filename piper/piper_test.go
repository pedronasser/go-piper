package piper

import (
	"testing"
	"sync"
)

// Build a default int-typed Piper
func BuildTestPipe(t *testing.T) (piper Handler) {
	piper, err := New(
		P(
			1,
			func(d int) int { return d + 1 },
		),
	)

	if err != nil {
		t.Fatalf(err.Error())
	}

	return
}

func TestNewPipe(t *testing.T) {
	BuildTestPipe(t)
}

func TestNewPipeInputOutput(t *testing.T) {
	piper := BuildTestPipe(t)

	input, ok := piper.Input().(chan int)

	if !ok {
		t.Errorf("Expected `input` to be `chan int` (got %+v)", input)
	}

	output, ok2 := piper.Output().(chan int)

	if !ok2 {
		t.Errorf("Expected `output` to be `chan int` (got %+v)", output)
	}

	var result int
	expect := 2

	input <- 1
	result = <-output

	if result != expect {
		t.Errorf("Expected `result` to be `%d` (got %d)", expect, result)
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
	piper := BuildTestPipe(t)
	piper.Close()

	err := piper.Close()
	if err != ErrPiperClosed {
		t.Errorf("Expected `err` to be '%v' (got '%v')", ErrPiperClosed, err)
	}
}

func TestPipeCheck(t *testing.T) {
	_, err := newPiper(
		[]Pipe{},
	)

	if err != ErrNoPiperFunc {
		t.Errorf("Expected `err` to be '%v' (got '%v')", ErrNoPiperFunc, err)
	}

	piper, _ := newPiper(
		[]Pipe{
			P(
				0,
				func(d int) int { return d },
			),
		},
	)

	var expected uint = 1000
	piper.pipes[0].SetWorkers(1001)
	w := piper.pipes[0].Workers()
	if w != expected {
		t.Errorf("Expected `w` to be '%d' (got '%d')", expected, w)
	}

	output := piper.GetOutput(&piper.pipes[0]).Interface()
	if _, ok := output.(chan int); !ok {
		t.Errorf("Expected `pipe.GetOutput` to return a '%+v' (got '%+v')", "chan int", output)
	}

}

// Benchmarks

func BenchmarkInputOutputSingle(b *testing.B) {
	piper, _ := New(
		P(
			1,
			func(d int) int { return d },
		),
	)

	for i := 0; i < b.N; i++ {
		piper.Input().(chan int) <- 1
		<-piper.Output().(chan int)
	}
}

func BenchmarkInputOutputMultiple(b *testing.B) {
	piper, _ := New(
		P(
			uint(b.N),
			func(d int) int { return d },
		),
	)

    for d := 0; d < b.N; d++ {
	    piper.Input().(chan int) <- 1
    }

	for i := 0; i < b.N; i++ {
		<-piper.Output().(chan int)
	}
}

func BenchmarkInputOutputMultipleBuffered(b *testing.B) {
    var wg sync.WaitGroup

	piper, _ := New(
		P(
			uint(b.N),
			func(d int) int { wg.Wait(); return d },
		),
	)

    wg.Add(1)
    for d := 0; d < b.N; d++ {
	    piper.Input().(chan int) <- 1
    }

    wg.Done()
	for i := 0; i < b.N; i++ {
		<-piper.Output().(chan int)
	}
}

func BenchmarkThousandsPipes(b *testing.B) {
	pipes := make([]Pipe, 10000)

	for i := 0; i < 10000; i++ {
		pipes[i] = P(
			1,
			func(d int) int { return d },
		)
	}

	piper, _ := newPiper(
		pipes,
	)

	for i := 0; i < b.N; i++ {
		piper.Input().(chan int) <- 1
		<-piper.Output().(chan int)
	}
}
