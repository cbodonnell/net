package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type Cipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
	EncryptStream(dst io.Writer, src io.Reader) error
	DecryptStream(dst io.Writer, src io.Reader) error
	EncryptRoundTrip(dst, src io.ReadWriter) error
	DecryptRoundTrip(dst, src io.ReadWriter) error
}

type AESCipher struct {
	gcm     cipher.AEAD
	bufSize uint
	// TODO: Configurable buffer size
}

type AESCipherOpts struct {
	Key        []byte
	BufferSize uint
}

func NewAESCipher(opts AESCipherOpts) (*AESCipher, error) {
	// generate a new aes cipher using our 32 byte long key
	c, err := aes.NewCipher(opts.Key)
	if err != nil {
		return nil, err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	bufSize := opts.BufferSize
	if bufSize == 0 {
		bufSize = 1024
	}

	return &AESCipher{gcm, bufSize}, nil
}

func (e *AESCipher) Encrypt(message []byte) ([]byte, error) {
	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, e.gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	encryptedMessage := e.gcm.Seal(nonce, nonce, message, nil)

	return encryptedMessage, nil
}

func (e *AESCipher) Decrypt(encryptedMessage []byte) ([]byte, error) {
	// validate the message length is at least the size of the nonce
	nonceSize := e.gcm.NonceSize()
	if len(encryptedMessage) < nonceSize {
		return nil, errors.New("message too short")
	}

	// extract the nonce from the message and use it to decrypt the message
	nonce, ciphertext := encryptedMessage[:nonceSize], encryptedMessage[nonceSize:]
	message, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (e *AESCipher) EncryptStream(dst io.Writer, src io.Reader) error {
	buf := make([]byte, e.bufSize)
	n, err := src.Read(buf)
	if err != nil {
		return err
	}
	encryptedMessage, err := e.Encrypt(buf[:n])
	if err != nil {
		return err
	}
	_, err = dst.Write(encryptedMessage)
	if err != nil {
		return err
	}
	return nil
}

func (e *AESCipher) DecryptStream(dst io.Writer, src io.Reader) error {
	buf := make([]byte, e.bufSize)
	n, err := src.Read(buf)
	if err != nil {
		return err
	}
	decryptedMessage, err := e.Decrypt(buf[:n])
	if err != nil {
		return err
	}
	_, err = dst.Write(decryptedMessage)
	if err != nil {
		return err
	}
	return nil
}

func (e *AESCipher) EncryptRoundTrip(dst, src io.ReadWriter) error {
	err := e.EncryptStream(dst, src)
	if err != nil {
		return err
	}
	return e.DecryptStream(src, dst)
}

func (e *AESCipher) DecryptRoundTrip(dst, src io.ReadWriter) error {
	err := e.DecryptStream(dst, src)
	if err != nil {
		return err
	}
	return e.EncryptStream(src, dst)
}
