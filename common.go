package semver

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

const (
	whitespace = " \t"
	tagchars   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"
)
