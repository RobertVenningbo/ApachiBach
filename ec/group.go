/*
 * Copyright 2017 XLAB d.o.o.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package ec

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"math/big"

	"github.com/xlab-si/emmy/crypto/common"
)

// TODO Insert appropriate comment with description of this struct
 type PublicKey struct {
	Curve *elliptic.Curve
	X, Y *big.Int
 }
 
 func NewPublicKey(x, y *big.Int) *PublicKey {
	 return &PublicKey{
		 X: x,
		 Y: y,
	 }
 }
 
 func (e *PublicKey) Equals(b *PublicKey) bool {
	 return e.X.Cmp(b.X) == 0 && e.Y.Cmp(b.Y) == 0
 }
 
 // PrivateKey is a wrapper around elliptic.Curve. It is a cyclic PrivateKey with generator
 // (c.Params().Gx, c.Params().Gy) and order c.Params().N (which is exposed as Q in a wrapper).
 type PrivateKey struct {
	 PublicKey
	 Q *big.Int
 }
 
 func NewPrivateKey() *PrivateKey {
	 c, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	 privateKey := PrivateKey{
		 Curve: c.Curve,
		 Q:     c.Params().N, // order of generator G
	 }
	 return &privateKey
 }
 
 // GetRandomElement returns a random element from this PrivateKey.
 func (g *PrivateKey) GetRandomElement() *PublicKey {
	 r := common.GetRandomInt(g.Q)
	 el := g.ExpBaseG(r)
	 return el
 }
 
 func GetRandomInt(max *big.Int) *big.Int {
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        log.Fatal(err)
    }
    return n
}
 // Mul computes a * b in PrivateKey. This actually means a + b as this is additive PrivateKey.
 func (g *PrivateKey) Mul(a, b *PublicKey) *PublicKey {
	 // computes (x1, y1) + (x2, y2) as this is g on elliptic curves
	 x, y := g.PublicKey.Curve.Add(a.X, a.Y, b.X, b.Y)
	 return NewPublicKey(x, y)
 }
 
 // Exp computes base^exponent in PrivateKey. This actually means exponent * base as this is
 // additive PrivateKey.
 func (g *PrivateKey) Exp(base *PublicKey, exponent *big.Int) *PublicKey {
	 // computes (x, y) * exponent
	 hx, hy := g.PublicKey.Curve.ScalarMult(base.X, base.Y, exponent.Bytes())
	 return NewPublicKey(hx, hy)
 }
 
 // Exp computes base^exponent in PrivateKey where base is the generator.
 // This actually means exponent * G as this is additive PrivateKey.
 func (g *PrivateKey) ExpBaseG(exponent *big.Int) *PublicKey {
	 // computes g ^^ exponent or better to say g * exponent as this is elliptic ((gx, gy) * exponent)
	 hx, hy := g.Curve.ScalarBaseMult(exponent.Bytes())
	 return NewPublicKey(hx, hy)
 }
 
 // Inv computes inverse of x in PrivateKey. This is done by computing x^(order-1) as:
 // x * x^(order-1) = x^order = 1. Note that this actually means x * (order-1) as this is
 // additive PrivateKey.
 func (g *PrivateKey) Inv(x *PublicKey) *PublicKey {
	 orderMinOne := new(big.Int).Sub(g.Q, big.NewInt(1))
	 inv := g.Exp(x, orderMinOne)
	 return inv
 }