package main

import (
	"crypto/rand"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"testing"
)

const size = 16 * 1024

var (
	crc32k = crc32.MakeTable(crc32.Koopman)

	crc64iso  = crc64.MakeTable(crc64.ISO)
	crc64ecma = crc64.MakeTable(crc64.ECMA)
)

func randomBuf(b *testing.B) []byte {
	buf := make([]byte, size)
	n, err := rand.Read(buf)
	if n != size || err != nil {
		b.Fatalf("failed to generate random data: %v", err)
	}
	return buf
}

func BenchmarkCRC32C(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc32.Checksum(buf, crc32c)
	}
}

func BenchmarkCRC32K(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc32.Checksum(buf, crc32k)
	}
}

func BenchmarkCRC32IEEE(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc32.ChecksumIEEE(buf)
	}
}

func BenchmarkAdler32(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adler32.Checksum(buf)
	}
}

func BenchmarkCRC64ISO(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc64.Checksum(buf, crc64iso)
	}
}

func BenchmarkCRC64ECMA(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc64.Checksum(buf, crc64ecma)
	}
}

func BenchmarkFNV32(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := fnv.New32()
		h.Write(buf)
		h.Sum(nil)
	}
}

func BenchmarkFNV64(b *testing.B) {
	buf := randomBuf(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := fnv.New64a()
		h.Write(buf)
		h.Sum(nil)
	}
}
