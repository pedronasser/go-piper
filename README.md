# go-piper

A functional pipeliner builder written in Go.

For usage, please refer to [![GoDoc](https://godoc.org/github.com/pedronasser/go-piper/piper?status.png)](https://godoc.org/github.com/pedronasser/go-piper/piper)

## Install

```bash
go get github.com/pedronasser/go-piper/piper
```

## Example

```go
package main

import (
    "fmt"
    "strconv"

    "github.com/pedronasser/go-piper/piper"
)

func main () {

    pipe, err := piper.NewPiper(
        make(chan string),
        &piper.F{ make(chan int), func (d int) int { return d * 2 } },
        &piper.F{ make(chan int), func (d int) int {  return d * d } },
        &piper.F{ make(chan int), func (d int) string { return strconv.Itoa(d) } },
    )

    if err != nil {
        panic(err)
    }

    var result string
    pipe.In(2, &result)

    fmt.Println(result)
}
```

## License

See LICENSE file.