package main

import(
	"github.com/spf13/pflag"
)

var (
	pflag.String("actor-address", "localhost:8080", "The address for reaching the actor service")
)

func main() {
	pflag.Parse()
}