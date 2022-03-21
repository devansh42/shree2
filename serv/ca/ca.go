package ca

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

const (
	organization = "Shree"
	country      = "IN"
	oneYear      = time.Hour * 24 * 365
)

type Certificate struct {
	Certificate, PrivateKey *bytes.Buffer
}
type CA struct {
	rootCertificate *x509.Certificate
	rootPriv        *rsa.PrivateKey
}

func New(caRootCertificate *x509.Certificate,
	caRootPrivKey *rsa.PrivateKey) (CA, error) {
	return CA{
		rootCertificate: caRootCertificate,
		rootPriv:        caRootPrivKey,
	}, nil
}

func (c CA) Create() (Certificate, error) {
	now := time.Now()
	var certificate Certificate
	cert := x509.Certificate{
		SerialNumber: big.NewInt(now.UnixNano()),
		Subject: pkix.Name{
			Organization: []string{organization},
			Country:      []string{country},
		},
		NotBefore:    now,
		NotAfter:     now.Add(oneYear),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privK, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Print("Couldn't create rsa private key: ", err)
		return certificate, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, c.rootCertificate, &privK.PublicKey, &c.rootPriv)
	if err != nil {
		log.Print("Couldn't create rsa certificate: ", err)
		return certificate, err
	}
	certBuffer := new(bytes.Buffer)
	privBuffer := new(bytes.Buffer)
	err = pem.Encode(certBuffer, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		log.Print("Couldn't encode rsa certificate: ", err)
		return certificate, err
	}
	err = pem.Encode(privBuffer, &pem.Block{

		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privK),
	})
	if err != nil {
		log.Print("Couldn't encode rsa private key: ", err)
		return certificate, err
	}
	certificate = Certificate{
		Certificate: certBuffer,
		PrivateKey:  privBuffer,
	}
	return certificate, nil
}

func DecodeCertificate(pemEncodedBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemEncodedBytes)
	cert, err := x509.ParseCertificate(block.Bytes)
	return cert, err
}

func DecodePrivateKey(pemEncodedBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemEncodedBytes)
	privK, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	return privK, err
}
