package miniline

// InterruptedError represents the user having exited the prompt with ^C
type InterruptedError struct{}

// Error is just the string "Interrupted"
func (e InterruptedError) Error() string {
	return "Interrupted"
}

// ErrInterrupted is a singleton InterruptedError
var ErrInterrupted error = InterruptedError{}

