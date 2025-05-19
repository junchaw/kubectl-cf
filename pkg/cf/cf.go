package cf

import (
	"flag"
	"path/filepath"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/junchaw/kubectl-cf/pkg/log"
	"github.com/junchaw/kubectl-cf/pkg/sys"
	"github.com/junchaw/kubectl-cf/pkg/term"
	"github.com/junchaw/kubectl-cf/pkg/translations"
)

const (
	// DefaultKubeconfigBaseName is the base name of the file that will be used
	// as a suggestion for backup when kubeconfig is not a symlink.
	DefaultKubeconfigBaseName = "default-kubeconfig"

	// KubeconfigSpecialPathKubeconfigDir is the special path that tells kubectl-cf to
	// read kubeconfig files from the directory of the given kubeconfig file.
	KubeconfigSpecialPathKubeconfigDir = "@kubeconfig-dir"

	// KubeconfigFilenameMatchPatternStrDefault is the default regex pattern for kubeconfig filename
	KubeconfigFilenameMatchPatternStrDefault = `^(?P<name>(config)|([^\.]+\.yaml))$`

	// KubeconfigFilenameMatchPatternNameGroup is the name of the regex group for kubeconfig name, "(?P<name>...)"
	KubeconfigFilenameMatchPatternNameGroup = "name"

	// PreviousKubeconfigFullPath is the file name which stores the previous kubeconfig file's full path
	PreviousKubeconfigFullPath = "previous"
)

var logger = log.DefaultLogger

// t is the translation function
var t = translations.T

var (
	warning = term.MakeFgStyle("1")   // red
	info    = term.MakeFgStyle("28")  // blue
	text    = term.MakeFgStyle("255") // white
)

var (
	homeDir = sys.HomeDir()
	kubeDir = filepath.Join(homeDir, ".kube")

	// kubectlCfConfigDir is the directory for kubectl-cf config files
	kubectlCfConfigDir           = "" // will be set in init()
	previousKubeconfigConfigPath = "" // will be set in init()

	// kubeconfigPaths is the list of kubeconfig paths,
	// parsed from environment variable KUBECTL_CF_PATHS, works like PATH environment variable
	kubeconfigPaths = []string{} // will be set in init()

	// kubeconfigDir is the directory for kubeconfig, for example, ~/.kube
	kubeconfigDir = "" // will be set in init()

	// kubeconfigPath is the current kubeconfig path, for example, ~/.kube/config
	kubeconfigPath = "" // will be set in init()

	// kubeconfigFilenameMatchPattern defines the name pattern of kubeconfig files,
	// it comes with a default value,
	// and can be overriden by environment variable KUBECONFIG_FILENAME_MATCH_PATTERN
	kubeconfigFilenameMatchPattern *regexp.Regexp = nil // will be set in init()
)
var Modal = &KubectlCfModal{}

func Run() error {
	flag.Parse()

	p := tea.NewProgram(Modal)

	_, err := p.Run()
	return err
}
