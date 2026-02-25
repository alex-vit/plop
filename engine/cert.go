package engine

import (
	"crypto/tls"

	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/tlsutil"
)

// GenerateCert creates a new TLS certificate and writes it to disk.
func GenerateCert(certFile, keyFile string) (tls.Certificate, error) {
	return tlsutil.NewCertificate(certFile, keyFile, "syncthing", 365*20)
}

// LoadCert loads an existing TLS certificate from disk.
func LoadCert(certFile, keyFile string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(certFile, keyFile)
}

// DeviceID derives the Syncthing device ID from a TLS certificate.
func DeviceID(cert tls.Certificate) protocol.DeviceID {
	return protocol.NewDeviceID(cert.Certificate[0])
}
