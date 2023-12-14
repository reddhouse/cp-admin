package main

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/term"
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

type menuItems struct {
	main string
	sub  []string
}

var menu = [...]menuItems{
	{"1", []string{"a", "b", "c"}},
	{"2", []string{"d", "e", "f"}},
	{"3", []string{"g", "h", "i"}},
}

type menuCoordinates struct {
	parent int
	child  int
}

func printMenu(mc menuCoordinates) {
	redBackground := "\033[41m"
	reset := "\033[0m"
	fmt.Println("Use arrows to select items, or press 'q' to quit")
	for i, v := range menu {
		if i == mc.parent && mc.child == -1 {
			fmt.Printf("%s%v%s\n", redBackground, v.main, reset)
		} else {
			fmt.Printf("%v\n", v.main)
		}

		if mc.parent == i && mc.child >= 0 {
			for j, w := range v.sub {
				if j == mc.child {
					fmt.Printf("%s%v%s\n", redBackground, w, reset)
				} else {
					fmt.Printf("%v\n", w)
				}
			}
		}
	}
}

var (
	quitKey  = []byte{113, 0, 0, 0}
	enterKey = []byte{13, 0, 0, 0}
	upKey    = []byte{27, 91, 65, 0}
	downKey  = []byte{27, 91, 66, 0}
	rightKey = []byte{27, 91, 67, 0}
	leftKey  = []byte{27, 91, 68, 0}
)

func updateMenuCoordinates(bs []byte, mc *menuCoordinates) {
	if bytes.Equal(bs, upKey) {
		// Cursor is "above" the parent list. Do nothing.
		if mc.parent == -1 {
			return
		}
		// Cursor is in the child list, but at the top. Move back to parent list.
		if mc.child >= 0 {
			mc.child = (mc.child - 1)
			return
		}
		// Default. Cursor is somewhere in the middle of the main menu. Move up one item.
		mc.parent = (mc.parent - 1)
	}
	if bytes.Equal(bs, downKey) {
		// Cursor is "above" the list. Move down to the first item.
		// This check should remain at the top, since it guards against index out of rage errors.
		if mc.parent == -1 {
			mc.parent = 0
			return
		}
		// Cursor is at the bottom of the parent list. Do Nothing.
		if mc.child == -1 && mc.parent == (len(menu)-1) {
			return
		}
		// Cursor is at the bottom of the child list. Do Nothing.
		if mc.child == (len(menu[mc.parent].sub) - 1) {
			return
		}
		// Cursor is located in a sub menu. Move down one item.
		if mc.child >= 0 {
			mc.child = (mc.child + 1)
			return
		}
		// Default. Cursor is somewhere in the middle of the main menu. Move down one item.
		mc.parent = (mc.parent + 1)
	}
	if bytes.Equal(bs, rightKey) {
		// Cursor is "above" the list. Do nothing.
		if mc.parent == -1 {
			return
		}
		// Cursor is already in the sub menu. Do nothing.
		if mc.child >= 0 {
			return
		}
		// Default. Cursor is in the main menu. Move to the first item in the sub menu.
		mc.child = 0
	}
	if bytes.Equal(bs, leftKey) {
		mc.child = -1
	}
}

func main() {
	// Byte slice will capture various key press events.
	bs := make([]byte, 4)
	var coords = menuCoordinates{-1, -1}
	// While loop until user presses 'q' to quit.
	for !bytes.Equal(bs, quitKey) {
		updateMenuCoordinates(bs, &coords)
		printMenu(coords)
		captureKey(&bs)
		instructionLines := 1
		parentLines := len(menu)
		childLines := 0
		if coords.child >= 0 {
			childLines = len(menu[coords.parent].sub)
		}
		totalLines := parentLines + childLines + instructionLines
		// Respond to enter key (command selection) without clearing menu
		if bytes.Equal(bs, enterKey) {
			fmt.Println("Enter was pressed")
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
