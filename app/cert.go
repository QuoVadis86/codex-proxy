package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (a *App) ensureCA() (*tls.Certificate, []byte) {
	certPath := filepath.Join(a.YuanshuDir, "ca.crt")
	keyPath := filepath.Join(a.YuanshuDir, "ca.key")

	if certPEM, err := os.ReadFile(certPath); err == nil {
		if keyPEM, err := os.ReadFile(keyPath); err == nil {
			cert, err := tls.X509KeyPair(certPEM, keyPEM)
			if err == nil {
				return &cert, certPEM
			}
		}
	}

	log.Printf("[ca] generating new CA key pair...")
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("[ca] generate key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "YuanshuCA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("[ca] create cert: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	os.MkdirAll(a.YuanshuDir, 0755)
	os.WriteFile(certPath, certPEM, 0644)
	os.WriteFile(keyPath, keyPEM, 0600)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("[ca] load new cert: %v", err)
	}

	log.Printf("[ca] generated new CA: %s", certPath)
	return &cert, certPEM
}

func (a *App) IsCertInstalled() bool {
	if runtime.GOOS == "darwin" {
		out, _ := exec.Command("security", "find-certificate", "-c", "YuanshuCA").Output()
		return strings.Contains(string(out), "YuanshuCA")
	} else if runtime.GOOS == "windows" {
		out, _ := exec.Command("certutil", "-store", "Root", "YuanshuCA").Output()
		return strings.Contains(string(out), "YuanshuCA")
	}
	return false
}

func (a *App) InstallCert() bool {
	if a.IsCertInstalled() {
		log.Printf("[ca] already installed, skipping")
		return true
	}

	_, caPEM := a.ensureCA()
	certPath := filepath.Join(a.YuanshuDir, "ca.crt")

	log.Printf("[ca] installing CA certificate...")
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf(
			`do shell script "security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain '%s'" with administrator privileges`,
			certPath,
		)
		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			log.Printf("[ca] install failed: %v", err)
			return false
		}
	} else if runtime.GOOS == "windows" {
		tmpFile := filepath.Join(os.TempDir(), "yuanshu-ca.crt")
		os.WriteFile(tmpFile, caPEM, 0644)
		defer os.Remove(tmpFile)
		cmd := exec.Command("certutil", "-addstore", "Root", tmpFile)
		if err := cmd.Run(); err != nil {
			log.Printf("[ca] install failed: %v", err)
			return false
		}
	}
	log.Printf("[ca] installed")
	return true
}

func (a *App) RemoveCert() {
	if !a.IsCertInstalled() {
		return
	}
	if runtime.GOOS == "darwin" {
		script := `do shell script "security delete-certificate -c 'YuanshuCA'" with administrator privileges`
		exec.Command("osascript", "-e", script).Run()
	} else if runtime.GOOS == "windows" {
		exec.Command("certutil", "-delstore", "Root", "YuanshuCA").Run()
	}
}
