package main

import (
	"github.com/junchaw/kubectl-cf/pkg/cf"
)

func main() {
	if err := cf.Run(); err != nil {
		panic(err)
	}
}
