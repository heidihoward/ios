// Package interactive provides a simple user facing terminal client for Ios.
package interactive

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/services"
	"os"
	"strings"
)

type Interative bufio.Reader

func Create(app string) *Interative {
	fmt.Printf("Starting Ios %s client in interactive mode.\n", app)
	fmt.Print(services.GetInteractiveText(app))
	return (*Interative)(bufio.NewReader(os.Stdin))

}

func (i *Interative) Next() (string, bool) {
	b := (*bufio.Reader)(i)
	fmt.Print("Enter command: ")
	text, err := b.ReadString('\n')
	if err != nil {
		glog.Fatal(err)
	}
	text = strings.Trim(text, "\n")
	glog.V(1).Info("User entered", text)
	return text, true
}

func (_ *Interative) Return(str string) {
	// , time time.Duration  "request took ", time
	fmt.Print(str + "\n")
}
