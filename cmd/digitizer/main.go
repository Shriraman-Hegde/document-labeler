package main

import (
	"os"

	"github.com/Shriraman-Hegde/document-labeler/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
