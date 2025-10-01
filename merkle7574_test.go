package rdx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test7574(t *testing.T) {
	a := Sha256Of([]byte("a"))
	assert.Equal(t, "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", a.String())
	aa := a.Merkle2(a)
	assert.Equal(t, "251a262291b87cb3c93a6ed71865da1f2c090c3d0196661a8f4a705b65836f71", aa.String())
}
