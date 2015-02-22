package piper

import (
	"testing"
)

// Build a default int-typed Piper
func BuildTestPipe(t *testing.T) (piper Handler) {
	piper, err := New(
		make(chan int),
		make(chan int),
		P{
			1,
			make(chan int),
			func(d int) int { return d + 1 },
		},
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
		make(chan int),
		make(chan int),
		[]P{},
	)

	if err != ErrNoPiperFunc {
		t.Errorf("Expected `err` to be '%v' (got '%v')", ErrNoPiperFunc, err)
	}

	piper, _ := newPiper(
		make(chan int),
		make(chan int),
		[]P{
			P{
				0,
				make(chan int),
				func(d int) int { return d },
			},
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

	_, err2 := newPiper(
		make(chan int),
		make(chan int),
		[]P{
			P{
				1,
				make(chan string),
				func(d int) int { return d },
			},
		},
	)

	if err2 != ErrInvalidPipeChannel {
		t.Errorf("Expected `err` to be '%+v' (got '%+v')", ErrInvalidPipeChannel, err2)
	}

}

func TestNewPipeInvalidOutput(t *testing.T) {
	_, err := newPiper(
		1,
		make(chan int),
		[]P{
			P{
				1,
				make(chan int),
				func(d int) int { return d + 1 },
			},
		},
	)

	if err != ErrInvalidInputChannel {
		t.Errorf("Expected `err` to be '%v' (got '%v')", ErrInvalidInputChannel, err)
	}

	_, err = newPiper(
		make(chan int),
		1,
		[]P{
			P{
				1,
				make(chan int),
				func(d int) int { return d + 1 },
			},
		},
	)

	if err != ErrInvalidOutputChannel {
		t.Errorf("Expected `err` to be '%v' (got '%v')", ErrInvalidOutputChannel, err)
	}
}

// Benchmarks

func BenchmarkInputOutput(b *testing.B) {
	piper, _ := New(
		make(chan int),
		P{
			1,
			make(chan int),
			func(d int) int { return d },
		},
	)

	for i := 0; i < b.N; i++ {
		piper.Input().(chan int) <- 1
		<-piper.Output().(chan int)
	}
}

func BenchmarkThousandsPipes(b *testing.B) {
	pipes := make([]P, 10000)

	for i := 0; i < 10000; i++ {
		pipes[i] = P{
			1,
			make(chan int),
			func(d int) int { return d },
		}
	}

	piper, _ := newPiper(
		make(chan int),
		make(chan int),
		pipes,
	)

	for i := 0; i < b.N; i++ {
		piper.Input().(chan int) <- 1
		<-piper.Output().(chan int)
	}
}
