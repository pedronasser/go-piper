# go-piper (experimental)

[![Join the chat at https://gitter.im/pedronasser/go-piper](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/pedronasser/go-piper?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

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

    {{ repeat for each Step }}
    piper.P( {{NumberOfWorkers uint}}, {{StepFunction func}} ),
    {{ end }}
    
)

// Input channel
{{piper var}}.Input().(chan {{FirstStep Argument Type}})

// Output channel
{{piper var}}.Output().(chan {{LastStep Return Type}})
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

	// Create a new piper
	pipe, err := piper.New(
		// Now the steps sequentially (...piper.Pipe)
		piper.P(
			2,                                // number of workers (goroutines), in this case 2
			func(d int) int { return d * 2 }, // step's function
			// Translation: Two workers (goroutines) running the step's function will send the output
		),
		piper.P(1, func(d int) int { return d * d }),
		piper.P(2, func(d int) string { return strconv.Itoa(d) }),
	)

	if err != nil {
		panic(err)
	}

	// Defer to close the pipeline at the and of this function
	defer pipe.Close()

    in := pipe.Input().(chan int)
    out := pipe.Output().(chan string)

	// Sending some data to be consumed
	in <- 1 // first data
	in <- 1 // second data

	// Waiting for first ouput
	fmt.Println(<- out)
	// Waiting for second ouput
	fmt.Println(<- out)

}
```

## License

See LICENSE file.