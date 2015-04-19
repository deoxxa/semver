package semver

import (
	"strconv"
	"strings"

	"go.bmatsuo.co/go-lexer"
)

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

const whitespace = " \t"

func stateRange(l *lexer.Lexer) lexer.StateFn {
	if l.AcceptRun(whitespace) > 0 {
		l.Emit(ItemWhitespace)
	}

	if lexer.IsEOF(l.Peek()) {
		return nil
	}

	switch {
	case l.Accept("^"):
		l.Emit(ItemCaret)
	case l.Accept("~"):
		l.Emit(ItemTilde)
	case l.AcceptString("="):
		l.Emit(ItemEQ)
	case l.AcceptString(">="):
		l.Emit(ItemGTE)
	case l.AcceptString("<="):
		l.Emit(ItemLTE)
	case l.Accept(">"):
		l.Emit(ItemGT)
	case l.Accept("<"):
		l.Emit(ItemLT)
	}

	if l.AcceptRun(whitespace) > 0 {
		l.Emit(ItemWhitespace)
	}

	if fn := lexPartialVersion(l); fn != nil {
		return fn
	}

	if l.AcceptRun(whitespace) > 0 {
		l.Emit(ItemWhitespace)
	}

	if l.AcceptString("-") {
		l.Emit(ItemDash)

		if l.AcceptRun(whitespace) > 0 {
			l.Emit(ItemWhitespace)
		}

		if fn := lexPartialVersion(l); fn != nil {
			return fn
		}
	}

	if l.AcceptRun(whitespace) > 0 {
		l.Emit(ItemWhitespace)
	}

	if l.AcceptString("||") {
		l.Emit(ItemPipe)
	}

	return stateRange
}

func lexPartialVersion(l *lexer.Lexer) lexer.StateFn {
	if l.Accept("vV") {
		l.Ignore()
	}

	if l.Accept("*xX") {
		l.Emit(ItemMajor)
	} else if l.AcceptRun("0123456789") > 0 {
		l.Emit(ItemMajor)
	} else {
		return l.Errorf("invalid major version")
	}

	if l.Accept(".") {
		l.Emit(ItemPeriod)
	} else {
		l.Emit(ItemComplete)

		return nil
	}

	if l.Accept("*xX") {
		l.Emit(ItemMinor)
	} else if l.AcceptRun("0123456789") > 0 {
		l.Emit(ItemMinor)
	} else {
		return l.Errorf("invalid minor version")
	}

	if l.Accept(".") {
		l.Emit(ItemPeriod)
	} else {
		l.Emit(ItemComplete)

		return nil
	}

	if l.Accept("*xX") {
		l.Emit(ItemPatch)
	} else if l.AcceptRun("0123456789") > 0 {
		l.Emit(ItemPatch)
	} else {
		return l.Errorf("invalid patch version")
	}

	if l.Accept("-") {
		l.Ignore()

		for {
			if l.Accept("*") {
				l.Emit(ItemComplete)

				return nil
			}

			if l.AcceptRun("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") == 0 {
				return l.Errorf("invalid prerelease component")
			} else {
				l.Emit(ItemPrerelease)
			}

			if l.Accept(".") {
				l.Emit(ItemPeriod)
			} else {
				break
			}
		}
	}

	if l.Accept("+") {
		l.Ignore()

		for {
			if l.Accept("*") {
				l.Emit(ItemComplete)

				return nil
			}

			if l.AcceptRun("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") == 0 {
				return l.Errorf("invalid prerelease component")
			} else {
				l.Emit(ItemBuild)
			}

			if l.Accept(".") {
				l.Emit(ItemPeriod)
			} else {
				break
			}
		}
	}

	l.Emit(ItemComplete)

	return nil
}

type Operator string

const (
	OperatorNone  Operator = ""
	OperatorTilde          = "~"
	OperatorCaret          = "^"
	OperatorLT             = "<"
	OperatorLTE            = "<="
	OperatorGT             = ">"
	OperatorGTE            = ">="
	OperatorEQ             = "="
)

