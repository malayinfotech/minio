// MinIO Cloud Storage, (C) 2021 MinIO, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bytes"
	"encoding/hex"
	"io"
	"io/ioutil"
	"testing"

	"minio/pkg/kms"
)

var encryptDecryptTests = []struct {
	Data    []byte
	Context kms.Context
}{
	{
		Data:    nil,
		Context: nil,
	},
	{
		Data:    []byte{1},
		Context: nil,
	},
	{
		Data:    []byte{1},
		Context: kms.Context{"key": "value"},
	},
	{
		Data:    make([]byte, 1<<20),
		Context: kms.Context{"key": "value", "a": "b"},
	},
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := hex.DecodeString("ddedadb867afa3f73bd33c25499a723ed7f9f51172ee7b1b679e08dc795debcc")
	if err != nil {
		t.Fatalf("Failed to decode master key: %v", err)
	}
	KMS, err := kms.New("my-key", key)
	if err != nil {
		t.Fatalf("Failed to create KMS: %v", err)
	}

	for i, test := range encryptDecryptTests {
		ciphertext, err := Encrypt(KMS, bytes.NewReader(test.Data), test.Context)
		if err != nil {
			t.Fatalf("Test %d: failed to encrypt stream: %v", i, err)
		}
		data, err := ioutil.ReadAll(ciphertext)
		if err != nil {
			t.Fatalf("Test %d: failed to encrypt stream: %v", i, err)
		}

		plaintext, err := Decrypt(KMS, bytes.NewReader(data), test.Context)
		if err != nil {
			t.Fatalf("Test %d: failed to decrypt stream: %v", i, err)
		}
		data, err = ioutil.ReadAll(plaintext)
		if err != nil {
			t.Fatalf("Test %d: failed to decrypt stream: %v", i, err)
		}

		if !bytes.Equal(data, test.Data) {
			t.Fatalf("Test %d: decrypted data does not match original data", i)
		}
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key, err := hex.DecodeString("ddedadb867afa3f73bd33c25499a723ed7f9f51172ee7b1b679e08dc795debcc")
	if err != nil {
		b.Fatalf("Failed to decode master key: %v", err)
	}
	KMS, err := kms.New("my-key", key)
	if err != nil {
		b.Fatalf("Failed to create KMS: %v", err)
	}

	benchmarkEncrypt := func(size int, b *testing.B) {
		var (
			data      = make([]byte, size)
			plaintext = bytes.NewReader(data)
			context   = kms.Context{"key": "value"}
		)
		b.SetBytes(int64(size))
		for i := 0; i < b.N; i++ {
			ciphertext, err := Encrypt(KMS, plaintext, context)
			if err != nil {
				b.Fatal(err)
			}
			if _, err = io.Copy(ioutil.Discard, ciphertext); err != nil {
				b.Fatal(err)
			}
			plaintext.Reset(data)
		}
	}
	b.Run("1KB", func(b *testing.B) { benchmarkEncrypt(1*1024, b) })
	b.Run("512KB", func(b *testing.B) { benchmarkEncrypt(512*1024, b) })
	b.Run("1MB", func(b *testing.B) { benchmarkEncrypt(1024*1024, b) })
	b.Run("10MB", func(b *testing.B) { benchmarkEncrypt(10*1024*1024, b) })
}
