package otel

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/rand/v2"
)

// Stream ids keep the draws of different record kinds independent.
const (
	streamGraph uint64 = 0x67726170 // "grap"
	streamTrace uint64 = 0x74726163 // "trac"
	streamLog   uint64 = 0x6c6f6773 // "logs"
	streamMet   uint64 = 0x6d657472 // "metr"
	streamID    uint64 = 0x69647300 // "ids"
)

// mix is splitmix64's finaliser: a bijective, well-distributed 64-bit hash.
func mix(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

// hash combines values into one 64-bit hash.
func hash(vs ...uint64) uint64 {
	h := uint64(0x6a09e667f3bcc908)
	for _, v := range vs {
		h = mix(h ^ v)
	}
	return h
}

// newRNG returns a PCG stream for one record: the same (seed, stream, index)
// always yields the same draws.
func newRNG(seed, stream, index uint64) *rand.Rand {
	return rand.New(rand.NewPCG(hash(seed, stream), hash(seed, stream, index)))
}

// unit maps a hash to [0, 1).
func unit(h uint64) float64 { return float64(h>>11) / (1 << 53) }

// TraceID is the 16-byte id of trace i. Never all-zero.
func TraceID(seed, i uint64) [16]byte {
	var id [16]byte
	binary.BigEndian.PutUint64(id[:8], hash(seed, streamID, 1, i)|1)
	binary.BigEndian.PutUint64(id[8:], hash(seed, streamID, 2, i))
	return id
}

// SpanID is the 8-byte id of span k of trace i. Never all-zero.
func SpanID(seed, i, k uint64) [8]byte {
	var id [8]byte
	binary.BigEndian.PutUint64(id[:], hash(seed, streamID, 3, i, k)|1)
	return id
}

// TraceIDHex is the lowercase hex form the exporter stores.
func TraceIDHex(seed, i uint64) string { id := TraceID(seed, i); return hex.EncodeToString(id[:]) }

// SpanIDHex is the lowercase hex form the exporter stores.
func SpanIDHex(seed, i, k uint64) string { id := SpanID(seed, i, k); return hex.EncodeToString(id[:]) }

// logNormal draws exp(N(mu, sigma)).
func logNormal(r *rand.Rand, mu, sigma float64) float64 {
	return math.Exp(mu + sigma*r.NormFloat64())
}

// pick returns a uniformly chosen element.
func pick[T any](r *rand.Rand, xs []T) T { return xs[r.IntN(len(xs))] }
