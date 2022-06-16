package main

import (
	"encoding/json"
	"fmt"

	"github.com/404tk/credcollect/runner"
)

func main() {
	options := &runner.Options{Silent: true}
	res := options.Enumerate()
	r, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r))
}
