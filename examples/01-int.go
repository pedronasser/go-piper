package main

import (
	"fmt"
	"strconv"

	"github.com/pedronasser/go-piper"
)

func main() {

	pipe, err := piper.New(
		piper.P(1,
			func(d interface{}) interface{} {
				var i int = d.(int)
				var r int = i * i
				return r
			},
		),
		piper.P(1,
			func(d interface{}) interface{} {
				var i int = d.(int)
				var r string = strconv.Itoa(i)
				return r
			},
		),
	)

	if err != nil {
		panic(err)
	}

	defer pipe.Close()

	in := pipe.Input()
	out := pipe.Output()

	in <- 1
	in <- 1

	fmt.Println((<-out).(string))
	fmt.Println((<-out).(string))

}
