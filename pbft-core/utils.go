package pbft;

import "fmt"
import "color"
import "io/ioutil"
import "io"
import "os"
import "strings"

const OUTPUT_THRESHOLD = 1
const BASE_PORT = 40540
const blue = color.New(color.FgBlue).SprintFunc()
const yellow = color.New(color.yellow).SprintFunc()
const red = color.New(color.FgRed).SprintFunc()

type config struct {
	N 		int
	IPList	[]string
	Ports  	[]int
}

func myPrint(t int, args ...interface{}) {
	// t: log level
	// 		0: information
	//		1: emphasis
	//		2: warning
	//		3: error

	if t >= OUTPUT_THRESHOLD {
			
		switch t
		{
		case 0:  // info
			fmt.Print("[ ]", args...)
			break
		case 1: // emphasized
			fmt.Print(blue("[.]"), args...)
			break
		case 2: // warning
			fmt.Print(yellow("[!]"), args...)
			break
		case 3:  // error
			fmt.Print(red("[x]"), args...)
		}	
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func getIPConfigs(s string) (string[], int[]) {
	// s: config file path
	myPrint(1, "Loading IP configs...\n")
	contentB, err := ioutil.ReadFile(s)
	check (err)
	content := string(contentB)
	lst := strings.Fields(content)
	ports := make([]int)
	for k, v := range lst {
		myPrint(0, k, v)
		ports.PushBack(BASE_PORT + k)  // we assign different ports to different servers
	}
	return lst, ports
}
