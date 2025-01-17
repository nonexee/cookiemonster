package monster

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

type djangoParsedData struct {
	data             string
	timestamp        string
	signature        string
	decodedSignature []byte
	algorithm        string

	compressed bool
	parsed     bool
}

func (d *djangoParsedData) String() string {
	if !d.parsed {
		return "Unparsed data"
	}

	return fmt.Sprintf("Compressed: %t\nData: %s\nTimestamp: %s\nSignature: %s\nAlgorithm: %s\n", d.compressed, d.data, d.timestamp, d.signature, d.algorithm)
}

const (
	djangoDecoder   = "django"
	djangoMinLength = 10

	djangoSeparator = `:`
	djangoSalt      = `django.contrib.sessions.backends.signed_cookiessigner`
)

var (
	djangoAlgorithmLength = map[int]string{
		20: "sha1",
		32: "sha256",
		48: "sha384",
		64: "sha512",
	}
)

func djangoDecode(c *Cookie) bool {
	if len(c.raw) < djangoMinLength {
		return false
	}

	rawData := c.raw
	var parsedData djangoParsedData

	// If the first character is a dot, it's compressed.
	if rawData[0] == '.' {
		parsedData.compressed = true
		rawData = rawData[1:]
	}

	// Break the cookie out into the session data, timestamp, and signature,
	// in that order. Note that we assume the use of `TimestampSigner`.
	components := strings.Split(rawData, djangoSeparator)
	if len(components) != 3 {
		return false
	}

	parsedData.data = components[0]
	parsedData.timestamp = components[1]
	parsedData.signature = components[2]

	// Django encodes the signature with URL-safe base64
	// without padding, so we must use `RawURLEncoding`.
	decodedSignature, err := base64.RawURLEncoding.DecodeString(parsedData.signature)
	if err != nil {
		return false
	}

	// Determine the algorithm from the digest length, or give up if we can't
	// figure it out.
	if alg, ok := djangoAlgorithmLength[len(decodedSignature)]; ok {
		parsedData.algorithm = alg
	} else {
		return false
	}

	parsedData.decodedSignature = decodedSignature
	parsedData.parsed = true
	c.wasDecodedBy(djangoDecoder, &parsedData)

	return true
}

func djangoUnsign(c *Cookie, secret []byte) bool {
	// We need to extract `toBeSigned` to prepare what we'll be signing.
	parsedData := c.parsedDataFor(djangoDecoder).(*djangoParsedData)
	toBeSigned := parsedData.data + djangoSeparator + parsedData.timestamp

	switch parsedData.algorithm {
	case "sha1":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha1Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha1HMAC(derivedKey, []byte(toBeSigned))

		// Compare this signature to the one in the `Cookie`.
		return bytes.Compare(parsedData.decodedSignature, computedSignature) == 0
	case "sha256":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha256Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha256HMAC(derivedKey, []byte(toBeSigned))

		// Compare this signature to the one in the `Cookie`.
		return bytes.Compare(parsedData.decodedSignature, computedSignature) == 0
	case "sha384":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha384Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha384HMAC(derivedKey, []byte(toBeSigned))

		// Compare this signature to the one in the `Cookie`.
		return bytes.Compare(parsedData.decodedSignature, computedSignature) == 0
	case "sha512":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha512Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha512HMAC(derivedKey, []byte(toBeSigned))

		// Compare this signature to the one in the `Cookie`.
		return bytes.Compare(parsedData.decodedSignature, computedSignature) == 0
	default:
		panic("unknown algorithm")
	}
}

func djangoResign(c *Cookie, data string, secret []byte) string {
	// We need to extract `toBeSigned` to prepare what we'll be signing.
	parsedData := c.parsedDataFor(djangoDecoder).(*djangoParsedData)

	// We need to assemble the TBS string with new data.
	toBeSigned := base64.RawURLEncoding.EncodeToString([]byte(data)) + djangoSeparator + parsedData.timestamp

	switch parsedData.algorithm {
	case "sha1":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha1Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha1HMAC(derivedKey, []byte(toBeSigned))
		return toBeSigned + djangoSeparator + base64.RawURLEncoding.EncodeToString(computedSignature)
	case "sha256":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha256Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha256HMAC(derivedKey, []byte(toBeSigned))
		return toBeSigned + djangoSeparator + base64.RawURLEncoding.EncodeToString(computedSignature)
	case "sha384":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha384Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha384HMAC(derivedKey, []byte(toBeSigned))
		return toBeSigned + djangoSeparator + base64.RawURLEncoding.EncodeToString(computedSignature)
	case "sha512":
		// Django forces us to derive a key for HMAC-ing.
		derivedKey := sha512Digest(djangoSalt + string(secret))

		// Derive the correct signature, if this was the correct secret key.
		computedSignature := sha512HMAC(derivedKey, []byte(toBeSigned))
		return toBeSigned + djangoSeparator + base64.RawURLEncoding.EncodeToString(computedSignature)
	default:
		panic("unknown algorithm")
	}
}
