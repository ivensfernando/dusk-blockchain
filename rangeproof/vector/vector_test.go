package vector

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toghrulmaharramov/dusk-go/ristretto"
)

func TestVectorAdd(t *testing.T) {
	N := 64
	a := make([]ristretto.Scalar, N)
	b := make([]ristretto.Scalar, N)

	var one ristretto.Scalar
	one.SetOne()

	for i := range a {
		a[i] = one
	}
	for i := range b {
		b[i] = one
	}
	res, err := Add(a, b)
	assert.Equal(t, nil, err)

	var two ristretto.Scalar
	two.SetBigInt(big.NewInt(2))

	expected := make([]ristretto.Scalar, N)

	for i := range expected {
		expected[i] = two
	}

	assert.Equal(t, len(res), len(expected))
	for i := range res {
		ok := res[i].Equals(&expected[i])
		assert.Equal(t, true, ok)
	}
}

func TestSumPowers(t *testing.T) {

	var ten ristretto.Scalar
	ten.SetBigInt(big.NewInt(10))

	expectedValues := []int64{0, 1, 11, 111, 1111, 11111, 111111, 1111111, 11111111}

	for n, expected := range expectedValues {
		res := ScalarPowersSum(ten, uint64(n))

		assert.Equal(t, expected, res.BigInt().Int64())
	}
}

func TestSumPowersNEqualZeroOne(t *testing.T) {

	var one ristretto.Scalar
	one.SetOne()

	n := uint64(0)

	res := ScalarPowersSum(one, n)
	assert.Equal(t, int64(n), res.BigInt().Int64())

	n = 1
	res = ScalarPowersSum(one, n)
	assert.Equal(t, int64(n), res.BigInt().Int64())

}