type Comparator struct {
	Operator Operator
	Version  Version
}

func (c Comparator) String() string {
	return string(c.Operator) + c.Version.String()
}

func (c Comparator) SatisfiedBy(v Version) bool {
	switch c.Operator {
	case OperatorNone:
		fallthrough
	case OperatorEQ:
		if v.Major != c.Version.Major {
			return false
		}
		if v.Minor != c.Version.Minor {
			return false
		}
		if v.Patch != c.Version.Patch {
			return false
		}
		if len(v.Prerelease) != len(c.Version.Prerelease) {
			return false
		}
		if len(v.Build) != len(c.Version.Build) {
			return false
		}
		return true
	case OperatorGT, OperatorGTE:
		if v.Major > c.Version.Major {
			return true
		} else if v.Major < c.Version.Major {
			return false
		}

		if v.Minor > c.Version.Minor {
			return true
		} else if v.Minor < c.Version.Minor {
			return false
		}

		if v.Patch > c.Version.Patch {
			return true
		} else if v.Patch < c.Version.Patch {
			return false
		}

		for i, j := 0, min(len(v.Prerelease), len(c.Version.Prerelease)); i < j; i++ {
			if v.Prerelease[i] > c.Version.Prerelease[i] {
				return true
			} else if v.Prerelease[i] < c.Version.Prerelease[i] {
				return false
			}
		}

		for i, j := 0, min(len(v.Build), len(c.Version.Build)); i < j; i++ {
			if v.Build[i] > c.Version.Build[i] {
				return true
			} else if v.Build[i] < c.Version.Build[i] {
				return false
			}
		}

		return c.Operator == OperatorGTE
	case OperatorLT, OperatorLTE:
		if v.Major < c.Version.Major {
			return true
		} else if v.Major > c.Version.Major {
			return false
		}

		if v.Minor < c.Version.Minor {
			return true
		} else if v.Minor > c.Version.Minor {
			return false
		}

		if v.Patch < c.Version.Patch {
			return true
		} else if v.Patch > c.Version.Patch {
			return false
		}

		for i, j := 0, min(len(v.Prerelease), len(c.Version.Prerelease)); i < j; i++ {
			if v.Prerelease[i] < c.Version.Prerelease[i] {
				return true
			} else if v.Prerelease[i] > c.Version.Prerelease[i] {
				return false
			}
		}

		for i, j := 0, min(len(v.Build), len(c.Version.Build)); i < j; i++ {
			if v.Build[i] < c.Version.Build[i] {
				return true
			} else if v.Build[i] > c.Version.Build[i] {
				return false
			}
		}

		return c.Operator == OperatorLTE
	}

	return false
}

type Set []Comparator

func (s Set) String() string {
	l := make([]string, len(s))

	for i, v := range s {
		l[i] = v.String()
	}

	return strings.Join(l, " ")
}

func (s Set) SatisfiedBy(v Version) bool {
	for _, c := range s {
		if !c.SatisfiedBy(v) {
			return false
		}
	}

	return true
}

type Range []Set

func (r Range) String() string {
	l := make([]string, len(r))

	for i, v := range r {
		l[i] = v.String()
	}

	return strings.Join(l, " || ")
}

func (r Range) SatisfiedBy(v Version) bool {
	for _, s := range r {
		if s.SatisfiedBy(v) {
			return true
		}
	}

	return false
}

