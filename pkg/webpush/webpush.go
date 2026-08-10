package webpush

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Subscription struct {
	Endpoint string
	P256dh   string
	Auth     string
}

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	URL   string `json:"url,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type Pusher struct {
	vapidPrivateKey *ecdsa.PrivateKey
	vapidPublicKey  string
	subject         string
}

func New(vapidPrivateKeyBase64, vapidPublicKeyBase64, subject string) (*Pusher, error) {
	privBytes, err := base64.RawURLEncoding.DecodeString(vapidPrivateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode vapid private key: %w", err)
	}

	privKey := new(ecdsa.PrivateKey)
	privKey.PublicKey.Curve = elliptic.P256()
	privKey.D = new(big.Int).SetBytes(privBytes)
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privBytes)

	return &Pusher{
		vapidPrivateKey: privKey,
		vapidPublicKey:  vapidPublicKeyBase64,
		subject:         subject,
	}, nil
}

func (p *Pusher) Send(sub Subscription, payload Payload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	encrypted, err := p.encrypt(sub, payloadJSON)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	vapidHeader, err := p.vapidAuthorization(sub.Endpoint)
	if err != nil {
		return fmt.Errorf("vapid auth: %w", err)
	}

	req, err := http.NewRequest("POST", sub.Endpoint, bytes.NewReader(encrypted))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("TTL", "86400")
	req.Header.Set("Authorization", vapidHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push service returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (p *Pusher) vapidAuthorization(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	audience := u.Scheme + "://" + u.Host

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"aud": audience,
		"exp": time.Now().Add(12 * time.Hour).Unix(),
		"sub": p.subject,
	})

	signed, err := token.SignedString(p.vapidPrivateKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("vapid t=%s, k=%s", signed, p.vapidPublicKey), nil
}

func (p *Pusher) encrypt(sub Subscription, plaintext []byte) ([]byte, error) {
	uaPublicBytes, err := base64.RawURLEncoding.DecodeString(sub.P256dh)
	if err != nil {
		return nil, fmt.Errorf("decode p256dh: %w", err)
	}
	authSecret, err := base64.RawURLEncoding.DecodeString(sub.Auth)
	if err != nil {
		return nil, fmt.Errorf("decode auth: %w", err)
	}

	curve := ecdh.P256()
	localPriv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	localPub := localPriv.PublicKey()

	uaPub, err := curve.NewPublicKey(uaPublicBytes)
	if err != nil {
		return nil, fmt.Errorf("parse ua public key: %w", err)
	}

	sharedSecret, err := localPriv.ECDH(uaPub)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	prkInfoBuf := append([]byte("WebPush: info\x00"), uaPublicBytes...)
	prkInfoBuf = append(prkInfoBuf, localPub.Bytes()...)
	prk := hkdf(authSecret, sharedSecret, prkInfoBuf, 32)

	contentEncKey := hkdf(salt, prk, []byte("Content-Encoding: aes128gcm\x00"), 16)
	nonce := hkdf(salt, prk, []byte("Content-Encoding: nonce\x00"), 12)

	block, err := aes.NewCipher(contentEncKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	padded := append(plaintext, 2)
	ciphertext := gcm.Seal(nil, nonce, padded, nil)

	rs := uint32(4096)
	header := make([]byte, 0, 16+4+1+65)
	header = append(header, salt...)
	rsBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(rsBuf, rs)
	header = append(header, rsBuf...)
	localPubBytes := localPub.Bytes()
	header = append(header, byte(len(localPubBytes)))
	header = append(header, localPubBytes...)

	return append(header, ciphertext...), nil
}

func hkdf(salt, ikm, info []byte, length int) []byte {
	mac := hmacSHA256(salt, ikm)
	return hmacSHA256(mac, append(info, 1))[:length]
}

func hmacSHA256(key, data []byte) []byte {
	h := sha256.New
	ipad := make([]byte, 64)
	opad := make([]byte, 64)

	if len(key) > 64 {
		sum := sha256.Sum256(key)
		key = sum[:]
	}
	copy(ipad, key)
	copy(opad, key)

	for i := range ipad {
		ipad[i] ^= 0x36
		opad[i] ^= 0x5c
	}

	inner := h()
	inner.Write(ipad)
	inner.Write(data)
	innerSum := inner.Sum(nil)

	outer := h()
	outer.Write(opad)
	outer.Write(innerSum)
	return outer.Sum(nil)
}

func GenerateVAPIDKeys() (privateKey, publicKey string, err error) {
	curve := ecdh.P256()
	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	privateKey = base64.RawURLEncoding.EncodeToString(priv.Bytes())
	publicKey = base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes())
	return privateKey, publicKey, nil
}

// IsGone returns true when the push service endpoint is permanently invalid.
func IsGone(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "410") || strings.Contains(s, "404")
}
