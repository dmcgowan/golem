package jsonsign

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/docker/libtrust"
)

func createTestJSON(sigKey string, indent string) (map[string]interface{}, []byte) {
	testMap := map[string]interface{}{
		"name": "dmcgowan/mycontainer",
		"config": map[string]interface{}{
			"ports": []int{9101, 9102},
			"run":   "/bin/echo \"Hello\"",
		},
		"layers": []string{
			"2893c080-27f5-11e4-8c21-0800200c9a66",
			"c54bc25b-fbb2-497b-a899-a8bc1b5b9d55",
			"4d5d7e03-f908-49f3-a7f6-9ba28dfe0fb4",
			"0b6da891-7f7f-4abf-9c97-7887549e696c",
			"1d960389-ae4f-4011-85fd-18d0f96a67ad",
		},
	}
	formattedSection := `{"config":{"ports":[9101,9102],"run":"/bin/echo \"Hello\""},"layers":["2893c080-27f5-11e4-8c21-0800200c9a66","c54bc25b-fbb2-497b-a899-a8bc1b5b9d55","4d5d7e03-f908-49f3-a7f6-9ba28dfe0fb4","0b6da891-7f7f-4abf-9c97-7887549e696c","1d960389-ae4f-4011-85fd-18d0f96a67ad"],"name":"dmcgowan/mycontainer","%s":[{"header":{`
	formattedSection = fmt.Sprintf(formattedSection, sigKey)
	if indent != "" {
		buf := bytes.NewBuffer(nil)
		json.Indent(buf, []byte(formattedSection), "", indent)
		return testMap, buf.Bytes()
	}
	return testMap, []byte(formattedSection)

}

func TestSignJSON(t *testing.T) {
	key, err := GenerateECP256PrivateKey()
	if err != nil {
		t.Fatalf("Error generating EC key: %s", err)
	}

	testMap, _ := createTestJSON("buildSignatures", "   ")
	indented, err := json.MarshalIndent(testMap, "", "   ")
	if err != nil {
		t.Fatalf("Marshall error: %s", err)
	}

	js, err := NewJSONSignature(indented)
	if err != nil {
		t.Fatalf("Error creating JSON signature: %s", err)
	}
	err = js.Sign(key)
	if err != nil {
		t.Fatalf("Error signing content: %s", err)
	}

	keys, err := js.Verify()
	if err != nil {
		t.Fatalf("Error verifying signature: %s", err)
	}
	if len(keys) != 1 {
		t.Fatalf("Error wrong number of keys returned")
	}
	if keys[0].KeyID() != key.KeyID() {
		t.Fatalf("Unexpected public key returned")
	}

}

func TestSignMap(t *testing.T) {
	key, err := GenerateECP256PrivateKey()
	if err != nil {
		t.Fatalf("Error generating EC key: %s", err)
	}

	testMap, _ := createTestJSON("buildSignatures", "   ")
	js, err := NewJSONSignatureFromMap(testMap)
	if err != nil {
		t.Fatalf("Error creating JSON signature: %s", err)
	}
	err = js.Sign(key)
	if err != nil {
		t.Fatalf("Error signing JSON signature: %s", err)
	}

	keys, err := js.Verify()
	if err != nil {
		t.Fatalf("Error verifying signature: %s", err)
	}
	if len(keys) != 1 {
		t.Fatalf("Error wrong number of keys returned")
	}
	if keys[0].KeyID() != key.KeyID() {
		t.Fatalf("Unexpected public key returned")
	}
}

