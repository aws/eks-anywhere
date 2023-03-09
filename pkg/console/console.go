package console

import (
	"bufio"
	"fmt"
	"io"
)

// Confirm prompts the user with question expecting a Y/n response. If the user responds Y, it
// returns true. All other responses return false.
func Confirm(question string, out io.Writer, in io.Reader) bool {
	fmt.Fprintf(out, "%v [Y/n]: ", question)
	scanner := bufio.NewScanner(in)
	for {
		scanner.Scan()
		if scanner.Err() != nil {
			fmt.Fprintf(out, "An error occured while reading input, please try again (%v).", scanner.Err())
			continue
		}

		return scanner.Text() == "Y"
	}
}
