package chromaprint

import (
	"encoding/base64"
	"encoding/binary"
	"github.com/pkg/errors"
)

var (
	ErrInvalidFingerprint = errors.New("invalid fingerprint")
)

// Fingerprint contains raw fingerprint data.
type Fingerprint struct {
	Version int      // version of the algorithm that generated the fingerprint
	Hashes  []uint32 // the fingerprint
}

// DecodeFingerprintString decodes base64-encoded fingerprint string into binary data.
func DecodeFingerprintString(str string) ([]byte, error) {
	if len(str) == 0 {
		return nil, ErrInvalidFingerprint
	}
	return base64.RawURLEncoding.DecodeString(str)
}

// EncodeFingerprintToString encodes binary fingerprint data to a base64-encoded string.
func EncodeFingerprintToString(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// ParseFingerprint reads binary fingerprint data and returns a parsed Fingerprint structure.
func ParseFingerprint(data []byte) (*Fingerprint, error) {
	var fp Fingerprint
	err := unpackFingerprint(data, &fp)
	if err != nil {
		return nil, err
	}
	return &fp, nil
}

// ParseFingerprintString reads base64-encoded fingerprint string and returns a parsed Fingerprint structure.
func ParseFingerprintString(str string) (*Fingerprint, error) {
	data, err := DecodeFingerprintString(str)
	if err != nil {
		return nil, errors.Wrap(err, "decoding failed")
	}
	return ParseFingerprint(data)
}

// ValidateFingerprint returns true if the input data is a valid fingerprint.
func ValidateFingerprint(data []byte) bool {
	err := unpackFingerprint(data, nil)
	return err == nil
}

// ValidateFingerprintString returns true if the input string is a valid base64-encoded fingerprint.
func ValidateFingerprintString(str string) bool {
	data, err := DecodeFingerprintString(str)
	if err != nil {
		return false
	}
	return ValidateFingerprint(data)
}

func unpackFingerprint(data []byte, fp *Fingerprint) error {
	if len(data) < 4 {
		return errors.Wrap(ErrInvalidFingerprint, "encoded fingerprint is less than 4 bytes")
	}

	header := binary.BigEndian.Uint32(data)
	offset := 4

	version := int((header >> 24) & 0xff)
	totalValues := int(header & 0xffffff)

	if totalValues == 0 {
		return errors.Wrap(ErrInvalidFingerprint, "fingerprint contains no items")
	}

	bits := unpackInt3Array(data[offset:])
	numValues := 0
	numExceptionalBits := 0
	for bi, bit := range bits {
		if bit == 0 {
			numValues += 1
			if numValues == totalValues {
				bits = bits[:bi+1]
				offset += (len(bits)*3 + 8) / 8
				break
			}
		} else if bit == 7 {
			numExceptionalBits += 1
		}
	}

	if numValues != totalValues {
		return errors.Wrap(ErrInvalidFingerprint, "not enough data to decode normal bits")
	}

	if numExceptionalBits > 0 {
		exceptionalBits := unpackInt5Array(data[offset:])
		if len(exceptionalBits) != numExceptionalBits {
			return errors.Wrap(ErrInvalidFingerprint, "not enough data to decode exceptional bits")
		}
		ei := 0
		for bi, bit := range bits {
			if bit == 7 {
				bits[bi] += exceptionalBits[ei]
				ei += 1
			}
		}
	}

	if fp != nil {
		hashes := make([]uint32, totalValues)
		hi := 0
		lastBit := int8(0)
		for _, bit := range bits {
			if bit == 0 {
				if hi > 0 {
					hashes[hi] ^= hashes[hi-1]
				}
				lastBit = 0
				hi += 1
			} else {
				lastBit += bit
				hashes[hi] |= 1 << uint(lastBit-1)
			}
		}
		fp.Version = version
		fp.Hashes = hashes
	}

	return nil
}
