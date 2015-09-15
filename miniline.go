package miniline

import (
	"fmt"
	"syscall"
	"unsafe"
)

func tcgetattr(fd uintptr) (termios syscall.Termios, err error) {
	fmt.Println("tcgetattr", fd)
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, ioctlReadTermios, fd, uintptr(unsafe.Pointer(&termios)))
	return
}

func tcsetattr(fd uintptr, termios syscall.Termios) (err error) {
	fmt.Println("tcsetattr", fd, termios)
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, ioctlWriteTermios, fd, uintptr(unsafe.Pointer(&termios)))
	return
}

func WithRaw(fd uintptr, fn func() error) (err error) {
	origTermios, err := tcgetattr(fd)
	if err != nil {
		return
	}
	rawTermios := origTermios
	rawTermios.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG

	err = tcsetattr(fd, rawTermios)
	if err != nil {
		return
	}

	defer tcsetattr(fd, origTermios)

	err = fn()
	return
}

func IsTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
