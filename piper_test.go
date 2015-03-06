package piper

import "testing"

// Build a default Piper
func BuildTestPipe(t *testing.T) (piper Handler) {
	piper, err := New(
		P(
			1,
			func(d interface{}) interface{} { return d.(int) + 1 },
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

	input := piper.Input()
	output := piper.Output()

	expect := 2

	input <- 1
	result := <-output

	if result.(int) != expect {
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
		[]*Pipe{},
	)

	if err != ErrNoPiperFunc {
		t.Errorf("Expected `err` to be '%v' (got '%v')", ErrNoPiperFunc, err)
	}

	piper, _ := newPiper(
		[]*Pipe{
			P(
				0,
				func(d interface{}) interface{} { return d },
			),
		},
	)

	var expected uint = 1000
	piper.pipes[0].SetWorkers(1001)
	w := piper.pipes[0].Workers()
	if w != expected {
		t.Errorf("Expected `w` to be '%d' (got '%d')", expected, w)
	}

	output := piper.GetOutput(piper.pipes[0])
	if _, ok := output.(chan interface{}); !ok {
		t.Errorf("Expected `pipe.GetOutput` to return a '%+v' (got '%+v')", "chan int", output)
	}

}

// Benchmarks

func BenchmarkInputOutputSingle(b *testing.B) {

	fn := func(d interface{}) interface{} { return d.(int) }

	piper, _ := New(
		P(100, fn),
	)
	input := piper.Input()
	output := piper.Output()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input <- 1
		<-output
	}
}

func BenchmarkInputOutputSingleNoReflect(b *testing.B) {

	fn := func(d int) int { return d }
	steps := 1
	workers := 100

	input := make(chan int)
	in := input

	for s := 0; s < steps; s++ {
		sin := in
		out := make(chan int)
		kill := make(chan bool)
		for w := 0; w < workers; w++ {
			go func() {
				for {
					select {
					case v := <-sin:
						out <- fn(v)
					case <-kill:
						return
					}
				}
			}()
		}
		in = out
	}

	output := in

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input <- 1
		<-output
	}
}

func BenchmarkInputOutputMultiple(b *testing.B) {
	piper, _ := New(
		P(
			uint(b.N),
			func(d interface{}) interface{} { return d },
		),
	)

	input := piper.Input()
	output := piper.Output()

	b.ResetTimer()

	for d := 0; d < b.N; d++ {
		input <- 1
	}

	for i := 0; i < b.N; i++ {
		<-output
	}
}

func BenchmarkThousandsPipes(b *testing.B) {
	pipes := make([]*Pipe, 1000)

	for i := 0; i < 1000; i++ {
		pipes[i] = P(
			uint(1),
			func(d interface{}) interface{} { return d },
		)
	}

	piper, _ := newPiper(
		pipes,
	)

	input := piper.Input()
	output := piper.Output()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input <- 1
		<-output
	}
}