func TestFormattedJson(t *testing.T) {
	key, err := GenerateECP256PrivateKey()
	if err != nil {
		t.Fatalf("Error generating EC key: %s", err)
	}

	testMap, firstSection := createTestJSON("buildSignatures", "     ")
	indented, err := json.MarshalIndent(testMap, "", "     ")
	if err != nil {
		t.Fatalf("Marshall error: %s", err)
	}

	js, err := NewJSONSignature(indented)
	if err != nil {
		t.Fatalf("Error creating JSON signature: %s", err)
	}
	err = js.Sign(key)
	if err != nil {
		t.Fatalf("Error signing content: %s", err)
	}

	b, err := js.PrettySignature("buildSignatures")
	if err != nil {
		t.Fatalf("Error signing map: %s", err)
	}

	if bytes.Compare(b[:len(firstSection)], firstSection) != 0 {
		t.Fatalf("Wrong signed value\nExpected:\n%s\nActual:\n%s", firstSection, b[:len(firstSection)])
	}

	parsed, err := ParsePrettySignature(b, "buildSignatures")
	if err != nil {
		t.Fatalf("Error parsing formatted signature: %s", err)
	}

	keys, err := parsed.Verify()
	if err != nil {
		t.Fatalf("Error verifying signature: %s", err)
	}
	if len(keys) != 1 {
		t.Fatalf("Error wrong number of keys returned")
	}
	if keys[0].KeyID() != key.KeyID() {
		t.Fatalf("Unexpected public key returned")
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(b, &unmarshalled)
	if err != nil {
		t.Fatalf("Could not unmarshall after parse: %s", err)
	}

}

func TestFormattedFlatJson(t *testing.T) {
	key, err := GenerateECP256PrivateKey()
	if err != nil {
		t.Fatalf("Error generating EC key: %s", err)
	}

	testMap, firstSection := createTestJSON("buildSignatures", "")
	unindented, err := json.Marshal(testMap)
	if err != nil {
		t.Fatalf("Marshall error: %s", err)
	}

	js, err := NewJSONSignature(unindented)
	if err != nil {
		t.Fatalf("Error creating JSON signature: %s", err)
	}
	err = js.Sign(key)
	if err != nil {
		t.Fatalf("Error signing JSON signature: %s", err)
	}

	b, err := js.PrettySignature("buildSignatures")
	if err != nil {
		t.Fatalf("Error signing map: %s", err)
	}

	if bytes.Compare(b[:len(firstSection)], firstSection) != 0 {
		t.Fatalf("Wrong signed value\nExpected:\n%s\nActual:\n%s", firstSection, b[:len(firstSection)])
	}

	parsed, err := ParsePrettySignature(b, "buildSignatures")
	if err != nil {
		t.Fatalf("Error parsing formatted signature: %s", err)
	}

	keys, err := parsed.Verify()
	if err != nil {
		t.Fatalf("Error verifying signature: %s", err)
	}
	if len(keys) != 1 {
		t.Fatalf("Error wrong number of keys returned")
	}
	if keys[0].KeyID() != key.KeyID() {
		t.Fatalf("Unexpected public key returned")
	}
}

func generateTrustCA() (PrivateKey, *x509.Certificate) {
	key, err := GenerateECP256PrivateKey()
	if err != nil {
		panic(err)
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "CA Root",
		},
		NotBefore:             time.Now().Add(-time.Second),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader, cert, cert,
		key.CryptoPublicKey(), key.CryptoPrivateKey(),
	)
	if err != nil {
		panic(err)
	}

	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		panic(err)
	}

	return key, cert
}

func generateIntermediate(key libtrust.PublicKey, parentKey PrivateKey, parent *x509.Certificate) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "Intermediate",
		},
		NotBefore:             time.Now().Add(-time.Second),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader, cert, parent,
		key.CryptoPublicKey(), parentKey.CryptoPrivateKey(),
	)
	if err != nil {
		panic(err)
	}

	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		panic(err)
	}

	return cert
}

func generateTrustCert(key libtrust.PublicKey, parentKey PrivateKey, parent *x509.Certificate) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "Trust Cert",
		},
		NotBefore:             time.Now().Add(-time.Second),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader, cert, parent,
		key.CryptoPublicKey(), parentKey.CryptoPrivateKey(),
	)
	if err != nil {
		panic(err)
	}

	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		panic(err)
	}

	return cert
}

func generateTrustChain(key PrivateKey, ca *x509.Certificate) (PrivateKey, []*x509.Certificate) {
	parent := ca
	parentKey := key
	chain := make([]*x509.Certificate, 6)
	for i := 5; i > 0; i-- {
		intermediatekey, err := GenerateECP256PrivateKey()
		if err != nil {
			panic(err)
		}
		chain[i] = generateIntermediate(intermediatekey, parentKey, parent)
		parent = chain[i]
		parentKey = intermediatekey
	}
	trustKey, err := GenerateECP256PrivateKey()
	if err != nil {
		panic(err)
	}
	chain[0] = generateTrustCert(trustKey, parentKey, parent)

	return trustKey, chain
}

func TestChainVerify(t *testing.T) {
	caKey, ca := generateTrustCA()
	trustKey, chain := generateTrustChain(caKey, ca)

	testMap, _ := createTestJSON("verifySignatures", "   ")
	js, err := NewJSONSignatureFromMap(testMap)
	if err != nil {
		t.Fatalf("Error creating JSONSignature from map: %s", err)
	}

	err = js.SignWithChain(trustKey, chain)
	if err != nil {
		t.Fatalf("Error signing with chain: %s", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)
	chains, err := js.VerifyChains(pool)
	if err != nil {
		t.Fatalf("Error verifying content: %s", err)
	}
	if len(chains) != 1 {
		t.Fatalf("Unexpected chains length: %d", len(chains))
	}
	if len(chains[0]) != 7 {
		t.Fatalf("Unexpected chain length: %d", len(chains[0]))
	}
}

func TestInvalidChain(t *testing.T) {
	caKey, ca := generateTrustCA()
	trustKey, chain := generateTrustChain(caKey, ca)

	testMap, _ := createTestJSON("verifySignatures", "   ")
	js, err := NewJSONSignatureFromMap(testMap)
	if err != nil {
		t.Fatalf("Error creating JSONSignature from map: %s", err)
	}

	err = js.SignWithChain(trustKey, chain[:5])
	if err != nil {
		t.Fatalf("Error signing with chain: %s", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)
	chains, err := js.VerifyChains(pool)
	if err == nil {
		t.Fatalf("Expected error verifying with bad chain")
	}
	if len(chains) != 0 {
		t.Fatalf("Unexpected chains returned from invalid verify")
	}
}
