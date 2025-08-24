package util

import "fmt"

func Log(message string) {
	fmt.Println("[" + Now() + "] > " + message)
}

func LogRoute(route string, message string) {
	fmt.Println("[" + Now() + "] " + route + " > " + message)
}
