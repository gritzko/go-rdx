package rdx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIter(t *testing.T) {
	rdx := []byte{}
	idA := ID{0xA, 0x1}
	val := ZipInt64(-123)
	rdx = WriteRDX(rdx, LitInteger, idA, val)
	rdx = AppendInteger(rdx, 123)
	str := make([]byte, 300)
	str[222] = 33
	rdx = AppendString(rdx, str)

	it := NewIter(rdx)
	assert.True(t, it.Read())
	assert.Nil(t, it.Error())
	assert.Equal(t, byte(LitInteger), it.Lit())
	assert.Equal(t, val, it.Value())
	assert.Equal(t, idA, it.ID())
	assert.Equal(t, byte('i'), it.Rest()[0])
	assert.Equal(t, byte('i'), it.Record()[0])
	assert.False(t, it.IsLive())

	assert.True(t, it.Read())
	assert.Nil(t, it.Error())
	assert.Equal(t, byte(LitInteger), it.Lit())
	assert.Equal(t, ID{}, it.ID())
	assert.Equal(t, ZipInt64(123), it.Value())
	assert.True(t, it.IsLive())

	assert.True(t, it.Read())
	assert.Nil(t, it.Error())
	assert.Equal(t, byte(LitString), it.Lit())
	assert.Equal(t, ID{}, it.ID())
	assert.Equal(t, 0, len(it.Rest()))
	assert.True(t, it.IsLive())
	assert.Equal(t, str, it.Value())

	assert.False(t, it.Read())
	assert.True(t, it.IsEmpty())
	assert.Equal(t, nil, it.Error())
}
