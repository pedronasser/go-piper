# go-piper (v1.0.0)

[![Join the chat at https://gitter.im/pedronasser/go-piper](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/pedronasser/go-piper?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

An easy way to build your Go programs using a pipeline pattern using channels and multiple goroutines.

This package creates N goroutines for each pipeline step all connected through unbuffered channels and manages everything.

For more documentation, please refer to [![GoDoc](https://godoc.org/github.com/pedronasser/go-piper?status.png)](https://godoc.org/github.com/pedronasser/go-piper)

## Install

```bash
go get github.com/pedronasser/go-piper
```

## Example

```go
package main

import (
        "fmt"
        "strconv"

        // Import go-piper
        "github.com/pedronasser/go-piper"
)

func main() {
        // Create new Piper
        pipe, err := piper.New(

                // Creating first step
                piper.P(1, // Number of workers
                        
                        // First step's function
                        func(d interface{}) interface{} { // Should always receive and return interface{}
                                var i int = d.(int) // Asserting `d` as integer
                                var r int = i * i
                                return r
                        },
                ),

                // Creating second step
                piper.P(1, // Number of workers

                        // Second step's function
                        func(d interface{}) interface{} { // Should always receive and return interface{}
                                var i int = d.(int) // Asserting `d ` as integer
                                var r string = strconv.Itoa(i)
                                return r // returning as string
                        },
                ),
        )

        // Error check
        if err != nil {
                panic(err)
        }

        // Defering close
        defer pipe.Close()

        // Getting input and output channels
        in := pipe.Input()
        out := pipe.Output()

        in <- 1 // Sending first data
        in <- 1 // Sending second data

        fmt.Println((<-out).(string)) // Receiving first result
        fmt.Println((<-out).(string)) // Receiving second result
}
```

## Other examples

- https://github.com/pedronasser/go-piper/blob/master/examples/

## License

See LICENSE file.
