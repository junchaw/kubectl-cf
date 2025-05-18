package cf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/junchaw/kubectl-cf/pkg/log"
	"github.com/junchaw/kubectl-cf/pkg/sys"
	"github.com/junchaw/kubectl-cf/pkg/term"
	"github.com/junchaw/kubectl-cf/pkg/translations"
)

const (
	// DefaultKubeconfigBaseName is the base name of the file that will be used
	// as a suggestion for backup when kubeconfig is not a symlink.
	DefaultKubeconfigBaseName = "default-kubeconfig"

	// KubeconfigFilenameMatchPatternNameGroup is the name of the regex group for kubeconfig name, "(?P<name>...)"
	KubeconfigFilenameMatchPatternNameGroup = "name"

	// PreviousKubeconfigFullPath is the file name which stores the previous kubeconfig file's full path
	PreviousKubeconfigFullPath = "previous"

	// CursorMark is the mark for the cursor (current selection) in the list
	CursorMark = ">"

	// CurrentKubeconfigMark is the mark for the current kubeconfig in the list
	CurrentKubeconfigMark = "*"
)

var logger = log.DefaultLogger

// t is the translation function
var t = translations.T

var (
	warning = term.MakeFgStyle("1")   // red
	info    = term.MakeFgStyle("28")  // blue
	subtle  = term.MakeFgStyle("241") // gray
	text    = term.MakeFgStyle("255") // white
)

var (
	homeDir = sys.HomeDir()
	kubeDir = filepath.Join(homeDir, ".kube")

	kubectlCfConfigDir           = "" // will be set in init()
	previousKubeconfigConfigPath = "" // will be set in init()

	kubeconfigDir  = "" // will be set in init()
	kubeconfigPath = "" // will be set in init()

	// kubeconfigFilenameMatchPattern defines the name pattern of kubeconfig files
	kubeconfigFilenameMatchPattern *regexp.Regexp = nil // will be set in init()
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
