package util

import "fmt"

func Log(message string) {
	fmt.Println("[" + Now() + "] > " + message)
}
