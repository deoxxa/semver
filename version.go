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

func stateVersion(l *lexer.Lexer) lexer.StateFn {
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
