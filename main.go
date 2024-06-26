package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/joho/godotenv"
	"golang.org/x/term"
)

type command struct {
	desc string
	cmd  func()
}

type menuItems struct {
	parent   string
	children []command
}

type menuSelections struct {
	selectedParent int
	selectedChild  int
}

var cpPrivateKey *rsa.PrivateKey
var adminAuthToken string

// Hetzner API client.
var hcloudClient *hcloud.Client

var menu = [...]menuItems{
	{
		parent: "PROVISION LOCAL",
		children: []command{
			{
				desc: "Copy Private Key to Local API Server",
				cmd:  wrappedCopyPrivateKeyLocal,
			},
		},
	},
	{
		parent: "PROVISION REMOTE",
		children: []command{
			{
				desc: "Get/Set Current Resources",
				cmd:  hetznerGetAndSetCurrentResources,
			},
			{
				desc: "Create SSH Key",
				cmd:  hetznerCreateSSHKey,
			},
			{
				desc: "Write user_data_test.yml to Disk for Debugging",
				cmd:  writeUserDataToFile,
			},
			{
				desc: "Create Server 1",
				cmd:  hetznerCreateServerOne,
			},
			{
				desc: "Delete Server 1",
				cmd:  hetznerDeleteServerOne,
			},
		},
	},
	{
		parent: "API",
		children: []command{
			{
				desc: "Signup New User",
				cmd:  wrappedSignup,
			},
			{
				desc: "Login",
				cmd:  wrappedLogin,
			},
			{
				desc: "Login Code",
				cmd:  wrappedLoginCode,
			},
			{
				desc: "Create Exim",
				cmd:  wrappedCreateExim,
			},
			{
				desc: "Get Exim Details",
				cmd:  wrappedGetEximDetails,
			},
			{
				desc: "Get (All) Exims",
				cmd:  getExims,
			},
			{
				desc: "Logout",
				cmd:  wrappedLogout,
			},
		},
	},
	{
		parent: "ADMIN",
		children: []command{
			{
				desc: "Shutdown Server",
				cmd:  shutdown,
			},
			{
				desc: "Log USER_EMAIL Bucket",
				cmd:  logUserEmailBucket,
			},
			{
				desc: "Log USER_AUTH Bucket",
				cmd:  logUserAuthBucket,
			},
			{
				desc: "Log ADMIN_EMAIL Bucket",
				cmd:  logAdminEmailBucket,
			},
			{
				desc: "Log MOD_EXIM Bucket",
				cmd:  logModEximBucket,
			},
		},
	},
	{
		parent: "E2E",
		children: []command{
			{
				desc: "Run E2E Locally",
				cmd:  runEndToEndLocal,
			},
		},
	},
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
	// Use term package to make terminal raw, making characters available to be
	// read one by one as they are typed, instead of being grouped into lines.
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
	fmt.Println()
	fmt.Println("---Use arrows or press 'q' to quit---")
	for i, v := range menu {
		// Print the parent menu items.
		if i == ms.selectedParent && ms.selectedChild == -1 {
			fmt.Printf("%s%s%v%s\n", redBackground, "\u2022", v.parent, reset)
		} else {
			fmt.Printf("%s%v\n", "\u2022", v.parent)
		}
		// Print the children menu items.
		if i == ms.selectedParent && ms.selectedChild >= 0 {
			for ii, vv := range v.children {
				if ii == ms.selectedChild {
					fmt.Printf("\t%s%s%v%s\n", redBackground, "\u25E6", vv.desc, reset)
				} else {
					fmt.Printf("\t%s%v\n", "\u25E6", vv.desc)
				}
			}
		}
	}
	fmt.Println("-------------------------------------")
}

func updateMenuSelections(bs []byte, ms *menuSelections) {
	if bytes.Equal(bs, upKey) {
		// Cursor is "above" the parent menu. Do nothing.
		if ms.selectedParent == -1 {
			return
		}
		// Cursor is at the top of the children menu. Move back to parent menu.
		if ms.selectedChild >= 0 {
			ms.selectedChild = (ms.selectedChild - 1)
			return
		}
		// Default. Cursor in the middle of the parent menu. Move up one item.
		ms.selectedParent = (ms.selectedParent - 1)
	}
	if bytes.Equal(bs, downKey) {
		// Cursor is "above" the parent menu. Move down to the first item.
		// This check should remain at the top, since it guards against index out of rage errors.
		if ms.selectedParent == -1 {
			ms.selectedParent = 0
			return
		}
		// Cursor is at the bottom of the parent menu. Do Nothing.
		if ms.selectedChild == -1 && ms.selectedParent == (len(menu)-1) {
			return
		}
		// Cursor is at the bottom of the child menu. Do Nothing.
		if ms.selectedChild == (len(menu[ms.selectedParent].children) - 1) {
			return
		}
		// Cursor is in the middle of the children menu. Move down one item.
		if ms.selectedChild >= 0 {
			ms.selectedChild = (ms.selectedChild + 1)
			return
		}
		// Default. Cursor is in the middle of the parent menu. Move down one item.
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
		// Default. Cursor is in the parent menu. Move to the first item in the corresponding children menu.
		ms.selectedChild = 0
	}
	if bytes.Equal(bs, leftKey) {
		// Close the children menu and return to the parent menu.
		ms.selectedChild = -1
	}
}

func loadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("[err][admin] loading .env file: %v [%s]\n", err, cts())
		os.Exit(1)
	}
}

func setHetznerCloudClient() {
	token := os.Getenv("HETZNER_API_TOKEN")
	hcloudClient = hcloud.NewClient(hcloud.WithToken(token))
}

func runSelectedCommands() {
	// Capture various key press events in 4-byte slice.
	bs := make([]byte, 4)
	var ms = menuSelections{-1, -1}
	// Loop until user presses 'q' to quit.
	for !bytes.Equal(bs, quitKey) {
		updateMenuSelections(bs, &ms)
		printMenu(ms)
		captureKey(&bs)
		parentLines := len(menu)
		childLines := 0
		dividerLines := 3
		if ms.selectedChild >= 0 {
			childLines = len(menu[ms.selectedParent].children)
		}
		totalLines := parentLines + childLines + dividerLines
		// Respond to enter key (command selection) without clearing menu.
		if bytes.Equal(bs, enterKey) && ms.selectedChild >= 0 {
			selectedCommand := menu[ms.selectedParent].children[ms.selectedChild]
			// Run synchronous command and block until completion.
			selectedCommand.cmd()
		} else {
			// Some other (non-enter) key was pressed.
			// Clear the space that the current menu is occupying so the next
			// loop iteration will appear to update/expand the menu "in place".
			for i := 0; i < totalLines; i++ {
				// Move the cursor up one line (see VT100 escape codes).
				fmt.Printf("\033[1A")
				// Clear the line.
				fmt.Printf("\033[K")
			}
		}
	}
}

func main() {
	loadEnvVariables()

	// Generate private key file if it doesn't already exist.
	_, err := os.Stat("cp.pem")
	if os.IsNotExist(err) {
		generatePrivateKeyFile()
	}

	setPrivateKey()
	setAdminAuthToken()
	setHetznerCloudClient()
	runSelectedCommands()
	fmt.Printf("[admin] exiting... [%s]\n", cts())
}
