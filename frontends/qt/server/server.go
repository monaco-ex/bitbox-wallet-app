package main

/*
#ifndef BACKEND_H
#define BACKEND_H
#include <string.h>
#include <stdint.h>

typedef void (*pushNotificationsCallback) (const char*);
static void pushNotify(pushNotificationsCallback f, const char* msg) {
    f(msg);
}

typedef struct ConnectionData {
    int port;
    char* token;
    char* certFilename;
} ConnectionData;
#endif
*/
import "C"

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/shiftdevices/godbb/util/errp"
	"github.com/shiftdevices/godbb/util/jsonp"
	"github.com/shiftdevices/godbb/util/random"

	"github.com/shiftdevices/godbb/backend"
	backendHandlers "github.com/shiftdevices/godbb/backend/handlers"
	"github.com/shiftdevices/godbb/util/freeport"
	"github.com/shiftdevices/godbb/util/logging"
	"github.com/sirupsen/logrus"
)

const (
	// RSA key size.
	rsaBits = 2048
)

var theBackend *backend.Backend
var handlers *backendHandlers.Handlers
var token string

// generateRSAPrivateKey generates an RSA key pair and wraps it in the type rsa.PrivateKey.
func generateRSAPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		log.Fatalf("Failed to create private key: %s", err)
		return nil, err
	}
	return privateKey, nil
}

// createSelfSignedCertificate creates a self-signed certificate from the given rsa.PrivateKey.
func createSelfSignedCertificate(privateKey *rsa.PrivateKey, log *logrus.Entry) ([]byte, error) {
	serialNumber := big.Int{}
	notBefore := time.Now()
	// Invalid after one day.
	notAfter := notBefore.AddDate(0, 0, 1)
	template := x509.Certificate{
		SerialNumber: &serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CH"},
			Organization:       []string{"Shift Cryptosecurity"},
			OrganizationalUnit: []string{"godbb"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.ParseIP("::1")},
		DNSNames:              []string{"localhost"},
		IsCA:                  true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		log.WithField("error", err).Error("Failed to create x.509 certificate")
		return nil, err
	}
	return derBytes, nil
}

// saveAsPEM saves the given PEM block as a file
func saveAsPEM(name string, pemBytes *pem.Block, log *logrus.Entry) error {
	certificateDir := filepath.Dir(name)
	err := os.MkdirAll(certificateDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return errp.WithContext(errp.WithMessage(err, "Failed to create directory for server certificate"),
			errp.Context{"certificate-directory": certificateDir})
	}
	pemFile, err := os.Create(name)
	if err != nil {
		return errp.WithContext(errp.WithMessage(err, "Failed to create server certificate"),
			errp.Context{"file": name})
	}
	err = pem.Encode(pemFile, pemBytes)
	if err != nil {
		return errp.WithContext(errp.WithMessage(err, "Failed to write PEM encoded server certificate file"),
			errp.Context{"file": pemFile.Name()})
	}
	err = pemFile.Close()
	if err != nil {
		return errp.WithContext(errp.WithMessage(err, "Failed to close server certificate file"),
			errp.Context{"file": pemFile.Name()})
	}
	return nil
}

// derToPem wraps the givem PEM bytes and PEM type in a PEM block.
func derToPem(pemType string, pemBytes []byte) *pem.Block {
	return &pem.Block{Type: pemType, Bytes: pemBytes}
}

// Copied and adapted from package http server.go.
//
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

// accept enables TCP keep alive and sets the period to 3 minutes.
func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

//export backendCall
func backendCall(s *C.char) *C.char {
	if handlers == nil {
		return C.CString("null")
	}
	query := map[string]string{}
	jsonp.MustUnmarshal([]byte(C.GoString(s)), &query)
	if query["method"] != "POST" && query["method"] != "GET" {
		panic(errp.Newf("method must be POST or GET, got: %s", query["method"]))
	}
	rec := httptest.NewRecorder()
	request, err := http.NewRequest(query["method"], "/api/"+query["endpoint"], strings.NewReader(query["body"]))
	if err != nil {
		panic(errp.WithStack(err))
	}
	request.Header.Set("Authorization", "Basic "+token)
	handlers.Router.ServeHTTP(rec, request)
	response := rec.Result()
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(errp.WithStack(err))
	}
	return C.CString(string(responseBytes))
}

//export serve
func serve(pushNotificationsCallback C.pushNotificationsCallback) C.struct_ConnectionData {
	log := logging.Log.WithGroup("server")
	log.Info("--------------- Started application --------------")
	var err error
	token, err = random.HexString(16)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to generate random string")
	}
	port, err := freeport.FreePort(log)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to find free port")
	}
	cWrappedConnectionData := C.struct_ConnectionData{
		token: C.CString(token),
		port:  C.int(port),
	}
	log.WithField("port", port).Debug("Serve backend")

	connectionData := backendHandlers.NewConnectionData(port, token)
	productionArguments := backend.ProductionArguments()
	theBackend := backend.NewBackend(productionArguments)
	events := theBackend.Events()
	go func() {
		for {
			C.pushNotify(pushNotificationsCallback, C.CString(string(jsonp.MustMarshal(<-events))))
		}
	}()
	handlers = backendHandlers.NewHandlers(theBackend, connectionData)

	privateKey, err := generateRSAPrivateKey()
	if err != nil {
		log.WithField("error", err).Fatal("Failed to generate RSA key")
	}
	certificate, err := createSelfSignedCertificate(privateKey, log)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create self-signed certificate")
	}
	certificatePEM := derToPem("CERTIFICATE", certificate)
	tlsServerCertificatePath := path.Join(
		productionArguments.MainDirectoryPath(), "server.pem")
	saveAsPEM(tlsServerCertificatePath, certificatePEM, log)
	cWrappedConnectionData.certFilename = C.CString(tlsServerCertificatePath)
	var certAndKey tls.Certificate
	certAndKey.Certificate = [][]byte{certificate}
	certAndKey.PrivateKey = privateKey

	go func() {
		server := &http.Server{
			Addr:    fmt.Sprintf("localhost:%d", port),
			Handler: handlers.ServeFrontendHandler(),
			TLSConfig: &tls.Config{
				NextProtos:   []string{"http/1.1"},
				Certificates: []tls.Certificate{certAndKey},
			},
		}
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "address": server.Addr}).Fatal("Failed to listen on address")
		}
		log.WithField("address", server.Addr).Debug("Listening")
		//tlsListener := tls.NewListener(tcpKeepAliveListener{listener.(*net.TCPListener)}, server.TLSConfig)
		err = server.Serve(listener)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "address": server.Addr}).Fatal("Failed to establish TLS endpoint")
		}
	}()
	return cWrappedConnectionData
}

// Don't remove - needed for the C compilation.
func main() {
}
