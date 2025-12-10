package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"git-opener/pkg/tmuxxer"

	"github.com/gdamore/tcell/v2" // Import tcell for handling key events
	"github.com/rivo/tview"
)

// Configuration variables
var (
	// Path to the git repositories
	gitReposPath string
	// Path to the log file
	logFilePath = "/tmp/app.log"
)

func init() {
	// Get the GIT_REPOS_PATH environment variable
	gitReposPath = os.Getenv("GIT_REPOS_PATH")

	// If GIT_REPOS_PATH is not set or empty, use the default ~/git
	if gitReposPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			gitReposPath = homeDir + "/git"
		} else {
			// Fallback
			gitReposPath = "~/git"
		}
	}

	// Open log file and set it as the default log output
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}

	// Set the output of the log package to the log file
	log.SetOutput(file)
}

type TableItem struct {
	data  string
	click func()
}

// Kinda ridiculous to use int since rows,columms and current should never be negative :/
type NaviTable struct {
	table         *tview.Table
	flex          *tview.Flex
	helpBar       *tview.TextView
	app           *tview.Application
	items         []TableItem
	current       int
	searchInput   *tview.InputField
	allItems      []TableItem
	searchMode    bool
	searchText    string
	originalFocus tview.Primitive
}

// ShowHelpModal displays a help dialog with instructions and hotkeys
func ShowHelpModal(app *tview.Application, currentFlex *tview.Flex) *tview.Modal {
	// Build help text with current git repos path
	helpText := fmt.Sprintf(`
Git Opener Help

Currently using Git Repository Path: %s

Navigation:
- ↑/k: Move selection up
- ↓/j: Move selection down
- /: Search projects
- Enter: Open selected project
- Esc: Exit application or return to main menu from modals

Configuration:
To change the Git Repositories Path, set the GIT_REPOS_PATH environment variable:
  export GIT_REPOS_PATH=/path/to/your/git/repos

Example:
  export GIT_REPOS_PATH=$HOME/workspace/git
  ./git-opener

Projects are loaded from the configured git repositories path.
`, gitReposPath)

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(currentFlex, true).SetFocus(currentFlex)
			app.SetInputCapture(nil)
		})

	// Add escape key handler for help modal
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.SetRoot(currentFlex, true).SetFocus(currentFlex)
			app.SetInputCapture(nil)
			return nil
		}
		return event
	})

	return modal
}

func (nav *NaviTable) Move(event *tcell.EventKey) *tcell.EventKey {
	// If in search mode, let the input field handle the key events
	if nav.searchMode && event.Key() != tcell.KeyEsc {
		return event
	}

	switch event.Key() {
	case tcell.KeyEnter:
		if len(nav.items) > 0 {
			nav.items[nav.current].click()
		}
		return nil
	case tcell.KeyEsc:
		if nav.searchMode {
			// Exit search mode and restore original items
			nav.exitSearchMode()
			return nil
		}
		// Exit application when Escape is pressed
		nav.app.Stop()
		return nil
	case tcell.KeyDown:
		if nav.current+1 != len(nav.items) {
			nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorWhite)
			nav.current++
			nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorRed)
		}
		return nil
	case tcell.KeyUp:
		if nav.current != 0 {
			nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorWhite)
			nav.current--
			nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorRed)
		}
		return nil
	}

	if len(nav.items) > 0 {
		nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorWhite)
	}

	switch event.Rune() {
	case 'j':
		if nav.current+1 != len(nav.items) {
			nav.current++
		}
	case 'k':
		if nav.current != 0 {
			nav.current--
		}
	case '?':
		// Show help modal when ? is pressed
		helpModal := ShowHelpModal(nav.app, nav.flex)
		nav.app.SetRoot(helpModal, true).SetFocus(helpModal)
		return nil
	case '/':
		// Enable search mode
		nav.showSearchInput()
		return nil
	}

	if len(nav.items) > 0 {
		nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorRed)
	}

	return event
}

