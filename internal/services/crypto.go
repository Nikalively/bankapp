package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

var ErrPGPMissingKeys = fmt.Errorf("PGP keys not found")

func EncryptPGP(plainText []byte, publicKeyPath string) ([]byte, error) {
	pubKeyFile, err := os.Open(publicKeyPath)
	if err != nil {
		return nil, err
	}
	defer pubKeyFile.Close()

	block, err := armor.Decode(pubKeyFile)
	var entities openpgp.EntityList
	if err == nil && block.Type == openpgp.PublicKeyType {
		entities, err = openpgp.ReadKeyRing(block.Body)
	} else {
		// бинарный ключ
		pubKeyFile.Seek(0, io.SeekStart)
		entities, err = openpgp.ReadKeyRing(pubKeyFile)
	}
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w, err := armor.Encode(&buf, "PGP MESSAGE", nil)
	if err != nil {
		return nil, err
	}
	encryptWriter, err := openpgp.Encrypt(w, entities, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	_, err = encryptWriter.Write(plainText)
	encryptWriter.Close()
	w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ComputeHMAC(data string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func HashCVV(cvv string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	return string(h), err
}

func CheckCVVHash(plain, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
