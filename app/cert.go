package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func deterministicReader(seed string) io.Reader {
	hash := sha256.Sum256([]byte(seed))
	block, _ := aes.NewCipher(hash[:16])
	stream := cipher.NewCTR(block, hash[16:32])
	return cipher.StreamReader{S: stream, R: zeroReader{}}
}

func (a *App) CACertPEM() []byte {
	_, pem, _ := a.ensureCA()
	return pem
}

func (a *App) ensureCA() (cert *tls.Certificate, certPEM []byte, fresh bool) {
	certPath := filepath.Join(a.YuanshuDir, "ca.crt")
	keyPath := filepath.Join(a.YuanshuDir, "ca.key")

	if pEM, err := os.ReadFile(certPath); err == nil {
		if keyPEM, err := os.ReadFile(keyPath); err == nil {
			c, err := tls.X509KeyPair(pEM, keyPEM)
			if err == nil {
				return &c, pEM, false
			}
		}
		os.Remove(keyPath)
		os.Remove(certPath)
	}

	seed := machineID()
	log.Printf("[ca] generating CA from machine seed (%s...) ", seed[:8])

	reader := deterministicReader(seed)
	key, err := rsa.GenerateKey(reader, 2048)
	if err != nil {
		log.Fatalf("[ca] generate key: %v", err)
	}

	serial, _ := rand.Int(reader, big.NewInt(1<<62))
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: fmt.Sprintf("YuanshuCA-%x", serial.Bytes())},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("[ca] create cert: %v", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	os.MkdirAll(a.YuanshuDir, 0755)
	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0600)

	c, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("[ca] load new cert: %v", err)
	}

	log.Printf("[ca] generated CA from machine ID")
	return &c, certPEM, true
}
