// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vault

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// Stashing functions here for generating a CA cert in the tests. Pulled mostly
// from the vault-k8s cert package.

func generateCA() ([]byte, error) {
	// Create the private key we'll use for this CA cert.
	signer, _, err := privateKey()
	if err != nil {
		return nil, err
	}

	// The serial number for the cert
	sn, err := serialNumber()
	if err != nil {
		return nil, err
	}

	signerKeyId, err := keyId(signer.Public())
	if err != nil {
		return nil, err
	}

	// Create the CA cert
	template := x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{CommonName: "Testing CA"},
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		NotAfter:              time.Now().Add(1 * time.Hour),
		NotBefore:             time.Now().Add(-1 * time.Minute),
		AuthorityKeyId:        signerKeyId,
		SubjectKeyId:          signerKeyId,
	}

	bs, err := x509.CreateCertificate(
		rand.Reader, &template, &template, signer.Public(), signer)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: bs})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// privateKey returns a new ECDSA-based private key. Both a crypto.Signer
// and the key in PEM format are returned.
func privateKey() (crypto.Signer, string, error) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", err
	}

	bs, err := x509.MarshalECPrivateKey(pk)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: bs})
	if err != nil {
		return nil, "", err
	}

	return pk, buf.String(), nil
}

// serialNumber generates a new random serial number.
func serialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
}

// keyId returns a x509 KeyId from the given signing key. The key must be
// an *ecdsa.PublicKey currently, but may support more types in the future.
func keyId(raw interface{}) ([]byte, error) {
	switch raw.(type) {
	case *ecdsa.PublicKey:
	default:
		return nil, fmt.Errorf("invalid key type: %T", raw)
	}

	// This is not standard; RFC allows any unique identifier as long as they
	// match in subject/authority chains but suggests specific hashing of DER
	// bytes of public key including DER tags.
	bs, err := x509.MarshalPKIXPublicKey(raw)
	if err != nil {
		return nil, err
	}

	// String formatted
	kID := sha256.Sum256(bs)
	return []byte(strings.Replace(fmt.Sprintf("% x", kID), " ", ":", -1)), nil
}
