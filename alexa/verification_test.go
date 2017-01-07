package alexa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func readFile(fileName string) []byte {
	testCert, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return testCert
}

func getCertFromFile(file string) (*x509.Certificate, error) {
	block, _ := pem.Decode(readFile(file))
	cert, err := x509.ParseCertificate(block.Bytes)
	return cert, err
}

func Test_verifyRequestSignerURL(t *testing.T) {
	type args struct {
		signatureCertChainURLStr string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid URL", wantErr: false,
			args: args{signatureCertChainURLStr: "https://s3.amazonaws.com:443/echo.api/echo-api-cert.pem"},
		}, {
			name: "Relative Path In URL", wantErr: false,
			args: args{signatureCertChainURLStr: "https://s3.amazonaws.com/echo.api/../echo.api/echo-api-cert.pem"},
		}, {
			name: "Invalid Protocol", wantErr: true,
			args: args{signatureCertChainURLStr: "http://s3.amazonaws.com/echo.api/echo-api-cert.pem"},
		}, {
			name: "Invalid Hostname", wantErr: true,
			args: args{signatureCertChainURLStr: "https://notamazon.com/echo.api/echo-api-cert.pem"},
		}, {
			name: "Invalid Path", wantErr: true,
			args: args{signatureCertChainURLStr: "https://s3.amazonaws.com/EcHo.aPi/echo-api-cert.pem"},
		}, {
			name: "Invalid Path", wantErr: true,
			args: args{signatureCertChainURLStr: "https://s3.amazonaws.com/invalid.path/echo-api-cert.pem"},
		}, {
			name: "Invalid Port", wantErr: true,
			args: args{signatureCertChainURLStr: "https://s3.amazonaws.com:563/echo.api/echo-api-cert.pem"},
		}, {
			name: "Invalid URL", wantErr: true,
			args: args{signatureCertChainURLStr: "not a valid url"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := verifyRequestSignerURL(tt.args.signatureCertChainURLStr); (err != nil) != tt.wantErr {
				t.Errorf("verifyRequestSignerURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_verifyCertChain(t *testing.T) {
	type args struct {
		certFilename string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Valid Cert", wantErr: false,
			args: args{certFilename: "_testdata/testCA/valid-chain.crt.pem"},
		},
		{
			name: "Expired Cert", wantErr: true,
			args: args{certFilename: "_testdata/testCA/expired-chain.crt.pem"},
		},
		{
			name: "Expired Intermediate", wantErr: true,
			args: args{certFilename: "_testdata/testCA/expired-intermediate.crt.pem"},
		},
		{
			name: "Missing Intermediate", wantErr: true,
			args: args{certFilename: "_testdata/testCA/missing-intermediate.crt.pem"},
		},
		{
			name: "Missing Root", wantErr: true,
			args: args{certFilename: "_testdata/testCA/missing-root.crt.pem"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signingCert, intermediates, root := parseCertChain(readFile(tt.args.certFilename))
			if err := verifyCertChain(signingCert, intermediates, root); (err != nil) != tt.wantErr {
				t.Errorf("verifyCertChain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getCert(t *testing.T) {
	type args struct {
		url string
	}
	//good, err := getCertFromFile("_testdata/testCA/valid-chain.crt.pem")
	//if err != nil {
	//	t.Errorf("Failed to load cert good.pem")
	//}
	tests := []struct {
		name    string
		args    args
		want    *x509.Certificate
		wantErr bool
	}{
	//{"Valid Cert", args{url: "https://s3.amazonaws.com:443/echo.api/echo-api-cert-4.pem"}, good},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCert(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getRequest(body []byte, headers map[string]string) *gin.Context {
	req := httptest.NewRequest("GET", "/", bytes.NewReader(body))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return &gin.Context{
		Request: req,
	}
}

func Test_VerifyRequestSignature(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	cert, err := getCertFromFile("_testdata/testCA/server.crt.pem")
	if err != nil {
		t.Fatalf("Failed to load cert: %s", err)
	}
	certStore["https://test-cert.localhost/server.crt.pem"] = cert

	msg := []byte(`This is the message I want signed`)

	h := crypto.SHA1.New()
	h.Write(msg)
	d := h.Sum(nil)

	rawKey, err := ioutil.ReadFile("_testdata/testCA/server.key.pem")
	if err != nil {
		t.Fatalf("Failed to load cert: %s", err)
	}

	block, _ := pem.Decode(rawKey)
	if block == nil {
		t.Fatalf("Failed to load cert: server.key.pem, no pem block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to load cert: %s", err)
	}
	keyObj := key.(*rsa.PrivateKey)
	signed, err := rsa.SignPKCS1v15(rand.Reader, keyObj, crypto.SHA1, d)
	if err != nil {
		t.Fatalf("Failed to sign data: %s", err)
	}

	encoded := base64.StdEncoding.EncodeToString(signed)

	tests := []struct {
		name string
		args args
		want error
	}{
		{"valid", args{ctx: getRequest(msg, map[string]string{"Signature": encoded, "SignatureCertChainUrl": "https://test-cert.localhost/server.crt.pem"})}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyRequestSignature(tt.args.ctx); got != tt.want {
				t.Errorf("VerifyRequestSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
