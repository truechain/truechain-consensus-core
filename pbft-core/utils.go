/*
The MIT License (MIT)

Copyright (c) 2018 TrueChain Foundation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package pbft

import "fmt"
import "github.com/fatih/color"
import "io/ioutil"
// import "io"
// import "os"
import "strings"

func myPrint(t int, format string, args ...interface{}) {
	// t: log level
	// 		0: information
	//		1: emphasis
	//		2: warning
	//		3: error
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if t >= OUTPUT_THRESHOLD {

		switch t {
		case 0:  // info
			fmt.Printf("[ ]" + format, args...)
			break
		case 1: // emphasized
			fmt.Printf(blue("[.]") + format, args...)
			break
		case 2: // warning
			fmt.Printf(yellow("[!]") + format, args...)
			break
		case 3:  // error
			fmt.Printf(red("[x]") + format, args...)
		}
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func getIPConfigs(s string) ([]string, []int) {
	// s: config file path
	myPrint(1, "Loading IP configs...\n")
	contentB, err := ioutil.ReadFile(s)
	checkErr (err)
	content := string(contentB)
	lst := strings.Fields(content)
	ports := make([]int, 0)
	for k, v := range lst {
		myPrint(0, string(k), v)
		ports = append(ports, BASE_PORT + k)
	}
	return lst, ports
}
