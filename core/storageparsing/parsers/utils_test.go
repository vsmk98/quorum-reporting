package parsers

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_RoundUp32_0(t *testing.T) {
	res := roundUpTo32(0)
	assert.Equal(t, uint64(0), res, "wrong rounded output")
}

func Test_RoundUp32_1(t *testing.T) {
	res := roundUpTo32(1)
	assert.Equal(t, uint64(32), res, "wrong rounded output")
}

func Test_RoundUp32_32(t *testing.T) {
	res := roundUpTo32(32)
	assert.Equal(t, uint64(32), res, "wrong rounded output")
}

func Test_RoundUp32_33(t *testing.T) {
	res := roundUpTo32(33)
	assert.Equal(t, uint64(64), res, "wrong rounded output")
}

func Test_Hash_SmallBigInt(t *testing.T) {
	input := new(big.Int)
	result := hash(input)

	assert.Equal(t, "0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563", result.String(), "wrong hash output")
}

func Test_Hash_LargeInt(t *testing.T) {
	input, _ := new(big.Int).SetString("df6966c971051c3d54ec59162606531493a51404a002842f56009d7e5cf4a8ca", 16)
	result := hash(input)

	assert.Equal(t, "0xafc64d4667876823fbd3f2510daa71752dbb32dda014f138587218722b444b5a", result.String(), "wrong hash output")
}
