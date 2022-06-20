package main

import (
	"encoding/json"
	"fmt"

	"github.com/404tk/credcollect"
)

func main() {
	options := &credcollect.Options{Silent: true}
	res := options.Enumerate()
	r, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r))
}
