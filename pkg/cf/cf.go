package cf

import (
	"flag"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var Modal = &KubectlCfModal{}

func Run() error {
	flag.Parse()

	p := tea.NewProgram(Modal)

	_, err := p.Run()
	return err
}

// updatePreviousKubeconfig updates the previous kubeconfig file to the given kubeconfig path
func updatePreviousKubeconfig(kubeconfigPath string) error {
	return os.WriteFile(previousKubeconfigConfigPath, []byte(kubeconfigPath), 0644)
}
