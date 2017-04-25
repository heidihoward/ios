// Package cli provides a command line interface for interactive with Ios, useful for testing
package cli

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/services"
	"os"
	"strings"
)

type Interative bufio.Reader

func CreateInteractiveTerminal(app string) *Interative {
	fmt.Printf("Starting Ios %s client in interactive mode.\n", app)
	fmt.Print(services.GetInteractiveText(app))
	return (*Interative)(bufio.NewReader(os.Stdin))

}

func (i *Interative) FetchTerminalInput() (string, bool, bool) {
	b := (*bufio.Reader)(i)
	for {
		fmt.Print("Enter command: ")
		text, err := b.ReadString('\n')
		if err != nil {
			glog.Fatal(err)
		}
		text = strings.Trim(text, "\n")
		text = strings.Trim(text, "\r")
		glog.V(1).Info("User entered", text)
		ok, read := services.Parse("kv-store", text)
		if ok {
			return text, read, true
		} else {
			fmt.Print("Invalid command\n")
		}
	}
}

func (_ *Interative) ReturnToTerminal(str string) {
	// , time time.Duration  "request took ", time
	fmt.Print(str + "\n")
}
