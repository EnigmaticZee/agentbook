//go:build windows

package early

// readUntilDA1 isn't supported on Windows consoles; the theme defaults to dark there.
func readUntilDA1(fd int) string { return "" }
