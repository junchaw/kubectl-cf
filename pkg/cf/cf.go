package cf

import (
	"flag"
	"os"
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
