package efincore

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math"
	"math/big"
	"sync"
	"time"

	_ "embed"
)

//go:embed files/efin-proxy-ca.key
var defaultProxyPrivateKey []byte

//go:embed files/efin-proxy-ca.crt
var defaultProxyCertificate []byte

type CA struct {
	proxyCertificate []byte
	proxyPrivateKey  []byte
	cache            *sync.Map
}

func NewCA() *CA {
	return &CA{
		proxyCertificate: defaultProxyCertificate,
		proxyPrivateKey:  defaultProxyPrivateKey,
		cache:            &sync.Map{},
	}
}

func (c *CA) GetCertificateFor(domain string) (tls.Certificate, error) {
	cachedCert, ok := c.cache.Load(domain)
	if ok {
		return cachedCert.(tls.Certificate), nil
	}

	caTLSCert, err := tls.X509KeyPair(defaultProxyCertificate, defaultProxyPrivateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	caCert, err := x509.ParseCertificate(caTLSCert.Certificate[0])
	if err != nil {
		return tls.Certificate{}, err
	}

	if err != nil {
		return tls.Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{
		SerialNumber: serial,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		DNSNames:     []string{domain},
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &certPrivKey.PublicKey, caTLSCert.PrivateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return tls.Certificate{}, err
	}

	c.cache.Store(domain, serverCert)
	return serverCert, nil
}

func DefaultProxyCertificate() string {
	return string(defaultProxyCertificate)
}
