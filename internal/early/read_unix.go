//go:build !windows

package early

import (
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// readUntilDA1 reads the terminal's replies until the DA1 answer ("\x1b[?…c") or 250ms pass.
func readUntilDA1(fd int) string {
	var b strings.Builder
	buf := make([]byte, 256)
	deadline := time.Now().Add(250 * time.Millisecond)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return b.String()
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		n, err := unix.Poll(fds, int(left.Milliseconds())+1)
		if err != nil || n == 0 {
			return b.String()
		}
		k, err := unix.Read(fd, buf)
		if err != nil || k <= 0 {
			return b.String()
		}
		b.Write(buf[:k])
		s := b.String()
		if i := strings.LastIndex(s, "\x1b[?"); i >= 0 && strings.HasSuffix(s, "c") {
			return s
		}
	}
}
