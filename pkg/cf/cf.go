package cf

import (
	"os"
)

// updatePreviousKubeconfig updates the previous kubeconfig file to the given kubeconfig path
func updatePreviousKubeconfig(kubeconfigPath string) error {
	return os.WriteFile(previousKubeconfigConfigPath, []byte(kubeconfigPath), 0644)
}
