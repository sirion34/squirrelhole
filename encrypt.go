package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

var (
	randomKey = generateRandomKey()
)

func generateRandomKey() []byte {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func encrypt(data []byte, encryptFileName string) ([]byte, error) {
	block, err := aes.NewCipher(randomKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	ciphertext = append(nonce, ciphertext...)

	filename := fmt.Sprintf("%s.glb", encryptFileName)
	err = ioutil.WriteFile(filename, ciphertext, 0644)

	return ciphertext, nil
}

func decrypt(inputFile string) ([]byte, error) {
	filename := fmt.Sprintf("%s.glb", inputFile)
	ciphertext, err := ioutil.ReadFile(filename)

	block, err := aes.NewCipher(randomKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
