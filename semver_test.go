package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	a := assert.New(t)

	v, err := ParseVersion("1.2.3-a.b+x.y.z")
	a.NoError(err)

	a.Equal(1, v.Major)
	a.Equal(2, v.Minor)
	a.Equal(3, v.Patch)

	a.Len(v.Prerelease, 2)
	a.Equal(v.Prerelease[0], "a")
	a.Equal(v.Prerelease[1], "b")

	a.Len(v.Build, 3)
	a.Equal(v.Build[0], "x")
	a.Equal(v.Build[1], "y")
	a.Equal(v.Build[2], "z")
}
