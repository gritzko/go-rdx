package rdx

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

const Sha256Bytes = 32

type Sha256 [Sha256Bytes]byte
type Sha256Merkle7574 [64]Sha256

var Sha256Zero = Sha256{}

var ErrBadSha256Hex = errors.New("malformed SHA256 hex")

func (sha *Sha256) Clear() {
	copy(sha[:], Sha256Zero[:])
}

func (sha Sha256) Equal(b Sha256) bool {
	return bytes.Equal(sha[:], b[:])
}

func (sha Sha256) Bytes() []byte {
	return sha[:]
}

func (sha Sha256) String() string {
	return hex.EncodeToString(sha[:])
}

func ParseSha256(str []byte) (sha Sha256, err error) {
	if len(str) != Sha256Bytes*2 {
		err = ErrBadSha256Hex
		return
	}
	_, err = hex.Decode(sha[:], str)
	if err != nil {
		err = ErrBadSha256Hex
	}
	return
}

func (sha Sha256) IsEmpty() bool {
	return bytes.Equal(sha[:], Sha256Zero[:])
}

func Sha256Of(data []byte) (ret Sha256) {
	hash := sha256.New()
	hash.Write(data[:])
	hash.Sum(ret[:0])
	return
}

func (sha Sha256) Merkle2(b Sha256) (sum Sha256) {
	hash := sha256.New()
	hash.Write(sha[:])
	hash.Write(b[:])
	hash.Sum(sum[:0])
	return
}

var ErrOutOfRange = errors.New("out of hash tree range")

func (line *Sha256Merkle7574) Append(next Sha256) error {
	p := next
	i := 0
	for ; !line[i].IsEmpty() && i < len(line); i++ {
		p = line[i].Merkle2(p)
		line[i] = Sha256Zero
	}
	if i == len(line) {
		return ErrOutOfRange
	}
	line[i] = p
	return nil
}

func (line *Sha256Merkle7574) Sum() (sum Sha256) {
	hash := sha256.New()
	for i := 0; i < len(line); i++ {
		hash.Write(line[i][:])
	}
	b := sum[:0]
	hash.Sum(b)
	return
}
