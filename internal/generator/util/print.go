package util

import (
	"fmt"
)

// PrintDockerfile outputs path and content to stdout in a consistent format.
func PrintDockerfile(path string, content []byte) {
	fmt.Println("Saved Dockerfile to:", path)
	fmt.Println("----- Dockerfile -----")
	fmt.Println(string(content))
	fmt.Println("----- end -----")
}
