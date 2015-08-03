package base58check

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"math/big"

	"github.com/prettymuchbryce/hellobitcoin/base58check/base58"
)

func Encode(prefix byte, byteData []byte) string {
	length := len(byteData) + 1
	encoded := make([]byte, length)
	encoded[0] = prefix
	copy(encoded[1:], byteData)

	//Perform SHA-256 twice
	shaHash := sha256.New()
	shaHash.Write(encoded)
	var hash []byte = shaHash.Sum(nil)

	shaHash2 := sha256.New()
	shaHash2.Write(hash)
	hash2 := shaHash2.Sum(nil)

	//First 4 bytes if this double-sha'd byte array is the checksum
	checksum := hash2[0:4]

	//Append this checksum to the input bytes
	encodedChecksum := append(encoded, checksum...)

	//base58 alone is not enough. We need to first count each of the zero bytes
	//which are at the beginning of the encodedCheckSum
	zeroBytes := 0
	for _, cksum := range encodedChecksum {
		if cksum == 0 {
			zeroBytes += 1
		} else {
			break
		}
	}

	//Convert this checksum'd version to a big Int
	bigIntEncodedChecksum := big.NewInt(0)
	bigIntEncodedChecksum.SetBytes(encodedChecksum)

	//Encode the big int checksum'd version into a Base58Checked string
	base58EncodedChecksum := string(base58.EncodeBig(nil, bigIntEncodedChecksum))

	//Now for each zero byte we counted above we need to prepend a 1 to our
	//base58 encoded string. The rational behind this is that base58 removes 0's (0x00).
	//So bitcoin demands we add leading 0s back on as 1s.
	var buffer bytes.Buffer
	for i := 0; i < zeroBytes; i++ {
		buffer.WriteString("1")
	}

	buffer.WriteString(base58EncodedChecksum)

	return buffer.String()
}

func Decode(value string) ([]byte, byte, error) {
	zeroBytes := 0
	for i := 0; i < len(value); i++ {
		if value[i] == 49 {
			zeroBytes += 1
		} else {
			break
		}
	}

	publicKeyInt, err := base58.DecodeToBig([]byte(value))
	if err != nil {
		return nil, 0, err
	}

	encodedChecksum := publicKeyInt.Bytes()

	encoded := encodedChecksum[:len(encodedChecksum)-4]
	cksum := encodedChecksum[len(encodedChecksum)-4:]

	//Perform SHA-256 twice
	shaHash := sha256.New()
	shaHash.Write(encoded)
	var hash []byte = shaHash.Sum(nil)

	shaHash2 := sha256.New()
	shaHash2.Write(hash)
	hash2 := shaHash2.Sum(nil)

	if !bytes.Equal(hash2[:4], cksum) {
		return nil, 0, errors.New("checksum not matched")
	}

	var buffer bytes.Buffer
	for i := 0; i < zeroBytes; i++ {
		buffer.WriteByte(0)
	}

	buffer.Write(encoded)

	return buffer.Bytes()[1:], buffer.Bytes()[0], nil
}
