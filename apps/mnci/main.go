package main

import (
	"fmt"

	"github.com/MoNecromanCI/mnci/apps/mnci/hello"
)

func main() {
	fmt.Println(hello.Greet("world").Message)
}
