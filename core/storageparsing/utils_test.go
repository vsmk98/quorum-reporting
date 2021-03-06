package storageparsing

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/types"
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
	result := hash(types.NewHash(""))

	assert.Equal(t, "0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563", result.String(), "wrong hash output")
}

func Test_Hash_LargeInt(t *testing.T) {
	result := hash(types.NewHash("df6966c971051c3d54ec59162606531493a51404a002842f56009d7e5cf4a8ca"))

	assert.Equal(t, "0xafc64d4667876823fbd3f2510daa71752dbb32dda014f138587218722b444b5a", result.String(), "wrong hash output")
}

func Test_ExtractFromSingleStorage_NoOffset(t *testing.T) {
	in, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000001023456789")
	expected, _ := hex.DecodeString("00000000001023456789")

	out := ExtractFromSingleStorage(0, 10, in)
	assert.Equal(t, expected, out)
}

func Test_ExtractFromSingleStorage_WithOffset(t *testing.T) {
	in, _ := hex.DecodeString("0000000000000000000000000000000000000000000010234567890000000000")
	expected, _ := hex.DecodeString("00000000001023456789")

	out := ExtractFromSingleStorage(5, 10, in)
	assert.Equal(t, expected, out)
}
