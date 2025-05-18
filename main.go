package main

import (
	"flag"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/junchaw/kubectl-cf/pkg/cf"
)

func main() {
	flag.Parse()

	p := tea.NewProgram(cf.Modal)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
