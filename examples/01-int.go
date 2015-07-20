// This example just shows a way of creating a single 2-step pipeline.
// Each step will have only one worker.
// Both steps will input and output integers.

package simpleint

import (
	"fmt"
	"strconv"

	"github.com/pedronasser/go-piper"
)

func main() {

	pipe, err := piper.New(
		piper.P(1,
			func(d interface{}) interface{} {
				var i int
				var ok bool

				if _, ok = d.(int); !ok {
					// If not integer, discard
					return nil
				}

				r := i * i

				return r
			},
		),
		piper.P(1,
			func(d interface{}) interface{} {
				var i int
				var ok bool

				if _, ok = d.(int); !ok {
					// If not integer, discard
					return nil
				}

				r := strconv.Itoa(i)
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
