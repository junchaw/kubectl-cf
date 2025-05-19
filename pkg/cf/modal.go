package cf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/junchaw/kubectl-cf/pkg/sys"
)

const (
	// Any mode means the modal is ready.

	ModeSelect = iota
	ModeAskIfRenameKubeconfig
	ModeQuit
)

type KubectlCfModal struct {
	mode int

	list list.Model

	// candidates is a list of (Candidate/kubeconfig)s
	candidates []Candidate

	// currentKubeconfigPath is the full path of current kubeconfig, could be empty
	currentKubeconfigPath string

	// kubeconfigPathSuggestion is the suggestion for the kubeconfig path,
	// used in mode: ModeAskIfRenameKubeconfig
	kubeconfigPathSuggestion string

	// farewell is the message which will be printed before quitting
	// used in mode: ModeQuit
	farewell string
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

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
	items := make([]list.Item, len(candidates))
	for i, c := range candidates {
		items[i] = c
	}
	modal.list.SetItems(items)
	return nil
}

func (modal *KubectlCfModal) focusOnCurrentKubeconfig() {
	for index, candidate := range modal.candidates {
		if candidate.FullPath == modal.currentKubeconfigPath {
			modal.list.Select(index)
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

	list := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0) // will set later
	list.Title = t("whatKubeconfig")
	list.SetShowPagination(true)
	list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}
	modal.list = list

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
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			modal.list.SetSize(msg.Width-h, msg.Height-v)
		case tea.KeyMsg: // Is it a key press?
			switch msg.String() { // The key pressed
			case "enter": // The "enter" key selects the current candidate
				return modal, modal.quit(modal.symlinkConfigPathTo(modal.list.SelectedItem().(Candidate).FullPath))
			}
		}

		updatedList, cmd := modal.list.Update(msg)
		modal.list = updatedList

		return modal, cmd

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
		return modal.list.View()

	default:
		return ""
	}
}
