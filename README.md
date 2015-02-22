# go-piper (experimental)

A functional concurrent pipeline builder and management for Go 

For more documentation, please refer to [![GoDoc](https://godoc.org/github.com/pedronasser/go-piper/piper?status.png)](https://godoc.org/github.com/pedronasser/go-piper/piper)

## Install

```bash
go get github.com/pedronasser/go-piper/piper
```

## API Schema

After importing: `github.com/pedronasser/go-piper/piper`

```go
    // Creating new piper

    {{piper var}}, {{error var}} := piper.New(
    
        {{InputChannel Reference chan}},
        {{OutputChannel Reference chan}},
        
        {{ repeat for each Step }}
        piper.P{ {{NumberOfWorkers uint}}, {{StepOutputChannel Reference chan}}, {{StepFunction func}} },
        {{ end }}
        
    )
    
    // Input
    {{InputChannel Reference}} <- {{Data}}
    
    // Output
    <- {{OutputChannel.Reference}}
```

## Example

```go
package main

import (
	"fmt"
	"github.com/pedronasser/go-piper/piper"
	"strconv"
)

func main() {

	// Create an input channel
	input := make(chan int, 2)

	// Create an output channel
	output := make(chan string, 2)

	// Create a new piper
	pipe, err := piper.New(

		input,  // First argument is always the input channel
		output, // Second argument is always the output channel

		// Now the steps sequentially (...*piper.P)
		piper.P{
			2,                                // number of workers (goroutines), in this case 2
			make(chan int, 2),                // step's output channel
			func(d int) int { return d * 2 }, // step's function
			// Translation: Two workers (goroutines) running the step's function will send the output
		},
		piper.P{1, make(chan int), func(d int) int { return d * d }},
		piper.P{2, make(chan string, 2), func(d int) string { return strconv.Itoa(d) }},
	)

	if err != nil {
		panic(err)
	}

	// Defer to close the pipeline at the and of this function
	defer pipe.Close()

	// Sending some data to be consumed
	input <- 1 // first data
	input <- 1 // second data

	// Waiting for first ouput
	fmt.Println(<-output)
	// Waiting for second ouput
	fmt.Println(<-output)
}
```

## License

See LICENSE file.