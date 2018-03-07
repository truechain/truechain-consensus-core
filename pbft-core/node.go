package pbft

import "time"
import "math/rand"
import "fmt"

const BUFFER_SIZE = 4096 * 32

type Node struct {
	max_requests	int
	kill_flag		bool
	ecdsaKey		[]byte
	id				int
	N 				int
	view			int
	viewInUse		bool
}