package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/term"
)

type menuItems struct {
	parent   string
	children []string
}

type menuSelections struct {
	selectedParent int
	selectedChild  int
}

var menu = [...]menuItems{
	{"1", []string{"a", "b", "c"}},
	{"2", []string{"d", "e", "f"}},
	{"3", []string{"g", "h", "i"}},
}

var (
	quitKey  = []byte{113, 0, 0, 0}
	enterKey = []byte{13, 0, 0, 0}
	upKey    = []byte{27, 91, 65, 0}
	downKey  = []byte{27, 91, 66, 0}
	rightKey = []byte{27, 91, 67, 0}
	leftKey  = []byte{27, 91, 68, 0}
)

func captureKey(bs *[]byte) {
	*bs = make([]byte, 4)
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)
	os.Stdin.Read(*bs)
}

func printMenu(ms menuSelections) {
	redBackground := "\033[41m"
	reset := "\033[0m"
	fmt.Println("Use arrows to select items, or press 'q' to quit")
	for i, v := range menu {
		// Print the parent menu items.
		if i == ms.selectedParent && ms.selectedChild == -1 {
			fmt.Printf("%s%v%s\n", redBackground, v.parent, reset)
		} else {
			fmt.Printf("%v\n", v.parent)
		}
		// Print the children menu items.
		if i == ms.selectedParent && ms.selectedChild >= 0 {
			for ii, vv := range v.children {
				if ii == ms.selectedChild {
					fmt.Printf("\t%s%v%s\n", redBackground, vv, reset)
				} else {
					fmt.Printf("\t%v\n", vv)
				}
			}
		}
	}
}

func updateMenuSelections(bs []byte, ms *menuSelections) {
	if bytes.Equal(bs, upKey) {
		// Cursor is "above" the parent list. Do nothing.
		if ms.selectedParent == -1 {
			return
		}
		// Cursor is in the child list, but at the top. Move back to parent list.
		if ms.selectedChild >= 0 {
			ms.selectedChild = (ms.selectedChild - 1)
			return
		}
		// Default. Cursor is somewhere in the middle of the parent menu. Move up one item.
		ms.selectedParent = (ms.selectedParent - 1)
	}
	if bytes.Equal(bs, downKey) {
		// Cursor is "above" the list. Move down to the first item.
		// This check should remain at the top, since it guards against index out of rage errors.
		if ms.selectedParent == -1 {
			ms.selectedParent = 0
			return
		}
		// Cursor is at the bottom of the parent list. Do Nothing.
		if ms.selectedChild == -1 && ms.selectedParent == (len(menu)-1) {
			return
		}
		// Cursor is at the bottom of the child list. Do Nothing.
		if ms.selectedChild == (len(menu[ms.selectedParent].children) - 1) {
			return
		}
		// Cursor is located in a children menu. Move down one item.
		if ms.selectedChild >= 0 {
			ms.selectedChild = (ms.selectedChild + 1)
			return
		}
		// Default. Cursor is somewhere in the middle of the parent menu. Move down one item.
		ms.selectedParent = (ms.selectedParent + 1)
	}
	if bytes.Equal(bs, rightKey) {
		// Cursor is "above" the list. Do nothing.
		if ms.selectedParent == -1 {
			return
		}
		// Cursor is already in the children menu. Do nothing.
		if ms.selectedChild >= 0 {
			return
		}
		// Default. Cursor is in the parent menu. Move to the first item in the children menu.
		ms.selectedChild = 0
	}
	if bytes.Equal(bs, leftKey) {
		ms.selectedChild = -1
	}
}

func main() {
	// Byte slice will capture various key press events.
	bs := make([]byte, 4)
	var ms = menuSelections{-1, -1}
	// While loop until user presses 'q' to quit.
	for !bytes.Equal(bs, quitKey) {
		updateMenuSelections(bs, &ms)
		printMenu(ms)
		captureKey(&bs)
		instructionLines := 1
		parentLines := len(menu)
		childLines := 0
		if ms.selectedChild >= 0 {
			childLines = len(menu[ms.selectedParent].children)
		}
		totalLines := parentLines + childLines + instructionLines
		// Respond to enter key (command selection) without clearing menu
		if bytes.Equal(bs, enterKey) {
			cmd := exec.Command("date")
			cmdOut, err := cmd.Output()
			if err != nil {
				panic(err)
			}
			fmt.Println(string(cmdOut))
		} else {
			for i := 0; i < totalLines; i++ {
				// VT100 escape code to move the cursor up one line
				// http://www.climagic.org/mirrors/VT100_Escape_Codes.html
				fmt.Printf("\033[1A")
				// Clear the line
				fmt.Printf("\033[K")
			}
		}
	}

	fmt.Println("exiting...")
}
