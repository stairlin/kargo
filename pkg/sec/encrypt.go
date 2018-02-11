package sec

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	// CKeySize is the cipher key size - AES-256
	CKeySize = 32
	// MKeySize is the HMAC key size - HMAC-SHA-256
	MKeySize = 32
	// KeySize is the encryption key size
	KeySize = CKeySize + MKeySize
	// NonceSize is the nonce size
	NonceSize = aes.BlockSize
	// SenderSize is the size allocated to add the sender ID
	SenderSize = 4
	// MACSize MAC size
	MACSize = 32

	nullCutset = "\x00"
)

var (
	// ErrEncrypt occurs when the encryption process fails. The reason of failure
	// is concealed for security reason
	ErrEncrypt = errors.New("sec: encryption failed")
	// ErrDecrypt occurs when the decryption process fails.
	ErrDecrypt = errors.New("sec: decryption failed")
)

// GenRandBytes generates a random slice of bytes with a length of l
func GenRandBytes(l int) ([]byte, error) {
	b := make([]byte, l)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.Wrap(err, "rand error")
	}
	return b, nil
}

// GenerateKey generates a new AES-256 key.
func GenerateKey() ([]byte, error) {
	return GenRandBytes(KeySize)
}

// GenerateNonce generates a new AES-GCM nonce.
func GenerateNonce() ([]byte, error) {
	return GenRandBytes(NonceSize)
}

// Encrypt secures a message using AES-GCM.
func Encrypt(key, message []byte) ([]byte, error) {
	c, err := aes.NewCipher(key[:CKeySize])
	if err != nil {
		return nil, ErrEncrypt
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, ErrEncrypt
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return nil, ErrEncrypt
	}

	// Seal will append the output to the first argument; the usage
	// here appends the ciphertext to the nonce. The final parameter
	// is any additional data to be authenticated.
	out := gcm.Seal(nonce, nonce, message, nil)
	return out, nil
}

// Decrypt recovers a message secured using AES-GCM.
func Decrypt(key, message []byte) ([]byte, error) {
	if len(message) <= NonceSize {
		return nil, ErrDecrypt
	}

	c, err := aes.NewCipher(key[:CKeySize])
	if err != nil {
		return nil, ErrDecrypt
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, ErrDecrypt
	}

	nonce := make([]byte, NonceSize)
	copy(nonce, message)

	out, err := gcm.Open(nil, nonce, message[NonceSize:], nil)
	if err != nil {
		return nil, ErrDecrypt
	}
	return out, nil
}

// Rotator is a encryption/decryption tool that supports key rotation
type Rotator struct {
	keys          map[uint32][]byte
	defaultSender uint32
}

// NewRotator creates a new Rotator with the given keys.
// The defaultSender will be used as the default sneder ID during the
// encryption process
func NewRotator(keys map[uint32][]byte, defaultSender uint32) *Rotator {
	return &Rotator{
		keys:          keys,
		defaultSender: defaultSender,
	}
}

// Encrypt secures a message and prepends the default 4-byte sender ID to the
// message.
func (r *Rotator) Encrypt(plaintext []byte) ([]byte, error) {
	return r.EncryptWithSender(plaintext, r.defaultSender)
}

func (r *Rotator) EncryptReader(
	plainReader io.Reader, dataBlockSize int,
) (io.ReadCloser, error) {
	return r.EncryptReaderWithSender(plainReader, dataBlockSize, r.defaultSender)
}

func (r *Rotator) EncryptReaderWithSender(
	plainreader io.Reader, dataBlockSize int, sender uint32,
) (io.ReadCloser, error) {
	key, ok := r.keys[sender]
	if !ok {
		return nil, ErrEncrypt
	}

	block, err := aes.NewCipher(key[:CKeySize])
	if err != nil {
		return nil, ErrEncrypt
	}

	out, in := io.Pipe()
	go func() {
		err := func() error {
			defer in.Close()

			header := make([]byte, SenderSize)
			binary.BigEndian.PutUint32(header, sender)
			_, err := in.Write(header)
			if err != nil {
				return ErrEncrypt
			}

			// Buffer to read by block
			bout, bin := io.Pipe()
			defer bout.Close()
			go func() {
				defer bin.Close()
				buf := bufio.NewWriterSize(bin, dataBlockSize)
				_, err := io.Copy(buf, plainreader)
				if err != nil {
					fmt.Println("Encrypt err", err)
				}
				buf.Flush()
			}()

			plaintext := make([]byte, dataBlockSize)
			signedtext := make([]byte, aes.BlockSize+dataBlockSize+MACSize)
			for {
				nr, err := bout.Read(plaintext)
				if nr > 0 {
					// Cipher data
					iv := signedtext[:aes.BlockSize]
					if _, err := io.ReadFull(rand.Reader, iv); err != nil {
						return ErrEncrypt
					}
					stream := cipher.NewCTR(block, iv)
					stream.XORKeyStream(
						signedtext[aes.BlockSize:aes.BlockSize+nr],
						plaintext[:nr],
					)

					// Sign
					h := hmac.New(sha256.New, key[CKeySize:])
					h.Write(signedtext[:aes.BlockSize+nr])
					signedtext = h.Sum(signedtext[:aes.BlockSize+nr])

					// Write block to the pipe
					nw, err := in.Write(signedtext[:aes.BlockSize+nr+MACSize])
					if err != nil {
						return errors.Wrap(err, "cannot write encrypted data to pipe")
					}
					if len(signedtext) != nw {
						return io.ErrShortWrite
					}
				}
				switch err {
				case nil:
				case io.EOF:
					// Ignore this error
					return nil
				default:
					return err
				}
			}
		}()
		if err != nil {
			fmt.Println("Encrypt error", err)
		}
	}()
	return out, nil
}

