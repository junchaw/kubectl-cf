package main

import (
	"flag"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/junchaw/kubectl-cf/pkg"
)

func main() {
	flag.Parse()

	p := tea.NewProgram(pkg.InitialModel)
	if err := p.Start(); err != nil {
		panic(err)
	}
}
