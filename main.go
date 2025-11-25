package main

import (
	"github.com/jack-kitto/yoink/cmd"
)

var version = "dev"

func main() {
	cmd.Execute(version)
}
