package api

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/likexian/selfca"
	"github.com/pkg/errors"
)

func generateCert() (*tls.Certificate, *x509.CertPool, error) {
	derBytes, privKey, err := selfca.GenerateCertificate(selfca.Certificate{
		IsCA:      true,
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Duration(365*24) * time.Hour),
		Hosts:     []string{"127.0.0.1"},
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "generating certificate")
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "marshaling private key")
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing key pair")
	}

	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing certificate")
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(cert.Leaf)

	return &cert, certPool, nil
}
