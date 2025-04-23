package platform

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"math"

	"github.com/vzhd1701/evertoken/internal/myerrors"
)

func AESDecrypt(ciphertext []byte, key []byte, iv []byte) (plaintext []byte) {
	block, err := aes.NewCipher(key)
	myerrors.PanicFail(err)

	if len(ciphertext) < aes.BlockSize {
		err = fmt.Errorf("ciphertext too short")
		myerrors.PanicFail(err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		err = fmt.Errorf("ciphertext has wrong size")
		myerrors.PanicFail(err)
	}

	deciphered := make([]byte, len(ciphertext))

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(deciphered, ciphertext)

	plaintext = aesUnpad(deciphered)

	return
}

func aesUnpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding >= length {
		return src
	}

	return src[:(length - unpadding)]
}

func Shannon(value []byte) (bits int) {
	frq := make(map[byte]float64)

	//get frequency of characters
	for _, i := range value {
		frq[i]++
	}

	var sum float64

	for _, v := range frq {
		f := v / float64(len(value))
		sum += f * math.Log2(f)
	}

	bits = int(math.Ceil(sum*-1)) * len(value)
	return
}