// Show search input field
func (nav *NaviTable) showSearchInput() {
	// Create search input field if it doesn't exist
	if nav.searchInput == nil {
		nav.searchInput = tview.NewInputField().
			SetLabel("Search: ").
			SetFieldWidth(0).
			SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEnter {
					// Directly execute the click function of the selected item
					if len(nav.items) > 0 && nav.current < len(nav.items) {
						// Store the click function to execute after exiting search mode
						clickFunc := nav.items[nav.current].click
						// Exit search mode first
						nav.exitSearchMode()
						// Then execute the click function
						clickFunc()
					} else {
						// Just exit search mode if no items
						nav.exitSearchMode()
					}
				} else if key == tcell.KeyEsc {
					nav.exitSearchMode()
				}
			}).
			SetChangedFunc(func(text string) {
				nav.searchText = text
				nav.filterItems(text)
			})

		// Capture specific keys on searchInput to delegate to NaviTable's navigation logic
		nav.searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyDown:
				if len(nav.items) > 0 && nav.current+1 < len(nav.items) {
					nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorWhite)
					nav.current++
					nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorRed)
				}
				return nil // handled
			case tcell.KeyUp:
				if len(nav.items) > 0 && nav.current > 0 {
					nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorWhite)
					nav.current--
					nav.table.GetCell(nav.current, 0).SetTextColor(tcell.ColorRed)
				}
				return nil // handled
			}
			// Do not consume runes; allow typing (j/k included).
			return event
		})
	}

	// Remember the original state
	nav.searchMode = true
	nav.originalFocus = nav.app.GetFocus()
	nav.searchText = ""
	nav.searchInput.SetText("")

	// Add search input to the flex layout
	nav.flex.RemoveItem(nav.helpBar)
	nav.flex.AddItem(nav.searchInput, 1, 0, true)
	nav.flex.AddItem(nav.helpBar, 1, 0, false)

	// Focus on the search input
	nav.app.SetFocus(nav.searchInput)
}

// Exit search mode and restore original view
func (nav *NaviTable) exitSearchMode() {
	if !nav.searchMode {
		return
	}

	// Remove search input and restore focus
	nav.flex.RemoveItem(nav.searchInput)
	nav.flex.RemoveItem(nav.helpBar)
	nav.flex.AddItem(nav.helpBar, 1, 0, false)

	// Reset search mode
	nav.searchMode = false
	nav.searchText = ""

	// Always restore all items when exiting search mode
	nav.restoreAllItems()

	// Reset focus to the table
	nav.app.SetFocus(nav.originalFocus)
}

// Filter items based on search text
func (nav *NaviTable) filterItems(searchText string) {
	// If this is the first search, backup all items
	if nav.allItems == nil || len(nav.allItems) == 0 {
		nav.allItems = make([]TableItem, len(nav.items))
		copy(nav.allItems, nav.items)
	}

	// Clear the table
	nav.table.Clear()
	nav.items = nil

	if searchText == "" {
		// If search text is empty, restore all items
		nav.restoreAllItems()
		return
	}

	// Filter items that contain the search text
	for _, item := range nav.allItems {
		if strings.Contains(strings.ToLower(item.data), strings.ToLower(searchText)) {
			// Add matching item to the table
			nav.items = append(nav.items, item)
			cell := tview.NewTableCell(item.data).
				SetAlign(tview.AlignCenter).
				SetBackgroundColor(tcell.ColorBlack)
			nav.table.SetCell(len(nav.items)-1, 0, cell)
		}
	}

	// Reset current selection
	nav.current = 0
	if len(nav.items) > 0 {
		nav.table.GetCell(0, 0).SetTextColor(tcell.ColorRed)
	}
}

// Restore all original items
func (nav *NaviTable) restoreAllItems() {
	if nav.allItems == nil || len(nav.allItems) == 0 {
		return
	}

	// Clear the table
	nav.table.Clear()

	// Restore all original items
	nav.items = make([]TableItem, len(nav.allItems))
	copy(nav.items, nav.allItems)

	for i, item := range nav.items {
		cell := tview.NewTableCell(item.data).
			SetAlign(tview.AlignCenter).
			SetBackgroundColor(tcell.ColorBlack)
		nav.table.SetCell(i, 0, cell)
	}

	// Reset current selection
	nav.current = 0
	if len(nav.items) > 0 {
		nav.table.GetCell(0, 0).SetTextColor(tcell.ColorRed)
	}
}

