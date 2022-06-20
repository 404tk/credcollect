package main

import (
	"flag"

	"github.com/404tk/credcollect"
)

func parseOptions() *credcollect.Options {
	options := &credcollect.Options{}
	flag.BoolVar(&options.Silent, "silent", false, "silent scan")
	flag.StringVar(&options.Output, "o", "", "output file, -o result.txt")
	flag.Parse()

	return options
}

func main() {
	options := parseOptions()
	options.Enumerate()
}
