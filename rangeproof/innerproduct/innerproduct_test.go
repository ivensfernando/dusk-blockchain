package innerproduct

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.dusk.network/dusk-core/dusk-go/rangeproof/pedersen"
	"gitlab.dusk.network/dusk-core/dusk-go/rangeproof/vector"
	"gitlab.dusk.network/dusk-core/dusk-go/ristretto"
)

func TestProofCreation(t *testing.T) {

	const n = 128
	P, G, H, Hpf, a, b, Q := testHelpCreate(n, t)

	proof, err := Generate(G, H, a, b, Hpf, Q)
	assert.Equal(t, nil, err)

	ok := proof.Verify(G, H, proof.L, proof.R, Hpf, Q, P, n)
	assert.Equal(t, true, ok)
}

// given an n returns P, G,H,HprimeFactors a, b, Q
func testHelpCreate(n uint32, t *testing.T) (ristretto.Point, []ristretto.Point, []ristretto.Point, []ristretto.Scalar, []ristretto.Scalar, []ristretto.Scalar, ristretto.Point) {
	a := randomScalarArr(n)
	b := randomScalarArr(n)
	c, err := vector.InnerProduct(a, b)

	assert.Equal(t, nil, err)

	var y ristretto.Scalar
	y.Rand()
	var yInv ristretto.Scalar
	yInv.Inverse(&y)

	var Q ristretto.Point
	Q.Rand()

	HprimeFactors := vector.ScalarPowers(yInv, n)
	bPrime := make([]ristretto.Scalar, n)
	copy(bPrime, b)

	for i := range bPrime {
		bPrime[i].Mul(&b[i], &HprimeFactors[i])
	}

	aPrime := make([]ristretto.Scalar, n)
	copy(aPrime, a)

	// P = aPrime * G + bPrime * H + c * Q = k1 + k2 + k3
	var k1 ristretto.Point
	var k2 ristretto.Point
	var k3 ristretto.Point

	genData := []byte("dusk.BulletProof.vec1")
	ped := pedersen.New(genData)
	ped.BaseVector.Compute(n)

	genData = append(genData, uint8(1))

	ped2 := pedersen.New(genData)
	ped2.BaseVector.Compute(n)

	H := ped2.BaseVector.Bases
	G := ped.BaseVector.Bases

	k1, err = vector.Exp(aPrime, G, int(n), 1)
	k2, err = vector.Exp(bPrime, H, int(n), 1)
	k3.ScalarMult(&Q, &c)

	var P ristretto.Point
	P.SetZero()
	P.Add(&k1, &k2)
	P.Add(&P, &k3)

	return P, G, H, HprimeFactors, a, b, Q
}

func randomScalarArr(n uint32) []ristretto.Scalar {
	res := make([]ristretto.Scalar, n)

	for i := range res {
		var rand ristretto.Scalar
		rand.Rand()
		res[i] = rand
	}
	return res
}