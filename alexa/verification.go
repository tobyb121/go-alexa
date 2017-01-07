package alexa

import (
	"io/ioutil"

	"crypto/x509"

	"net/http"
	"net/url"

	"strings"

	"encoding/json"

	"time"

	"math"

	"encoding/pem"

	"encoding/base64"

	"bytes"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/tobyb121/go-alexa/alexa/entities"
)

const signatureAlgorithm = x509.SHA1WithRSA

const validSigningCertChainProtocol = "https"
const validSigningCertChainURLHostName = "s3.amazonaws.com"
const validSigningCertChainURLPathPrefix = "/echo.api/"

const validSigningCertName = "echo-api.amazon.com"

var certStore = make(map[string]*x509.Certificate)

const timestampTolerance = 150

func (s *Server) VerifyRequest(ctx *gin.Context, req *entities.Request) error {
	if s.VerifyRequests {
		if err := VerifyRequestSignature(ctx); err != nil {
			return err
		}

		if err := VerifyRequestTimestamp(req); err != nil {
			return err
		}

		if err := s.VerifyRequestApplication(req); err != nil {
			return err
		}
	}
	return nil
}

func verifyRequestSignerURL(signatureCertChainURLStr string) error {
	signatureCertChainURL, err := url.Parse(signatureCertChainURLStr)
	if err != nil {
		return err
	}

	if signatureCertChainURL.Scheme != validSigningCertChainProtocol {
		return errors.New("invalid scheme: " + signatureCertChainURL.Scheme)
	}

	if strings.ContainsAny(signatureCertChainURL.Host, ":") {
		hostParts := strings.Split(signatureCertChainURL.Host, ":")
		if len(hostParts) == 2 {
			signatureCertChainURL.Host = hostParts[0]
			if hostParts[1] != "443" {
				return errors.New("bad port: " + string(hostParts[1]))
			}
		} else {
			return errors.New("invalid host format: " + signatureCertChainURL.Host)
		}
	}

	if strings.ToLower(signatureCertChainURL.Host) != validSigningCertChainURLHostName {
		return errors.New("invalid hostname: " + signatureCertChainURL.Host)
	}

	if !strings.HasPrefix(signatureCertChainURL.Path, validSigningCertChainURLPathPrefix) {
		return errors.New("invalid path prefix: " + signatureCertChainURL.Path)
	}

	return nil
}

func getCert(url string) (*x509.Certificate, error) {
	if certStore[url] == nil {
		err := verifyRequestSignerURL(url)
		if err != nil {
			return nil, err
		}
		certData, err := fetchCert(url)
		if err != nil {
			return nil, err
		}
		signingCert, intermediates, root := parseCertChain(certData)
		if signingCert == nil || root == nil {
			return nil, errors.New("incomplete certificate chain")
		}
		err = verifyCertChain(signingCert, intermediates, root)
		if err == nil {
			certStore[url] = signingCert
		}
	}
	return certStore[url], nil
}

func fetchCert(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func parseCertChain(buf []byte) (signingCert *x509.Certificate, intermediates []*x509.Certificate, root *x509.Certificate) {
	var signingCertData *pem.Block
	signingCertData, buf = pem.Decode(buf)
	if signingCertData == nil {
		return nil, nil, nil
	}
	signingCert, err := x509.ParseCertificate(signingCertData.Bytes)
	if err != nil {
		return nil, nil, nil
	}
	for len(buf) > 0 {
		var certData *pem.Block
		certData, buf = pem.Decode(buf)
		if certData == nil {
			break
		}
		cert, err := x509.ParseCertificate(certData.Bytes)
		if err != nil {
			continue
		}
		if bytes.Compare(cert.SubjectKeyId, cert.AuthorityKeyId) == 0 {
			root = cert
		} else {
			intermediates = append(intermediates, cert)
		}
	}
	return
}

func verifyCertChain(signingCert *x509.Certificate, intermediates []*x509.Certificate, root *x509.Certificate) error {
	intermediatesPool := x509.NewCertPool()
	for c := range intermediates {
		intermediatesPool.AddCert(intermediates[c])
	}
	rootsPool := x509.NewCertPool()
	if root != nil {
		rootsPool.AddCert(root)
	}
	opts := x509.VerifyOptions{
		Intermediates: intermediatesPool,
		Roots:         rootsPool,
	}
	_, err := signingCert.Verify(opts)
	if err != nil {
		return err
	}

	err = signingCert.VerifyHostname(validSigningCertName)
	if err != nil {
		return err
	}
	return err
}

func VerifyRequestSignature(ctx *gin.Context) error {
	signatureCertChainURLStr := ctx.Request.Header.Get("SignatureCertChainUrl")
	if signatureCertChainURLStr == "" {
		return errors.New("request missing header 'SignatureCertChainUrl'")
	}

	signingCert, err := getCert(signatureCertChainURLStr)
	if err != nil {
		return err
	}

	encodedSignature := ctx.Request.Header.Get("Signature")
	if encodedSignature == "" {
		return errors.New("request missing header 'Signature'")
	}
	signature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		return err
	}

	bodyData, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}

	// Recreate the body for future reads
	ctx.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyData))

	signErr := signingCert.CheckSignature(signatureAlgorithm, bodyData, signature)
	if signErr != nil {
		return signErr
	}
	return nil
}

func VerifyRequestTimestamp(req *entities.Request) error {
	var baseReq entities.BaseRequest
	if json.Unmarshal(req.Request, &baseReq) != nil {
		return errors.New("Invalid request format")
	}
	t, err := time.Parse(time.RFC3339, baseReq.Timestamp)
	if err != nil {
		return errors.New("Invalid timestamp: " + err.Error())
	}
	if math.Abs(time.Since(t).Seconds()) > timestampTolerance {
		return errors.New("Expired Request")
	}
	return nil
}

func (s *Server) VerifyRequestApplication(req *entities.Request) error {
	appID := req.Session.Application.ApplicationID
	if appID != s.Application.ApplicationID {
		return errors.New("Request sent with invalid Application ID")
	}
	return nil
}
