package pkg

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/junchaw/kubectl-cf/pkg/kubeconfigs"
	"github.com/junchaw/kubectl-cf/pkg/utils"
	"github.com/sirupsen/logrus"
)

const (
	// PreviousKubeconfigFullPath is the file name which stores the previous kubeconfig file's full path
	PreviousKubeconfigFullPath = "previous"
)

type Modal struct {
	// meta is extra information displayed on top of the output
	meta []string

	// candidates is a list of (Candidate/kubeconfig)s
	candidates []kubeconfigs.Candidate

	// cursor indicates which candidate our cursor is pointing at
	cursor int

	quitting bool

	// farewell is the message which will be printed before quitting
	farewell string

	// currentConfigPath is the full path of current kubeconfig
	currentConfigPath string
}

var InitialModal = &Modal{}

var (
	Warning = utils.MakeFgStyle("1")   // red
	Info    = utils.MakeFgStyle("28")  // blue
	Subtle  = utils.MakeFgStyle("241") // gray
)

var (
	homeDir               = utils.HomeDir()
	kubeDir               = filepath.Join(homeDir, ".kube")
	defaultKubeconfigPath = filepath.Join(kubeDir, "config")
	kubeconfigPath        = filepath.Join(kubeDir, "config") // same as defaultKubeconfigPath for now, maybe allow user to specify
	cfDir                 = filepath.Join(kubeDir, "kubectl-cf")
)

func (m *Modal) quit(farewell string) tea.Cmd {
	if !strings.HasSuffix(farewell, "\n") {
		farewell += "\n" // there must be a "\n" at the end of message
	}
	m.quitting = true
	m.farewell = farewell
	return tea.Quit
}

func (m *Modal) symlinkConfigPathTo(name string) string {
	if err := os.WriteFile(filepath.Join(cfDir, PreviousKubeconfigFullPath), []byte(m.currentConfigPath), 0644); err != nil {
		panic(err)
	}

	err := Symlink(name, kubeconfigPath)
	if err != nil {
		return Warning(t("createSymlinkError", err))
	}
	s := t("symlinkNowPointTo", Info(kubeconfigPath), Info(name))
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if !(kubeconfigEnv == kubeconfigPath || (kubeconfigPath == defaultKubeconfigPath && kubeconfigEnv == "")) {
		s += "\n" + Warning(t("kubeconfigEnvWarning", kubeconfigPath))
	}
	return s
}

func (modal *Modal) Init() tea.Cmd {
	if len(flag.Args()) > 1 {
		return modal.quit(t("wrongNumberOfArgumentExpect1"))
	}

	candidates, err := kubeconfigs.ListKubeconfigCandidatesInDir(kubeDir)
	if err != nil {
		panic(err)
	}
	InitialModal.candidates = candidates

	logger.Debugf("Path to config symlink: %s", kubeconfigPath)

	info, err := os.Lstat(kubeconfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		logger.Debugf("The symlink not exist, using the default kubeconfig: %s", defaultKubeconfigPath)
		InitialModal.currentConfigPath = defaultKubeconfigPath
	} else {
		if utils.IsSymlink(info) {
			target, err := os.Readlink(kubeconfigPath)
			if err != nil {
				panic(err)
			}
			logger.Debugf("The symlink points to: %s", target)
			InitialModal.currentConfigPath = target
		} else {
			logger.Debugf("The symlink is not a symlink")
			return modal.quit(Warning(t("kubeconfigNotSymlink", kubeconfigPath)))
		}
	}
	logger.Debugf("Current using kubeconfig: %s", InitialModal.currentConfigPath)

	if logger.GetLevel() == logrus.DebugLevel {
		f, err := os.Open(filepath.Join(cfDir, PreviousKubeconfigFullPath))
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			logger.Debugf("No previous kubeconfig")
		} else {
			b, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			logger.Debugf("Previous kubeconfig: %s", string(b))
		}
	}

	if search := flag.Arg(0); search != "" {
		if search == "-" {
			f, err := os.Open(filepath.Join(cfDir, PreviousKubeconfigFullPath))
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				return modal.quit(Warning(t("noPreviousKubeconfig")))
			}
			b, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			return modal.quit(modal.symlinkConfigPathTo(string(b)))
		}

		var guess []kubeconfigs.Candidate
		for _, candidate := range candidates {
			if candidate.Name == search {
				guess = []kubeconfigs.Candidate{candidate}
				break
			}
			if strings.HasPrefix(candidate.Name, search) {
				guess = append(guess, candidate)
			}
		}

		if guess == nil {
			return modal.quit(Warning(t("noMatchFound", search)))
		}

		if len(guess) == 1 {
			return modal.quit(modal.symlinkConfigPathTo(guess[0].FullPath))
		}

		var s []string
		for _, g := range guess {
			s = append(s, g.Name)
		}
		return modal.quit(Warning(t("moreThanOneMatchesFound", search, strings.Join(s, ", "))))
	}

	// focus on current config path
	for key, candidate := range candidates {
		if candidate.FullPath == modal.currentConfigPath {
			modal.cursor = key
		}
	}

	return nil
}

func (modal *Modal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case tea.KeyMsg:
		// The key pressed
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q", "esc":
			return modal, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if modal.cursor > 0 {
				modal.cursor--
			} else {
				modal.cursor = len(modal.candidates) - 1
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if modal.cursor < len(modal.candidates)-1 {
				modal.cursor++
			} else {
				modal.cursor = 0
			}

		case "enter":
			return modal, modal.quit(modal.symlinkConfigPathTo(modal.candidates[modal.cursor].FullPath))
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return modal, nil
}

func (modal *Modal) View() string {
	// The header
	s := strings.Join(modal.meta, "\n") + "\n"

	if modal.quitting {
		return s + modal.farewell
	}

	s += t("whatKubeconfig") + "\n\n"

	// Iterate over our candidates
	longestName := 0
	for _, candidate := range modal.candidates {
		if len(candidate.Name) > longestName {
			longestName = len(candidate.Name)
		}
	}
	for key, candidate := range modal.candidates {
		cursor := " "
		if modal.cursor == key {
			cursor = ">"
		}
		s += cursor

		suffix := ""
		if candidate.FullPath == modal.currentConfigPath {
			suffix = "*"
		}
		tmpl := fmt.Sprintf(" %%-%ds %%s%%s\n", longestName)
		ts := fmt.Sprintf(tmpl, candidate.Name, candidate.FullPath, suffix)
		if candidate.FullPath == modal.currentConfigPath {
			ts = Info(ts)
		}
		s += ts
	}

	// The footer
	s += Subtle("\n" + t("helpActions") + "\n")
	return s
}

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprint(flag.CommandLine.Output(), t("cfUsage"))
		flag.PrintDefaults()
	}

	// ensure config dir exists
	if _, err := os.Lstat(cfDir); err != nil {
		if os.IsNotExist(err) {
			logger.Debugf("Default config dir %s not exist, creating", cfDir)
			if err := os.Mkdir(cfDir, 0755); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}
