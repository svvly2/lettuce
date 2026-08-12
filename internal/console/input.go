package console

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mattn/go-tty"
)

var (
	OnInput     func(m string) (string, error)
	OnLongInput func(m string) (string, error)
)

func Input(m string) (string, error) {
	if OnInput != nil {
		return OnInput(m)
	}
	fmt.Print(m)

	reader := bufio.NewReader(os.Stdin)
	r, err := reader.ReadString('\n')
	return strings.TrimSpace(r), err
}

func getClipboard() (string, error) { // mac only
	cmd := exec.Command("pbpaste")

	o, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(o)), nil
}

func LongInput(m string) (string, error) {
	if OnLongInput != nil {
		return OnLongInput(m)
	}
	if runtime.GOOS != "darwin" { // just for mac because limit is 1024 characters
		return Input(m)
	}

	fmt.Print(("(Press enter when you have it copied) ") + m)

	tty, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer tty.Close()

	for {
		r, err := tty.ReadRune()
		if err != nil {
			return "", err
		}

		if r == '\r' || r == '\n' {
			break
		}
	}

	return getClipboard()
}
