package cf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/junchaw/kubectl-cf/pkg/sys"
)

const (
	ModeSelect = iota
	ModeAskIfRenameKubeconfig
	ModeQuit
)

type KubectlCfModal struct {
	mode int

	// candidates is a list of (Candidate/kubeconfig)s
	candidates []Candidate

	// currentKubeconfigPath is the full path of current kubeconfig, could be empty
	currentKubeconfigPath string

	// kubeconfigPathSuggestion is the suggestion for the kubeconfig path,
	// used in mode: ModeAskIfRenameKubeconfig
	kubeconfigPathSuggestion string

	// cursor indicates which candidate our cursor is pointing at
	// used in mode: ModeSelect
	cursor int

	// farewell is the message which will be printed before quitting
	// used in mode: ModeQuit
	farewell string
}

func (modal *KubectlCfModal) quit(farewell string) tea.Cmd {
	if !strings.HasSuffix(farewell, "\n") {
		farewell += "\n" // there must be a "\n" at the end of message
	}
	modal.mode = ModeQuit
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

func (modal *KubectlCfModal) refreshCandidates() error {
	candidates, err := ListKubeconfigCandidatesInDir(kubeconfigDir)
	if err != nil {
		return err
	}
	modal.candidates = candidates
	return nil
}

func (modal *KubectlCfModal) focusOnCurrentKubeconfig() {
	for index, candidate := range modal.candidates {
		if candidate.FullPath == modal.currentKubeconfigPath {
			modal.cursor = index
		}
	}
}

func (modal *KubectlCfModal) refreshCandidatesAndFocusOnCurrentKubeconfig() error {
	if err := modal.refreshCandidates(); err != nil {
		return err
	}
	modal.focusOnCurrentKubeconfig()
	return nil
}

func (modal *KubectlCfModal) Init() tea.Cmd {
	if len(flag.Args()) > 1 {
		return modal.quit(t("wrongNumberOfArgumentExpect", 1))
	}
	kubeconfigArg := flag.Arg(0)

	info, err := os.Lstat(kubeconfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		if err := sys.CreateSymlink("", kubeconfigPath); err != nil {
			panic(err)
		}
		modal.mode = ModeSelect
		logger.Infof("The kubeconfig not exist, created an empty symlink: %s", kubeconfigPath)
	} else {
		if sys.IsSymlink(info) {
			target, err := os.Readlink(kubeconfigPath)
			if err != nil {
				panic(err)
			}
			modal.mode = ModeSelect
			modal.currentKubeconfigPath = target
		} else {
			logger.Infof("The kubeconfig is not a symlink, need to ask user for confirmation")
			kubeconfigPathSuggestion, err := sys.GenerateBackUpName(filepath.Join(kubeconfigDir, DefaultKubeconfigBaseName), ".yaml")
			if err != nil {
				panic(err)
			}
			modal.mode = ModeAskIfRenameKubeconfig
			modal.kubeconfigPathSuggestion = kubeconfigPathSuggestion
			// return here, will parse candicates after confirmed
			return nil
		}
	}

	if err := modal.refreshCandidatesAndFocusOnCurrentKubeconfig(); err != nil {
		return modal.quit(warning(t("unableToRefreshCandidates", err.Error())))
	}

	if kubeconfigArg != "" {
		if kubeconfigArg == "-" {
			f, err := os.ReadFile(previousKubeconfigConfigPath)
			if err != nil {
				if !os.IsNotExist(err) {
					panic(err)
				}
				return modal.quit(warning(t("noPreviousKubeconfig")))
			}
			return modal.quit(modal.symlinkConfigPathTo(string(f)))
		}

		var guessCandidates []Candidate
		for _, candidate := range modal.candidates {
			if candidate.Name == kubeconfigArg {
				guessCandidates = []Candidate{candidate}
				break
			}
			if strings.HasPrefix(candidate.Name, kubeconfigArg) { // guess candidates by prefix
				guessCandidates = append(guessCandidates, candidate)
			}
		}

		if guessCandidates == nil {
			return modal.quit(warning(t("noMatchFound", kubeconfigArg)))
		}

		if len(guessCandidates) == 1 { // if there is only one guess candidate, use it
			return modal.quit(modal.symlinkConfigPathTo(guessCandidates[0].FullPath))
		}

		var names []string // if there are multiple guess candidates, show the names
		for _, g := range guessCandidates {
			names = append(names, g.Name)
		}
		return modal.quit(warning(t("moreThanOneMatchesFound", kubeconfigArg, strings.Join(names, ", "))))
	}

	return nil
}

func (modal *KubectlCfModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg: // Is it a key press?
		switch msg.String() { // The key pressed
		case "ctrl+c", "q", "esc": // These keys should exit the program.
			return modal, tea.Quit
		}
	}

	switch modal.mode {

	case ModeQuit:
		return modal, nil

	case ModeAskIfRenameKubeconfig:
		switch msg := msg.(type) {
		case tea.KeyMsg: // Is it a key press?
			switch msg.String() { // The key pressed
			case "y", "Y":
				if err := os.Rename(kubeconfigPath, modal.kubeconfigPathSuggestion); err != nil {
					return modal, modal.quit(warning(t("renameKubeconfigError", err.Error())))
				}
				if err := sys.CreateSymlink(modal.kubeconfigPathSuggestion, kubeconfigPath); err != nil {
					return modal, modal.quit(warning(t("createSymlinkError", err.Error())))
				}
				modal.currentKubeconfigPath = modal.kubeconfigPathSuggestion
				if err := modal.refreshCandidatesAndFocusOnCurrentKubeconfig(); err != nil {
					return modal, modal.quit(warning(t("unableToRefreshCandidates", err.Error())))
				}
				modal.mode = ModeSelect
				return modal, nil
			case "n", "N":
				return modal, modal.quit(t("renameKubeconfigCanceled"))
			}
		}
		return modal, nil

	case ModeSelect:
		switch msg := msg.(type) {
		case tea.KeyMsg: // Is it a key press?
			switch msg.String() { // The key pressed
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
		return modal, nil

	default:
		return modal, nil
	}
}

func (modal *KubectlCfModal) View() string {
	switch modal.mode {
	case ModeAskIfRenameKubeconfig:
		return fmt.Sprint(t("notASymlinkDoYouWantToMoveIt", info(kubeconfigPath), info(modal.kubeconfigPathSuggestion)))

	case ModeQuit:
		return modal.farewell

	case ModeSelect:
		content := t("whatKubeconfig") + "\n\n"

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
			candicateLine := fmt.Sprintf(" %-*s %s%s\n", longestName, candidate.Name, candidate.FullPath, suffix)
			if candidate.FullPath == modal.currentKubeconfigPath {
				candicateLine = info(candicateLine)
			} else {
				candicateLine = text(candicateLine) // we need to set the style for normal text to override active style
			}
			content += candicateLine
		}

		content += subtle("\n" + t("helpActions") + "\n") // The footer
		return content

	default:
		return "Unknown mode"
	}
}
