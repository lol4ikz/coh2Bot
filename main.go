package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("hello world")
	// Set the time as a seed value
	rand.Seed(time.Now().UnixNano())

}
