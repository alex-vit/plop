package engine

import (
	"crypto/tls"
	"os"

	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/tlsutil"
)

// GenerateCert creates a new TLS certificate and writes it to disk.
func GenerateCert(certFile, keyFile string) (tls.Certificate, error) {
	return tlsutil.NewCertificate(certFile, keyFile, "syncthing", 365*20, false)
}

// LoadOrGenerateCert loads an existing TLS certificate or generates a new one
// if the cert file doesn't exist.
func LoadOrGenerateCert(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		return cert, nil
	}
	if !os.IsNotExist(err) {
		return tls.Certificate{}, err
	}
	return GenerateCert(certFile, keyFile)
}

// DeviceID derives the Syncthing device ID from a TLS certificate.
func DeviceID(cert tls.Certificate) protocol.DeviceID {
	return protocol.NewDeviceID(cert.Certificate[0])
}
