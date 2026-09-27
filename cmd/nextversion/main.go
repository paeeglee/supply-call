// Command nextversion reads git tags (one per line) from stdin and prints the
// next release tag. Usage: git tag --list | go run ./cmd/nextversion patch
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"sc2overlay/internal/release"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("uso: nextversion patch|minor|major < tags")
	}
	var tags []string
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		tags = append(tags, sc.Text())
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
	v, err := release.Next(tags, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(v)
}
