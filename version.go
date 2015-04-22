package semver

import (
	"fmt"
	"strconv"
	"strings"

	"go.bmatsuo.co/go-lexer"
)

type Version struct {
	Major, Minor, Patch int64
	Prerelease, Build   []string
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	if len(v.Prerelease) > 0 {
		s += "-" + strings.Join(v.Prerelease, ".")
	}

	if len(v.Build) > 0 {
		s += "+" + strings.Join(v.Build, ".")
	}

	return s
}

func compareTags(a, b string) int {
	if a == b {
		return 0
	}

	d1, err1 := strconv.Atoi(a)
	d2, err2 := strconv.Atoi(b)

	if err1 == nil && err2 == nil {
		if d1 > d2 {
			return 1
		} else if d1 < d2 {
			return -1
		}
	} else {
		if a > b {
			return 1
		} else if a < b {
			return -1
		}
	}

	return 0
}

func (v Version) EqualTo(other Version) bool {
	if v.Major != other.Major || v.Minor != other.Minor || v.Patch != other.Patch || len(v.Prerelease) != len(other.Prerelease) || len(v.Build) != len(other.Build) {
		return false
	}

	for i, j := 0, len(v.Prerelease); i < j; i++ {
		if v.Prerelease[i] != other.Prerelease[i] {
			return false
		}
	}

	for i, j := 0, len(v.Build); i < j; i++ {
		if v.Build[i] != other.Build[i] {
			return false
		}
	}

	return true
}

func (v Version) GreaterThan(other Version) bool {
	if v.Major > other.Major {
		return true
	} else if v.Major < other.Major {
		return false
	}

	if v.Minor > other.Minor {
		return true
	} else if v.Minor < other.Minor {
		return false
	}

	if v.Patch > other.Patch {
		return true
	} else if v.Patch < other.Patch {
		return false
	}

	if len(v.Prerelease) == 0 && len(other.Prerelease) > 0 {
		return true
	} else if len(v.Prerelease) > 0 && len(other.Prerelease) == 0 {
		return false
	}

	for i, j := 0, min(len(v.Prerelease), len(other.Prerelease)); i < j; i++ {
		if c := compareTags(v.Prerelease[i], other.Prerelease[i]); c > 0 {
			return true
		} else if c < 0 {
			return false
		}
	}

	if len(v.Prerelease) > len(other.Prerelease) {
		return true
	}

	for i, j := 0, min(len(v.Build), len(other.Build)); i < j; i++ {
		if c := compareTags(v.Build[i], other.Build[i]); c > 0 {
			return true
		} else if c < 0 {
			return false
		}
	}

	return false
}

func (v Version) LessThan(other Version) bool {
	if v.Major < other.Major {
		return true
	} else if v.Major > other.Major {
		return false
	}

	if v.Minor < other.Minor {
		return true
	} else if v.Minor > other.Minor {
		return false
	}

	if v.Patch < other.Patch {
		return true
	} else if v.Patch > other.Patch {
		return false
	}

	if len(v.Prerelease) > 0 && len(other.Prerelease) == 0 {
		return true
	} else if len(v.Prerelease) == 0 && len(other.Prerelease) > 0 {
		return false
	}

	for i, j := 0, min(len(v.Prerelease), len(other.Prerelease)); i < j; i++ {
		if c := compareTags(v.Prerelease[i], other.Prerelease[i]); c < 0 {
			return true
		} else if c > 0 {
			return false
		}
	}

	if len(v.Prerelease) < len(other.Prerelease) {
		return true
	}

	for i, j := 0, min(len(v.Build), len(other.Build)); i < j; i++ {
		if c := compareTags(v.Build[i], other.Build[i]); c < 0 {
			return true
		} else if c > 0 {
			return false
		}
	}

	return false
}

func stateVersion(l *lexer.Lexer) lexer.StateFn {
	if l.AcceptRun(whitespace) > 0 {
		l.Ignore()
	}

	if l.Accept("vV") {
		l.Ignore()
	}

	if l.AcceptRun(whitespace) > 0 {
		l.Ignore()
	}

	if l.AcceptRun("0123456789") == 0 {
		return l.Errorf("invalid major version")
	} else {
		l.Emit(ItemMajor)
	}

	if !l.Accept(".") {
		return l.Errorf("major version should be followed by a period")
	} else {
		l.Ignore()
	}

	if l.AcceptRun("0123456789") == 0 {
		return l.Errorf("invalid minor version")
	} else {
		l.Emit(ItemMinor)
	}

	if !l.Accept(".") {
		return l.Errorf("minor version should be followed by a period")
	} else {
		l.Ignore()
	}

	if l.AcceptRun("0123456789") == 0 {
		return l.Errorf("invalid patch version")
	} else {
		l.Emit(ItemPatch)
	}

	if l.Accept("-") {
		l.Ignore()

		for {
			if l.AcceptRun(tagchars) == 0 {
				return l.Errorf("invalid prerelease component")
			} else {
				l.Emit(ItemPrerelease)
			}

			if !l.Accept(".") {
				break
			} else {
				l.Ignore()
			}
		}
	}

	if l.Accept("+") {
		l.Ignore()

		for {
			if l.AcceptRun(tagchars) == 0 {
				return l.Errorf("invalid prerelease component")
			} else {
				l.Emit(ItemBuild)
			}

			if !l.Accept(".") {
				break
			} else {
				l.Ignore()
			}
		}
	}

	if lexer.IsEOF(l.Peek()) {
		return nil
	}

	return l.Errorf("junk data after version")
}

func ParseVersion(ver string) (Version, error) {
	var v Version

	l := lexer.New(stateVersion, ver)

	for {
		t := l.Next()

		if t.Type == lexer.ItemEOF {
			break
		}

		if t.Type == lexer.ItemError {
			return v, t.Err()
		}

		switch t.Type {
		case ItemMajor:
			if d, err := strconv.ParseInt(t.Value, 10, 32); err != nil {
				return v, err
			} else {
				v.Major = d
			}
		case ItemMinor:
			if d, err := strconv.ParseInt(t.Value, 10, 32); err != nil {
				return v, err
			} else {
				v.Minor = d
			}
		case ItemPatch:
			if d, err := strconv.ParseInt(t.Value, 10, 32); err != nil {
				return v, err
			} else {
				v.Patch = d
			}
		case ItemPrerelease:
			v.Prerelease = append(v.Prerelease, t.Value)
		case ItemBuild:
			v.Build = append(v.Build, t.Value)
		}
	}

	return v, nil
}
