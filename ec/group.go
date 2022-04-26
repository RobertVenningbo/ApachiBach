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
)

func NewPublicKey(c elliptic.Curve, a, b *big.Int) *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: c,
		X:     a,
		Y:     b,
	}
}

func Equals(e *ecdsa.PublicKey, b *ecdsa.PublicKey) bool {
	return e.X.Cmp(b.X) == 0 && e.Y.Cmp(b.Y) == 0
}

func NewPrivateKey() *ecdsa.PrivateKey {
	c, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privateKey := ecdsa.PrivateKey{
		PublicKey: c.PublicKey,
		D:         c.Params().N, // order of generator G
	}
	return &privateKey
}

// GetRandomElement returns a random element from this PrivateKey.
func GetRandomElement(g *ecdsa.PrivateKey) *ecdsa.PublicKey {
	r := GetRandomInt(g.D)
	el := ExpBaseG(g, r)
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
func Mul(g *ecdsa.PrivateKey, a *ecdsa.PublicKey, b *ecdsa.PublicKey) *ecdsa.PublicKey {
	// computes (x1, y1) + (x2, y2) as this is g on elliptic curves
	x, y := g.PublicKey.Curve.Add(a.X, a.Y, b.X, b.Y)
	return NewPublicKey(g.PublicKey.Curve, x, y)
}

// Exp computes base^exponent in PrivateKey. This actually means exponent * base as this is
// additive PrivateKey.
func Exp(g *ecdsa.PrivateKey, base *ecdsa.PublicKey, exponent *big.Int) *ecdsa.PublicKey {
	// computes (x, y) * exponent
	hx, hy := g.PublicKey.Curve.ScalarMult(base.X, base.Y, exponent.Bytes())
	return NewPublicKey(g.PublicKey.Curve, hx, hy)
}

// Exp computes base^exponent in PrivateKey where base is the generator.
// This actually means exponent * G as this is additive PrivateKey.
func ExpBaseG(g *ecdsa.PrivateKey, exponent *big.Int) *ecdsa.PublicKey {
	// computes g ^^ exponent or better to say g * exponent as this is elliptic ((gx, gy) * exponent)
	hx, hy := g.Curve.ScalarBaseMult(exponent.Bytes())
	return NewPublicKey(g.PublicKey.Curve, hx, hy)
}

// Inv computes inverse of x in PrivateKey. This is done by computing x^(order-1) as:
// x * x^(order-1) = x^order = 1. Note that this actually means x * (order-1) as this is
// additive PrivateKey.
func Inv(g *ecdsa.PrivateKey, x *ecdsa.PublicKey) *ecdsa.PublicKey {
	orderMinOne := new(big.Int).Sub(g.D, big.NewInt(1))
	inv := Exp(g, x, orderMinOne)
	return inv
}
