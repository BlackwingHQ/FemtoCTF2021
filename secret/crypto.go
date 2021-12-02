package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs7Unpad(blocks []byte) ([]byte, error) {
	lastblock := int(blocks[len(blocks)-1])
	if verifyPKCS7(blocks) {
		return blocks[:len(blocks)-lastblock], nil
	}
	return nil, errors.New("Bad PKCS#7 padding")
}

func verifyPKCS7(blocks []byte) bool {
	lastbyte := uint(blocks[len(blocks)-1])
	if lastbyte > uint(len(blocks)-1) || lastbyte <= 0 {
		return false
	}
	for i := uint(len(blocks)) - lastbyte; i < uint(len(blocks)); i++ {
		if uint(blocks[i]) != lastbyte {
			return false
		}
	}
	return true
}

func generateRandomBytes(n uint) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func aesEncrypt(key []byte, text []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(text))
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext, text)
	return ciphertext, nil
}

func aesDecrypt(key []byte, ciphertext []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	text := make([]byte, len(ciphertext))
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(text, ciphertext)
	return text, nil
}
