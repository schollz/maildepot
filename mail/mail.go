package mail

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/schollz/maildepot/keypair"
	"golang.org/x/crypto/nacl/secretbox"
)

// Message contains the sender, recipients and encrypted message body.
type Message struct {
	// Sender is the public key of the sender encrypted by message key
	Sender string `json:"s"`
	// Recipients is a list where each is a message key encrypted by the world for the public key of the intended recipient
	Recipients []string `json:"r"`
	// Message is the payload encrypted by the message key
	Message string `json:"m"`
}

type OpenMessage struct {
	// Sender is the public key of sender
	Sender string `json:"s"`
	// Recipients is a list public key of recipients
	Recipients []keypair.KeyPair `json:"r"`
	// Message is the payload
	MessageBytes []byte `json:"m"`
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
func (m Message) Open(world keypair.KeyPair, mykeys []keypair.KeyPair) (openMsg OpenMessage, err error) {
	openMsg = OpenMessage{}

	// check if message is decodable
	encryptedMessage, err := base64.StdEncoding.DecodeString(m.Message)
	if err != nil {
		err = errors.Wrap(err, "message is not decodable")
		return
	}
	// check if sender is decodable
	encryptedSender, err := base64.StdEncoding.DecodeString(m.Sender)
	if err != nil {
		err = errors.Wrap(err, "sender is not decodable")
		return
	}

	var secretKey []byte
	for _, recipient := range m.Recipients {
		var decodedRecipient []byte
		decodedRecipient, err = base64.StdEncoding.DecodeString(recipient)
		for _, key := range mykeys {
			if err != nil {
				err = errors.Wrap(err, "malformed recipient")
				return
			}
			secretKey, err = key.Decrypt(decodedRecipient, world.Public)
			if err == nil {
				openMsg.Recipients = append(openMsg.Recipients, key)
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("could not find valid recipient")
		return
	}

	var secretKey32 [32]byte
	copy(secretKey32[:], secretKey[:32])
	openMsg.MessageBytes, err = decrypt(encryptedMessage, secretKey32)
	if err != nil {
		err = errors.Wrap(err, "could not decrypt message with key")
		return
	}

	senderBytes, err := decrypt(encryptedSender, secretKey32)
	if err != nil {
		err = errors.Wrap(err, "could not decrypt sender with key")
		return
	}
	openMsg.Sender = string(senderBytes)

	return
}

// New will generate a new message
func New(world keypair.KeyPair, sender keypair.KeyPair, recipients []string, msg []byte) (m Message, err error) {
	// generate new secretKey for the message key
	encrypted, secretKey, err := encryptWithRandomSecret(msg)
	if err != nil {
		return
	}

	// encrypt the sender with the message key
	encryptedSender, err := encryptWithSecret([]byte(sender.Public), secretKey)
	if err != nil {
		return
	}

	m = Message{
		Sender:     base64.StdEncoding.EncodeToString(encryptedSender),
		Message:    base64.StdEncoding.EncodeToString(encrypted),
		Recipients: make([]string, len(recipients)),
	}

	// encrypt each recipient with the message key
	for i, recipientPublicKey := range recipients {
		encrypted, err2 := world.Encrypt(secretKey[:], recipientPublicKey)
		if err != nil {
			err = errors.Wrap(err2, recipientPublicKey)
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

	encrypted, err = encryptWithSecret(msg, secretKey)
	return
}

func encryptWithSecret(msg []byte, secretKey [32]byte) (encrypted []byte, err error) {
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