func CreateNavTable(app *tview.Application) *NaviTable {
	// Create a table with borders
	table := tview.NewTable().SetBorders(true)
	table.SetBackgroundColor(tcell.ColorBlack)

	// Create the help bar
	helpBar := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("↑/k: Up | ↓/j: Down | /: Search | Enter: Select | ?: Help | Esc: Exit")

	// Create status bar to show git repository path
	statusBar := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorYellow)

	// Set status message based on whether GIT_REPOS_PATH was set by environment variable
	statusMessage := fmt.Sprintf("Git folder: %s", gitReposPath)
	if os.Getenv("GIT_REPOS_PATH") == "" {
		statusMessage += " (default)"
	}
	statusBar.SetText(statusMessage)

	// Create the flex layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(statusBar, 1, 0, false).
		AddItem(helpBar, 1, 0, false)

	navTab := &NaviTable{
		table:         table,
		flex:          flex,
		helpBar:       helpBar,
		app:           app,
		items:         nil,
		current:       0,
		searchInput:   nil,
		allItems:      nil,
		searchMode:    false,
		searchText:    "",
		originalFocus: nil,
	}

	navTab.flex.SetInputCapture(navTab.Move)

	return navTab
}

func (nav *NaviTable) AddItem(data string, callback func()) {
	tabItem := TableItem{data, callback}
	cell := tview.NewTableCell(data).
		SetAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorBlack)
	nav.items = append(nav.items, tabItem)
	nav.table.SetCell(len(nav.items)-1, 0, cell)
	if len(nav.items)-1 == 0 {
		nav.table.GetCell(0, 0).SetTextColor(tcell.ColorRed)
	}
}

func GetProjects() []string {
	res := []string{}

	entries, err := os.ReadDir(gitReposPath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", gitReposPath, err)
		return res
	}

	for _, entry := range entries {
		if entry.IsDir() {
			res = append(res, entry.Name())
		}
	}

	return res
}

func OpenProject(name string) {
	if tmuxxer.HasSession(name) {
		log.Printf("Session %s already exists. switching...\n", name)
		tmuxxer.ChangeSession(name)
		return
	}

	projectPath := fmt.Sprintf("%s/%s", gitReposPath, name)
	log.Printf("The path is %s", projectPath)

	err := tmuxxer.CreateSession(name, true)
	if err != nil {
		log.Printf("Error executing tmux command: %s\n", err)
	} else {
		log.Println("Tmux session created successfully.")
	}

	if tmuxxer.CreateWindow(name, "terminal") != nil {
		log.Printf("Failed to create windows")
		return
	}

	if tmuxxer.SendKeys(name, "1", "cd "+projectPath) != nil || tmuxxer.SendKeys(name, "1", "vim .") != nil {
		log.Printf("Failed to send keys")
		return
	}
	if tmuxxer.SendKeys(name, "2", "cd "+projectPath) != nil || tmuxxer.SendKeys(name, "2", "clear") != nil {
		log.Printf("Failed to send keys")
		return
	}

	if tmuxxer.SelectWindow(name, "1") != nil {
		log.Printf("Failed to select window")
		return
	}

	err = tmuxxer.ChangeSession(name)
	if err != nil {
		log.Printf("Error attaching to tmux session: %s", err)
	}
}

func main() {
	// Create a new application
	app := tview.NewApplication()

	// Create navigation table
	navTable := CreateNavTable(app)

	// Load projects and add them to the table
	projects := GetProjects()
	for _, project := range projects {
		project := project // Create local variable to avoid closure issues
		navTable.AddItem(project, func() {
			log.Println("Opening session " + project)
			OpenProject(project)
			app.Stop()
		})
	}

	// Make sure to initialize allItems with the full list of projects to ensure fresh state
	navTable.allItems = make([]TableItem, len(navTable.items))
	copy(navTable.allItems, navTable.items)
	navTable.searchMode = false
	navTable.searchText = ""

	// Add exit option
	navTable.AddItem("Exit", func() {
		// Silently exit without logging
		app.Stop()
	})

	// Set up and start the application
	if err := app.SetRoot(navTable.flex, true).SetFocus(navTable.flex).Run(); err != nil {
		panic(err)
	}
}
