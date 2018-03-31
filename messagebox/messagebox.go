package messagebox

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
	"github.com/schollz/messagebox/keypair"
	"golang.org/x/crypto/nacl/secretbox"
)

type Message struct {
	Recipients []string `json:"recipients"`
	Message    string   `json:"message"`
}

// Open will open a message
func (m *Message) Open(recipient keypair.KeyPair, senders []string) (originalSender string, decrypted []byte, err error) {
	var randomEncryption []byte
	var decodedRecipient []byte
	for _, rec := range m.Recipients {
		for _, sender := range senders {
			originalSender = sender
			decodedRecipient, err = base64.StdEncoding.DecodeString(rec)
			randomEncryption, err = recipient.Decrypt(decodedRecipient, sender)
			if err == nil {
				break
			}
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}

	encrypted, err := base64.StdEncoding.DecodeString(m.Message)
	if err != nil {
		return
	}

	var randomEncryption32 [32]byte
	copy(randomEncryption32[:], randomEncryption[:32])
	decrypted, err = decrypt(encrypted, randomEncryption32)
	return
}

// New will generate a new message
func New(sender keypair.KeyPair, recipients []string, msg []byte) (m Message, err error) {
	encrypted, secretKey, err := encryptWithRandomSecret(msg)
	if err != nil {
		return
	}

	m = Message{
		Message:    base64.StdEncoding.EncodeToString(encrypted),
		Recipients: make([]string, len(recipients)),
	}
	for i, rec := range recipients {
		encrypted, err2 := sender.Encrypt(secretKey[:], rec)
		if err != nil {
			err = errors.Wrap(err2, rec)
			return
		}
		m.Recipients[i] = base64.StdEncoding.EncodeToString(encrypted)
	}
	return
}

func encryptWithRandomSecret(msg []byte) (encrypted []byte, secretKey [32]byte, err error) {
	if _, err = io.ReadFull(crypto_rand.Reader, secretKey[:]); err != nil {
		return
	}

	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err = io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return
	}

	// This encrypts msg and appends the result to the nonce.
	encrypted = secretbox.Seal(nonce[:], msg, &nonce, &secretKey)
	return
}

func decrypt(encrypted []byte, secretKey [32]byte) (decrypted []byte, err error) {
	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok {
		err = errors.New("decryption failed")
	}
	return
}
