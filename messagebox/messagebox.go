package messagebox

import (
	crypto_rand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/schollz/messagebox/keypair"
	"golang.org/x/crypto/nacl/secretbox"
)

// Message contains the sender, recipients and encrypted message body.
type Message struct {
	// Sender is the public key of the sender
	Sender string `json:"sender"`
	// Recipients are a list of the public keys of the recipients
	Recipients []string `json:"recipients"`
	// Message is the payload that is encrypted by a random
	// string which is encoded for each of the recipients
	Message string `json:"message"`
	// Hash is a SHA256 hash of the message+sender+salt
	Hash string `json:"hash"`
}

func (m *Message) String() string {
	out, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (m *Message) HashMessage() string {
	out, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// IsSameWorld checks to make sure that the message is from the same domain.
func (m *Message) IsSameWorld(world keypair.KeyPair) bool {
	if world.Public == "" {
		return false
	}
	decodedSender, err := base64.StdEncoding.DecodeString(m.Sender)
	if err != nil {
		return false
	}
	_, err = world.Decrypt(decodedSender, world.Public)
	if err != nil {
		return false
	}
	return true
}

// Open will open a message by trying each of my keys and
// will return the key that opened the message and the
// descrypted contents
func (m *Message) Open(world keypair.KeyPair, mykeys []keypair.KeyPair) (keyOpened keypair.KeyPair, decrypted []byte, err error) {
	encrypted, err := base64.StdEncoding.DecodeString(m.Message)
	if err != nil {
		err = errors.Wrap(err, "message is not decodable")
		return
	}

	decodedSender, err := base64.StdEncoding.DecodeString(m.Sender)
	if err != nil {
		return
	}
	decryptedSender, err := world.Decrypt(decodedSender, world.Public)
	if err != nil {
		err = errors.Wrap(err, "world cannot decrypt sender")
		return
	}
	senderPublicKey := string(decryptedSender)

	var randomEncryption []byte
	for _, recipient := range m.Recipients {
		var decodedRecipient []byte
		decodedRecipient, err = base64.StdEncoding.DecodeString(recipient)
		for _, key := range mykeys {
			if err != nil {
				err = errors.Wrap(err, "malformed recipient")
				return
			}
			randomEncryption, err = key.Decrypt(decodedRecipient, senderPublicKey)
			if err == nil {
				keyOpened = key
				break
			}
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		err = errors.Wrap(err, "could not find valid recipient")
		return
	}
	var randomEncryption32 [32]byte
	copy(randomEncryption32[:], randomEncryption[:32])
	decrypted, err = decrypt(encrypted, randomEncryption32)
	return
}

// New will generate a new message
func New(world keypair.KeyPair, sender keypair.KeyPair, recipients []string, msg []byte) (m Message, err error) {
	encrypted, secretKey, err := encryptWithRandomSecret(msg)
	if err != nil {
		return
	}
	h := sha256.New()
	h.Write([]byte("messagebox101"))
	h.Write(msg)
	h.Write([]byte(sender.Public))

	encryptedSender, err := world.Encrypt([]byte(sender.Public), world.Public)
	if err != nil {
		return
	}

	m = Message{
		Sender:     base64.StdEncoding.EncodeToString(encryptedSender),
		Message:    base64.StdEncoding.EncodeToString(encrypted),
		Recipients: make([]string, len(recipients)),
		Hash:       fmt.Sprintf("sha256/%x", h.Sum(nil)),
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
