// aerospace-utils is a CLI for managing Aerospace workspace sizing.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mholtzscher/aerospace-utils/cmd"
)

func main() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