func ParseRange(ver string) (Range, error) {
	l := lexer.New(stateRange, ver)

	var r Range
	var s Set

	var (
		operator                     Operator
		hasMajor, hasMinor, hasPatch bool
		major, minor, patch          int64
		prerelease, build            []string
	)

	for {
		t := l.Next()

		if t.Type == lexer.ItemError {
			return nil, t.Err()
		}

		switch t.Type {
		case ItemWhitespace:
			continue
		case ItemPeriod:
			continue
		case ItemTilde:
			operator = OperatorTilde
			continue
		case ItemCaret:
			operator = OperatorCaret
			continue
		case ItemEQ:
			operator = OperatorEQ
			continue
		case ItemLT:
			operator = OperatorLT
			continue
		case ItemGT:
			operator = OperatorGT
			continue
		case ItemLTE:
			operator = OperatorLTE
			continue
		case ItemGTE:
			operator = OperatorGTE
			continue
		case ItemMajor:
			if t.Value != "*" && t.Value != "x" && t.Value != "X" {
				hasMajor = true
				major, _ = strconv.ParseInt(t.Value, 10, 64)
			}
			continue
		case ItemMinor:
			if t.Value != "*" && t.Value != "x" && t.Value != "X" {
				hasMinor = true
				minor, _ = strconv.ParseInt(t.Value, 10, 64)
			}
			continue
		case ItemPatch:
			if t.Value != "*" && t.Value != "x" && t.Value != "X" {
				hasPatch = true
				patch, _ = strconv.ParseInt(t.Value, 10, 64)
			}
			continue
		case ItemPrerelease:
			prerelease = append(prerelease, t.Value)
			continue
		case ItemBuild:
			build = append(build, t.Value)
			continue
		case ItemComplete:
			switch operator {
			case OperatorTilde:
				c1 := Comparator{
					Operator: OperatorGTE,
					Version: Version{
						Major:      major,
						Minor:      minor,
						Patch:      patch,
						Prerelease: prerelease,
						Build:      build,
					},
				}
				c2 := Comparator{
					Operator: OperatorLT,
					Version: Version{
						Major:      major,
						Minor:      minor,
						Patch:      patch,
						Prerelease: prerelease,
						Build:      build,
					},
				}

				switch {
				case hasPatch:
					c2.Version.Minor++
				case hasMinor:
					c2.Version.Minor++
				case hasMajor:
					c2.Version.Major++
				}

				s = append(s, c1, c2)
			case OperatorCaret:
				c1 := Comparator{
					Operator: OperatorGTE,
					Version: Version{
						Major:      major,
						Minor:      minor,
						Patch:      patch,
						Prerelease: prerelease,
						Build:      build,
					},
				}
				c2 := Comparator{
					Operator: OperatorLT,
					Version: Version{
						Major: major,
						Minor: minor,
						Patch: patch,
					},
				}

				switch {
				case major != 0:
					c2.Version.Major++
					c2.Version.Minor = 0
					c2.Version.Patch = 0
				case minor != 0:
					c2.Version.Minor++
					c2.Version.Patch = 0
				case patch != 0:
					c2.Version.Patch++
				}

				s = append(s, c1, c2)
			default:
				switch {
				case hasMajor && hasMinor && hasPatch:
					s = append(s, Comparator{
						Operator: operator,
						Version: Version{
							Major:      major,
							Minor:      minor,
							Patch:      patch,
							Prerelease: prerelease,
							Build:      build,
						},
					})
				case !hasMajor:
					s = append(s, Comparator{
						Operator: OperatorGTE,
					})
				default:
					c1 := Comparator{
						Operator: OperatorGTE,
						Version: Version{
							Major:      major,
							Minor:      minor,
							Patch:      patch,
							Prerelease: prerelease,
							Build:      build,
						},
					}
					c2 := Comparator{
						Operator: OperatorLT,
						Version: Version{
							Major:      major,
							Minor:      minor,
							Patch:      patch,
							Prerelease: prerelease,
							Build:      build,
						},
					}

					switch {
					case hasPatch:
						c2.Version.Minor++
					case hasMinor:
						c2.Version.Minor++
					case hasMajor:
						c2.Version.Major++
					}

					s = append(s, c1, c2)
				}
			}
		case ItemDash:
			s[len(s)-1].Operator = OperatorGTE
			operator = OperatorLTE
			continue
		case ItemPipe:
			r = append(r, s)
			s = nil
		}

		operator = OperatorNone
		hasMajor, hasMinor, hasPatch = false, false, false
		major, minor, patch = 0, 0, 0
		prerelease, build = nil, nil

		if t.Type == lexer.ItemEOF {
			r = append(r, s)

			break
		}
	}

	return r, nil
}
