// +build windows

package miniline

import(
  "bufio"
  "os"
  "fmt"
)

// ReadLine returns a line of user input (terminated by a newline or or ^D)
// read from the tty. The given prompt is printed first. If the user types ^C,
// ReadLine returns ErrInterrupted.
func ReadLine(prompt string) (line string, err error) {

    fmt.Print(prompt)
	
	in := bufio.NewReader(os.Stdin)

	line, err = in.ReadString('\n')  
	if err != nil {
		err = ErrInterrupted
	}
	return

}