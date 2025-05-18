package cf

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/junchaw/kubectl-cf/pkg/sys"
)

type KubectlCfModal struct {
	// meta is extra information displayed on top of the output
	meta []string

	// candidates is a list of (Candidate/kubeconfig)s
	candidates []Candidate

	// cursor indicates which candidate our cursor is pointing at
	cursor int

	// quitting is a flag to indicate if the program is quitting
	quitting bool

	// farewell is the message which will be printed before quitting
	farewell string

	// currentKubeconfigPath is the full path of current kubeconfig
	currentKubeconfigPath string
}

var Modal = &KubectlCfModal{}

func (modal *KubectlCfModal) quit(farewell string) tea.Cmd {
	if !strings.HasSuffix(farewell, "\n") {
		farewell += "\n" // there must be a "\n" at the end of message
	}
	modal.quitting = true
	modal.farewell = farewell
	return tea.Quit
}

func (modal *KubectlCfModal) symlinkConfigPathTo(name string) string {
	if err := updatePreviousKubeconfig(modal.currentKubeconfigPath); err != nil {
		return warning(t("updatePreviousKubeconfigError", err.Error()))
	}
	if err := sys.CreateSymlink(name, kubeconfigPath); err != nil {
		return warning(t("createSymlinkError", err.Error()))
	}
	return text(t("symlinkNowPointTo", info(kubeconfigPath), info(name)))
}

func (modal *KubectlCfModal) Init() tea.Cmd {
	if len(flag.Args()) > 1 {
		return modal.quit(t("wrongNumberOfArgumentExpect", 1))
	}
	kubeconfigArg := flag.Arg(0)

	candidates, err := ListKubeconfigCandidatesInDir(kubeDir)
	if err != nil {
		panic(err)
	}
	Modal.candidates = candidates

	info, err := os.Lstat(kubeconfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		logger.Warnf("Kubeconfig not exist: %s", kubeconfigPath)
	} else {
		if sys.IsSymlink(info) {
			target, err := os.Readlink(kubeconfigPath)
			if err != nil {
				panic(err)
			}
			Modal.currentKubeconfigPath = target
		} else {
			logger.Debugf("The symlink is not a symlink")
			return modal.quit(warning(t("kubeconfigNotSymlink", kubeconfigPath)))
		}
	}

	if kubeconfigArg != "" {
		if kubeconfigArg == "-" {
			f, err := os.Open(previousKubeconfigConfigPath)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				return modal.quit(warning(t("noPreviousKubeconfig")))
			}
			b, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			return modal.quit(modal.symlinkConfigPathTo(string(b)))
		}

		var guess []Candidate
		for _, candidate := range candidates {
			if candidate.Name == kubeconfigArg {
				guess = []Candidate{candidate}
				break
			}
			if strings.HasPrefix(candidate.Name, kubeconfigArg) {
				guess = append(guess, candidate)
			}
		}

		if guess == nil {
			return modal.quit(warning(t("noMatchFound", kubeconfigArg)))
		}

		if len(guess) == 1 {
			return modal.quit(modal.symlinkConfigPathTo(guess[0].FullPath))
		}

		var s []string
		for _, g := range guess {
			s = append(s, g.Name)
		}
		return modal.quit(warning(t("moreThanOneMatchesFound", kubeconfigArg, strings.Join(s, ", "))))
	}

	// focus on current kubeconfig path
	for index, candidate := range candidates {
		if candidate.FullPath == modal.currentKubeconfigPath {
			modal.cursor = index
		}
	}

	return nil
}

func (modal *KubectlCfModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg: // Is it a key press?
		switch msg.String() { // The key pressed
		case "ctrl+c", "q", "esc": // These keys should exit the program.
			return modal, tea.Quit

		case "up", "k": // The "up" and "k" keys move the cursor up
			if modal.cursor > 0 {
				modal.cursor--
			} else {
				modal.cursor = len(modal.candidates) - 1
			}

		case "down", "j": // The "down" and "j" keys move the cursor down
			if modal.cursor < len(modal.candidates)-1 {
				modal.cursor++
			} else {
				modal.cursor = 0
			}

		case "enter": // The "enter" key selects the current candidate
			return modal, modal.quit(modal.symlinkConfigPathTo(modal.candidates[modal.cursor].FullPath))
		}
	}

	return modal, nil // Return the updated model to the Bubble Tea runtime for processing.
}

func (modal *KubectlCfModal) View() string {
	content := strings.Join(modal.meta, "\n") + "\n" // The header

	if modal.quitting {
		return content + modal.farewell
	}

	content += t("whatKubeconfig") + "\n\n"

	longestName := 0
	for _, candidate := range modal.candidates { // Iterate over our candidates to find the longest name
		if len(candidate.Name) > longestName {
			longestName = len(candidate.Name)
		}
	}
	for key, candidate := range modal.candidates { // Iterate over our candidates to display them
		cursor := " " // The cursor at the beginning of the line
		if modal.cursor == key {
			cursor = CursorMark
		}
		content += cursor

		suffix := ""
		if candidate.FullPath == modal.currentKubeconfigPath {
			suffix = CurrentKubeconfigMark
		}
		tmpl := fmt.Sprintf(" %%-%ds %%s%%s\n", longestName)
		candicateLine := fmt.Sprintf(tmpl, candidate.Name, candidate.FullPath, suffix)
		if candidate.FullPath == modal.currentKubeconfigPath {
			candicateLine = info(candicateLine)
		} else {
			candicateLine = text(candicateLine) // we need to set the style for normal text to override active style
		}
		content += candicateLine
	}

	content += subtle("\n" + t("helpActions") + "\n") // The footer
	return content
}
