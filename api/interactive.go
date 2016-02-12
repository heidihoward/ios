// Interative handles terminal input and feedback
package api

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"os"
	"time"
)

type Interative bufio.Reader

func Create() *Interative {
	return (*Interative)(bufio.NewReader(os.Stdin))

}

func (i *Interative) GetUserInput() string {
	b := (*bufio.Reader)(i)
	fmt.Print("Enter command: ")
	text, err := b.ReadString('\n')
	if err != nil {
		glog.Fatal(err)
	}
	glog.Info("User entered", text)
	return text
}

func (_ *Interative) OutputToUser(str string, time time.Duration) {
	fmt.Print(str, "request took ", time)
}
