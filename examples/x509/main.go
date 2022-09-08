package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"path"
	"time"
)

func main() {
	// Create a root CA
	rootPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Write the private key to disk
	rootPrivateKeyBytes := x509.MarshalPKCS1PrivateKey(rootPrivateKey)

	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: rootPrivateKeyBytes}

	pemBytes := pem.EncodeToMemory(pemBlock)

	err = ioutil.WriteFile(path.Join("examples/x509/root", "key.pem"), pemBytes, 0600)
	if err != nil {
		panic(err)
	}

	rootTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "root"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		// TODO: Add KeyUsage
		// KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		// ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		PublicKeyAlgorithm:    x509.RSA,
		PublicKey:             &rootPrivateKey.PublicKey,
		BasicConstraintsValid: true,
	}

	// Create a root certificate authority
	rootCA, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		panic(err)
	}

	// Write the root CA to disk
	err = ioutil.WriteFile(path.Join("examples/x509/root", "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCA}), 0644)
	if err != nil {
		panic(err)
	}

	// Create a certificate for the server
	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Write the private key to disk
	serverPrivateKeyBytes := x509.MarshalPKCS1PrivateKey(serverPrivateKey)

	pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverPrivateKeyBytes}

	pemBytes = pem.EncodeToMemory(pemBlock)

	err = ioutil.WriteFile(path.Join("examples/x509/server", "key.pem"), pemBytes, 0600)
	if err != nil {
		panic(err)
	}

	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "server"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		// TODO: Add KeyUsage
		// KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		// ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		PublicKeyAlgorithm:    x509.RSA,
		PublicKey:             &serverPrivateKey.PublicKey,
		BasicConstraintsValid: true,
	}

	serverLeaf, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &rootTemplate, &serverPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		panic(err)
	}

	// Write the server certificate to disk
	err = ioutil.WriteFile(path.Join("examples/x509/server", "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverLeaf}), 0644)
	if err != nil {
		panic(err)
	}

	// Read the root CA from disk
	rootCABytes, err := ioutil.ReadFile(path.Join("examples/x509/root", "cert.pem"))
	if err != nil {
		panic(err)
	}

	pemBlock, _ = pem.Decode(rootCABytes)

	// Parse the root CA
	root, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		panic(err)
	}

	// Add the root CA to the pool
	roots := x509.NewCertPool()
	roots.AddCert(root)

	// Read the server certificate from disk
	serverBytes, err := ioutil.ReadFile(path.Join("examples/x509/server", "cert.pem"))
	if err != nil {
		panic(err)
	}

	// Parse the server certificate
	pemBlock, _ = pem.Decode(serverBytes)

	server, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		panic(err)
	}

	// Verify the server certificate
	opts := x509.VerifyOptions{
		Roots: roots,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	chains, err := server.Verify(opts)
	if err != nil {
		panic(err)
	}

	// Print the chains
	for i, chain := range chains {
		for _, cert := range chain {
			println(i, cert.Subject.CommonName)
		}
	}

	// Create a certificate for the client
	clientPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Write the private key to disk
	clientPrivateKeyBytes := x509.MarshalPKCS1PrivateKey(clientPrivateKey)

	pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: clientPrivateKeyBytes}

	pemBytes = pem.EncodeToMemory(pemBlock)

	err = ioutil.WriteFile(path.Join("examples/x509/client", "key.pem"), pemBytes, 0600)
	if err != nil {
		panic(err)
	}

	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "client"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		// TODO: Add KeyUsage
		// KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		// ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		PublicKeyAlgorithm:    x509.RSA,
		PublicKey:             &clientPrivateKey.PublicKey,
		BasicConstraintsValid: true,
	}

	clientLeaf, err := x509.CreateCertificate(rand.Reader, &clientTemplate, &rootTemplate, &clientPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		panic(err)
	}

	// Write the client certificate to disk
	err = ioutil.WriteFile(path.Join("examples/x509/client", "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientLeaf}), 0644)
	if err != nil {
		panic(err)
	}

	// Read the client certificate from disk
	clientBytes, err := ioutil.ReadFile(path.Join("examples/x509/client", "cert.pem"))
	if err != nil {
		panic(err)
	}

	// Parse the client certificate
	pemBlock, _ = pem.Decode(clientBytes)

	client, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		panic(err)
	}

	// Verify the client certificate
	opts = x509.VerifyOptions{
		Roots: roots,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	chains, err = client.Verify(opts)
	if err != nil {
		panic(err)
	}

	// Print the chains
	for i, chain := range chains {
		for _, cert := range chain {
			println(i, cert.Subject.CommonName)
		}
	}
}