// EncryptWithSender secures a message and prepends the given 4-byte sender ID
// to the message.
func (r *Rotator) EncryptWithSender(
	plaintext []byte, sender uint32,
) ([]byte, error) {
	key, ok := r.keys[sender]
	if !ok {
		return nil, ErrEncrypt
	}

	header := make([]byte, SenderSize)
	binary.BigEndian.PutUint32(header, sender)

	block, err := aes.NewCipher(key[:CKeySize])
	if err != nil {
		return nil, ErrEncrypt
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrEncrypt
	}

	nonce, err := GenRandBytes(NonceSize)
	if err != nil {
		return nil, ErrEncrypt
	}

	buf := append(header, nonce...)
	buf = gcm.Seal(buf, nonce, plaintext, buf[:4])
	return buf, nil
}

// Decrypt takes an incoming message and uses the sender ID to
// retrieve the appropriate key. It then attempts to recover the message
// using that key.
func (r *Rotator) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) <= NonceSize+4 {
		return nil, ErrDecrypt
	}

	sender := binary.BigEndian.Uint32(ciphertext[:4])
	key, ok := r.keys[sender]
	if !ok {
		return nil, ErrDecrypt
	}

	block, err := aes.NewCipher(key[:CKeySize])
	if err != nil {
		return nil, ErrDecrypt
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecrypt
	}

	nonce := make([]byte, NonceSize)
	copy(nonce, ciphertext[4:])

	// Decrypt the message, using the sender ID as the additional
	// data requiring authentication.
	out, err := gcm.Open(nil, nonce, ciphertext[4+NonceSize:], ciphertext[:4])
	if err != nil {
		return nil, ErrDecrypt
	}
	return out, nil
}

func (r *Rotator) DecryptReader(
	cipherreader io.Reader, dataBlockSize int,
) (io.ReadCloser, error) {

	out, in := io.Pipe()
	go func() {
		err := func() error {
			defer in.Close()

			// Retrieve sender from the cipher stream
			header := make([]byte, SenderSize)
			_, err := cipherreader.Read(header)
			if err != nil {
				return ErrDecrypt
			}

			sender := binary.BigEndian.Uint32(header)
			key, ok := r.keys[sender]
			if !ok {
				return ErrDecrypt
			}

			block, err := aes.NewCipher(key[:CKeySize])
			if err != nil {
				return ErrDecrypt
			}

			// Buffer to read by block
			bout, bin := io.Pipe()
			defer bout.Close()
			go func() {
				defer bin.Close()
				buf := bufio.NewWriterSize(bin, aes.BlockSize+dataBlockSize+MACSize)
				_, err := io.Copy(buf, cipherreader)
				if err != nil {
					fmt.Println("Decrypt err", err)
				}
				buf.Flush()
			}()

			// Then read blocks
			signedtext := make([]byte, aes.BlockSize+dataBlockSize+MACSize)
			plaintext := make([]byte, dataBlockSize)
			for {
				nr, err := bout.Read(signedtext)
				if nr > 0 {
					if nr <= (aes.BlockSize + MACSize) {
						return ErrDecrypt
					}

					// Check auth
					macStart := nr - MACSize
					tag := signedtext[macStart:nr]
					ciphertext := signedtext[:macStart]
					h := hmac.New(sha256.New, key[CKeySize:])
					h.Write(ciphertext)
					mac := h.Sum(nil)
					if !hmac.Equal(mac, tag) {
						fmt.Println("Not equal")
						return ErrDecrypt
					}

					// Decipher
					iv := ciphertext[:aes.BlockSize]
					stream := cipher.NewCTR(block, iv)
					stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

					// Write plain block to writer
					nw, err := in.Write(plaintext[:len(ciphertext)-aes.BlockSize])
					if err != nil {
						return errors.Wrap(err, "cannot write decrypted data to pipe")
					}
					if len(plaintext) != nw {
						return io.ErrShortWrite
					}
				}
				switch err {
				case nil:
				case io.EOF:
					return nil
				default:
					return err
				}
			}
		}()
		switch err {
		case nil, io.ErrShortWrite:
		default:
			fmt.Println("Decrypt error", err)
		}
	}()
	return out, nil
}
