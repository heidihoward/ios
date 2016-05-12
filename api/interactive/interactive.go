// Interative handles terminal input and feedback
package interactive

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"os"
	"strings"
	//"time"
)

type Interative bufio.Reader

func Create() *Interative {
	return (*Interative)(bufio.NewReader(os.Stdin))

}

func (i *Interative) Next() (string, bool, bool) {
	b := (*bufio.Reader)(i)
	fmt.Print("Enter command: ")
	text, err := b.ReadString('\n')
	if err != nil {
		glog.Fatal(err)
	}
	text = strings.Trim(text, "\n")
	glog.Info("User entered", text)
	return text, true, true
}

func (_ *Interative) Return(str string) {
	// , time time.Duration  "request took ", time
	fmt.Print(str)
}
