package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Logs input as std log or JSON")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	// parse args
	jsonPtr := flag.Bool("json", false, "determines whether to log input as json under the 'msg' key, defaults false")
	sepPtr := flag.String("sep", " ", "separator to using to join arguments, defaults ' '")

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s <flags> <input>\n", os.Args[0])
		os.Exit(1)
	}

	input := strings.Join(flag.Args(), *sepPtr)

	var handler slog.Handler

	if *jsonPtr {
		// if json specified print input as json
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		// otherwise print normally
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	logger := slog.New(handler)
	logger.Info(input)
}
