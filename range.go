package semver

import (
	"strconv"
	"strings"

	"go.bmatsuo.co/go-lexer"
)

func stateRange(l *lexer.Lexer) lexer.StateFn {
	if l.AcceptRun(" ") > 0 {
		l.Ignore()
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

	if l.AcceptRun(" ") > 0 {
		l.Ignore()
	}

	if fn := lexPartialVersion(l); fn != nil {
		return fn
	}

	if l.AcceptRun(" ") > 0 {
		l.Ignore()
	}

	if l.AcceptString("-") {
		l.Emit(ItemDash)

		if l.AcceptRun(" ") > 0 {
			l.Ignore()
		}

		if fn := lexPartialVersion(l); fn != nil {
			return fn
		}
	}

	if l.AcceptRun(" ") > 0 {
		l.Ignore()
	}

	if l.AcceptString("||") {
		l.Emit(ItemPipe)
	}

	return stateRange
}

func lexPartialVersion(l *lexer.Lexer) lexer.StateFn {
	if l.Accept("*xX") {
		l.Emit(ItemComplete)

		return nil
	}

	if l.AcceptRun("0123456789") == 0 {
		return l.Errorf("invalid major version")
	} else {
		l.Emit(ItemMajor)
	}

	if !l.Accept(".") {
		l.Emit(ItemComplete)

		return nil
	} else {
		l.Ignore()
	}

	if l.Accept("*xX") {
		l.Emit(ItemComplete)

		return nil
	}

	if l.AcceptRun("0123456789") == 0 {
		return l.Errorf("invalid minor version")
	} else {
		l.Emit(ItemMinor)
	}

	if !l.Accept(".") {
		l.Emit(ItemComplete)

		return nil
	} else {
		l.Ignore()
	}

	if l.Accept("*xX") {
		l.Emit(ItemComplete)

		return nil
	}

	if l.AcceptRun("0123456789") == 0 {
		return l.Errorf("invalid patch version")
	} else {
		l.Emit(ItemPatch)
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
			if l.Accept("*") {
				l.Emit(ItemComplete)

				return nil
			}

			if l.AcceptRun("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") == 0 {
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

type Set []Comparator

func (s Set) String() string {
	l := make([]string, len(s))

	for i, v := range s {
		l[i] = v.String()
	}

	return strings.Join(l, " ")
}

type Range []Set

func (r Range) String() string {
	l := make([]string, len(r))

	for i, v := range r {
		l[i] = v.String()
	}

	return strings.Join(l, " || ")
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
			hasMajor = true
			major, _ = strconv.ParseInt(t.Value, 10, 64)
			continue
		case ItemMinor:
			hasMinor = true
			minor, _ = strconv.ParseInt(t.Value, 10, 64)
			continue
		case ItemPatch:
			hasPatch = true
			patch, _ = strconv.ParseInt(t.Value, 10, 64)
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
						Major:      major,
						Minor:      minor,
						Patch:      patch,
						Prerelease: prerelease,
						Build:      build,
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
		case ItemStar:
			if !hasMajor {
				s = append(s, Comparator{
					Operator: OperatorGTE,
				})

				continue
			}

			c1 := Comparator{
				Operator: OperatorGTE,
				Version: Version{
					Major: major,
					Minor: minor,
					Patch: patch,
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
			case hasPatch:
				c2.Version.Patch++
			case hasMinor:
				c2.Version.Minor++
			case hasMajor:
				c2.Version.Major++
			}

			s = append(s, c1, c2)
		case ItemDash:
			s[len(s)-1].Operator = OperatorGTE
			operator = OperatorLTE
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
