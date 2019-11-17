package output

import (
	"fmt"
	"time"
)

var verbose bool

func SetOutput(b bool) {
	verbose = b
}

func Spit(s string) {
	if verbose {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), s)
	}
}