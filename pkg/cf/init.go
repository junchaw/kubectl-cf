package cf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	kubectlCfConfigDir = os.Getenv("KUBECTL_CF_CONFIG_DIR")
	if kubectlCfConfigDir == "" {
		kubectlCfConfigDir = filepath.Join(kubeDir, "kubectl-cf")
	}

	previousKubeconfigConfigPath = filepath.Join(kubectlCfConfigDir, PreviousKubeconfigFullPath)
	var filteredPaths []string // Filter out empty items
	for path := range strings.SplitSeq(os.Getenv("KUBECTL_CF_PATHS"), ":") {
		if path != "" {
			filteredPaths = append(filteredPaths, path)
		}
	}
	kubeconfigPaths = filteredPaths
	if len(kubeconfigPaths) == 0 { // by default, read kubeconfig files from the directory of the given kubeconfig file
		kubeconfigPaths = []string{KubeconfigSpecialPathKubeconfigDir}
	}

	kubeconfigPath = os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(kubeDir, "config")
	}

	kubeconfigDir = filepath.Dir(kubeconfigPath)

	flag.Usage = func() {
		_, _ = fmt.Fprint(flag.CommandLine.Output(), t("cfUsage"))
		flag.PrintDefaults()
	}

	kubeconfigFilenameMatchPatternStr := os.Getenv("KUBECTL_CF_KUBECONFIG_MATCH_PATTERN")
	if kubeconfigFilenameMatchPatternStr == "" {
		kubeconfigFilenameMatchPatternStr = KubeconfigFilenameMatchPatternStrDefault
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
