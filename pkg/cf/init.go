package cf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func init() {
	kubectlCfConfigDir = os.Getenv("KUBECTL_CF_CONFIG_DIR")
	if kubectlCfConfigDir == "" {
		kubectlCfConfigDir = filepath.Join(kubeDir, "kubectl-cf")
	}

	previousKubeconfigConfigPath = filepath.Join(kubectlCfConfigDir, PreviousKubeconfigFullPath)

	kubeconfigPath = os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(kubeDir, "config")
	}

	kubeconfigDir = filepath.Dir(kubeconfigPath)

	flag.Usage = func() {
		_, _ = fmt.Fprint(flag.CommandLine.Output(), t("cfUsage"))
		flag.PrintDefaults()
	}

	kubeconfigFilenameMatchPatternStr := os.Getenv("KUBECTL_CF_KUBECONFIG_FILENAME_MATCH_PATTERN")
	if kubeconfigFilenameMatchPatternStr == "" {
		kubeconfigFilenameMatchPatternStr = `^(?P<name>(config)|([^\.]+\.yaml))$`
	}
	kubeconfigFilenameMatchPattern = regexp.MustCompile(kubeconfigFilenameMatchPatternStr)

	// ensure config dir exists
	if _, err := os.Lstat(kubectlCfConfigDir); err != nil {
		if os.IsNotExist(err) {
			logger.Debugf("Default config dir %s not exist, creating", kubectlCfConfigDir)
			if err := os.Mkdir(kubectlCfConfigDir, 0755); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}
