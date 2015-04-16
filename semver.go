package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch int64
	Prerelease, Build   []string
}

type tokenType int

const (
	tokenNil tokenType = iota
	tokenVersionMajor
	tokenVersionMinor
	tokenVersionPatch
	tokenVersionPrerelease
	tokenVersionBuild
)

type token struct {
	t tokenType
	p int
	v string
}

func (t token) String() string {
	if t.t == tokenNil {
		return "nil"
	}

	if len(t.v) > 10 {
		return fmt.Sprintf("%.10q...", t.v)
	}

	return fmt.Sprintf("%q", t.v)
}

type scanner struct {
	data string
	pos  int
	c    chan token
	s    []statefn
}

func (s *scanner) has(n int) bool {
	return len(s.data)-s.pos >= n
}

func (s *scanner) peek() byte {
	return s.data[s.pos]
}

func (s *scanner) next() byte {
	r := s.peek()
	s.pos++
	return r
}

func (s *scanner) expect(set string) bool {
	if strings.IndexByte(set, s.peek()) != -1 {
		s.pos++

		return true
	} else {
		return false
	}
}

func (s *scanner) expectv(set string) int {
	var l int

	for {
		if !s.has(1) || !s.expect(set) {
			break
		}

		l++
	}

	return l
}

func (s *scanner) close() {
	close(s.c)
}

type statefn func(s *scanner)

func versionStateMajor(s *scanner) {
	if l := s.expectv("0123456789"); l > 0 {
		s.c <- token{tokenVersionMajor, s.pos - l, s.data[s.pos-l : s.pos]}
	}

	s.expect(".")
}

func versionStateMinor(s *scanner) {
	if l := s.expectv("0123456789"); l > 0 {
		s.c <- token{tokenVersionMinor, s.pos - l, s.data[s.pos-l : s.pos]}
	}

	s.expect(".")
}

func versionStatePatch(s *scanner) {
	if l := s.expectv("0123456789"); l > 0 {
		s.c <- token{tokenVersionPatch, s.pos - l, s.data[s.pos-l : s.pos]}
	}

	s.expect(".")
}

func versionStatePrerelease(s *scanner) {
	if s.has(1) && s.peek() == '-' {
		s.next()

		s.s = append([]statefn{versionStatePrereleaseElement}, s.s...)
	}
}

func versionStatePrereleaseElement(s *scanner) {
	if l := s.expectv("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"); l > 0 {
		s.c <- token{tokenVersionPrerelease, s.pos - l, s.data[s.pos-l : s.pos]}

		if s.has(1) && s.peek() == '.' {
			s.next()

			s.s = append([]statefn{versionStatePrereleaseElement}, s.s...)
		}
	}
}

func versionStateBuild(s *scanner) {
	if s.has(1) && s.peek() == '+' {
		s.next()

		s.s = append([]statefn{versionStateBuildElement}, s.s...)
	}
}

func versionStateBuildElement(s *scanner) {
	if l := s.expectv("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"); l > 0 {
		s.c <- token{tokenVersionBuild, s.pos - l, s.data[s.pos-l : s.pos]}

		if s.has(1) && s.peek() == '.' {
			s.next()

			s.s = append([]statefn{versionStateBuildElement}, s.s...)
		}
	}
}

func ParseVersion(ver string) (Version, error) {
	s := scanner{
		data: ver,
		pos:  0,
		c:    make(chan token, 10),
		s: []statefn{
			versionStateMajor,
			versionStateMinor,
			versionStatePatch,
			versionStatePrerelease,
			versionStateBuild,
		},
	}

	var v Version

loop:
	for {
		select {
		case t := <-s.c:
			switch t.t {
			case tokenVersionMajor:
				if d, err := strconv.ParseInt(t.v, 10, 32); err != nil {
					return v, err
				} else {
					v.Major = d
				}
			case tokenVersionMinor:
				if d, err := strconv.ParseInt(t.v, 10, 32); err != nil {
					return v, err
				} else {
					v.Minor = d
				}
			case tokenVersionPatch:
				if d, err := strconv.ParseInt(t.v, 10, 32); err != nil {
					return v, err
				} else {
					v.Patch = d
				}
			case tokenVersionPrerelease:
				v.Prerelease = append(v.Prerelease, t.v)
			case tokenVersionBuild:
				v.Build = append(v.Build, t.v)
			}
		default:
			if len(s.s) > 0 {
				var fn statefn
				fn, s.s = s.s[0], s.s[1:]
				fn(&s)
			} else {
				break loop
			}
		}
	}

	return v, nil
}
