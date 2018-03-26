/*
Copyright (c) 2018 TrueChain Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
		case 0: // info
			fmt.Printf("[ ]"+format, args...)
			break
		case 1: // emphasized
			fmt.Printf(blue("[.]")+format, args...)
			break
		case 2: // warning
			fmt.Printf(yellow("[!]")+format, args...)
			break
		case 3: // error
			fmt.Printf(red("[x]")+format, args...)
		}
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func GetIPConfigs(s string) ([]string, []int) {
	// s: config file path
	myPrint(1, "Loading IP configs...\n")
	contentB, err := ioutil.ReadFile(s)
	checkErr(err)
	content := string(contentB)
	lst := strings.Fields(content)
	ports := make([]int, 0)
	for k, v := range lst {
		myPrint(0, string(k), v)
		ports = append(ports, BASE_PORT+k)
	}
	return lst, ports
}
