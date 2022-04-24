package backend

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"math/big"
)

/*
	https://github.com/mikelodder7/commit_twin/blob/main/go/pkg/secp256k1.go
*/
type EqProof struct {
	C  *big.Int
	D  *big.Int
	D1 *big.Int
	D2 *big.Int
}

type Commitment struct {
	X *big.Int
	Y *big.Int
}

var (
	nothingUpMySleeveQ1 = []byte("Cowards die many times before their deaths; The valiant never taste of death but once")
	nothingUpMySleeveQ2 = []byte("Men at some time are masters of their fates")
	dstG1               = []byte("BLS12381G1_XMD:SHA-256_SSWU_RO_")
	dstG2               = []byte("BLS12381G2_XMD:SHA-256_SSWU_RO_")
)

func NewEqProofP256(x, r1, r2, nonce *big.Int, pk1 *ecdsa.PublicKey, pk2 *ecdsa.PublicKey) *EqProof {
	curve1 := elliptic.P256()
	curve := curve1.Params()
	q1x := pk1.X
	q1y := pk1.Y
	q2x := pk2.X
	q2y := pk2.Y
	w, _ := rand.Int(rand.Reader, curve.N)
	n1, _ := rand.Int(rand.Reader, curve.N)
	n2, _ := rand.Int(rand.Reader, curve.N)

	w1x1, w1y1 := curve.ScalarBaseMult(w.Bytes())
	w1x2, w1y2 := curve.ScalarMult(q1x, q1y, n1.Bytes())
	w1x, w1y := curve.Add(w1x1, w1y1, w1x2, w1y2)

	w2x2, w2y2 := curve.ScalarMult(q2x, q2y, n2.Bytes())
	w2x, w2y := curve.Add(w1x1, w1y1, w2x2, w2y2)

	hasher := sha512.New384()
	_, _ = hasher.Write(w1x.Bytes())
	_, _ = hasher.Write(w1y.Bytes())
	_, _ = hasher.Write(w2x.Bytes())
	_, _ = hasher.Write(w2y.Bytes())
	_, _ = hasher.Write(nonce.Bytes())

	c := new(big.Int).SetBytes(hasher.Sum(nil))
	c.Mod(c, curve.N)

	d := new(big.Int).Sub(w, new(big.Int).Mul(c, x))
	d.Mod(d, curve.N)

	d1 := new(big.Int).Sub(n1, new(big.Int).Mul(c, r1))
	d1.Mod(d1, curve.N)

	d2 := new(big.Int).Sub(n2, new(big.Int).Mul(c, r2))
	d2.Mod(d2, curve.N)

	return &EqProof{
		c, d, d1, d2,
	}
}

func (eq *EqProof) OpenP256(b, c *Commitment, nonce *big.Int, pk1 *ecdsa.PublicKey, pk2 *ecdsa.PublicKey) bool {
	curve1 := elliptic.P256()
	curve := curve1.Params()
	q1x := pk1.X
	q1y := pk1.Y
	q2x := pk2.X
	q2y := pk2.Y

	dx, dy := curve.ScalarBaseMult(eq.D.Bytes())
	lhsx1, lhsy1 := curve.ScalarMult(q1x, q1y, eq.D1.Bytes())
	lhsx2, lhsy2 := curve.ScalarMult(b.X, b.Y, eq.C.Bytes())
	lhsx1, lhsy1 = curve.Add(dx, dy, lhsx1, lhsy1)
	lhsx, lhsy := curve.Add(lhsx2, lhsy2, lhsx1, lhsy1)

	rhsx1, rhsy1 := curve.ScalarMult(q2x, q2y, eq.D2.Bytes())
	rhsx2, rhsy2 := curve.ScalarMult(c.X, c.Y, eq.C.Bytes())
	rhsx1, rhsy1 = curve.Add(dx, dy, rhsx1, rhsy1)
	rhsx, rhsy := curve.Add(rhsx2, rhsy2, rhsx1, rhsy1)

	hasher := sha512.New384()
	_, _ = hasher.Write(lhsx.Bytes())
	_, _ = hasher.Write(lhsy.Bytes())
	_, _ = hasher.Write(rhsx.Bytes())
	_, _ = hasher.Write(rhsy.Bytes())
	_, _ = hasher.Write(nonce.Bytes())

	chal := new(big.Int).SetBytes(hasher.Sum(nil))
	chal.Mod(chal, curve.N)

	return chal.Cmp(eq.C) == 0
}

func MsgToBigInt(msg []byte) *big.Int {
	curve1 := elliptic.P256()
	curve := curve1.Params()
	hashedMsg := sha512.Sum384(msg)
	hashedMsgToBigInt := new(big.Int).SetBytes(hashedMsg[:])
	qs := new(big.Int).Mod(hashedMsgToBigInt, curve.N)
	qs1 := new(big.Int).Mod(qs, curve.B)

	return qs1
}
