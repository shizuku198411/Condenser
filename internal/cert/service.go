package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func NewCertManager() *CertManager {
	return &CertManager{}
}

type CertManager struct{}

func (m *CertManager) EnsureSelfSignedCert(certPath string, keyPath string, cfg CertConfig) error {
	if m.isFileExists(certPath) && m.isFileExists(keyPath) {
		return nil
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: cfg.CommonName,
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(cfg.ValidFor),
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment |
			x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              cfg.DNSNames,
		IPAddresses:           cfg.IPAddresses,
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		&template,
		&privKey.PublicKey,
		privKey,
	)
	if err != nil {
		return err
	}
	if err := m.writePem(certPath, "CERTIFICATE", derBytes, 0644); err != nil {
		return err
	}

	if err := m.writePem(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privKey), 0600); err != nil {
		return err
	}

	return nil
}

func (m *CertManager) writePem(path string, typ string, der []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{
		Type:  typ,
		Bytes: der,
	})
}

func (m *CertManager) isFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
