package checker

import (
	"crypto/tls"
	"errors"
	"log"
	"net/url"
	"time"
)

// isHTTPS checks if the URL uses HTTPS
func isHTTPS(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "https"
}

// checkSSLCertificates checks the SSL certificates of the URL
func checkSSLCertificates(urlStr string) (remainingDays int, isExpired bool, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return 0, false, err
	}

	if u.Scheme != "https" {
		return 0, false, errors.New("not an https URL")
	}

	// default 443
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "443"
	}
	address := host + ":" + port

	conn, err := tls.Dial("tcp", address, &tls.Config{
		ServerName: host,
	})
	if err != nil {
		return 0, false, err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Error closing TLS connection: %v", closeErr)
		}
	}()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return 0, false, errors.New("no certificates found")
	}

	// Get the expiration date of the first certificate (leaf certificate)
	cert := certs[0]
	expirationDate := cert.NotAfter
	currentTime := time.Now()

	// Calculate remaining days
	remainingDuration := expirationDate.Sub(currentTime)
	remainingDays = int(remainingDuration.Hours() / 24)

	// Check if the certificate is expired
	isExpired = currentTime.After(expirationDate)

	// Additional certificate validation
	if currentTime.Before(cert.NotBefore) {
		return remainingDays, true, errors.New("certificate is not yet valid")
	}

	//log.Printf("Certificate details for %s: Issuer=%s, Subject=%s, NotBefore=%s, NotAfter=%s",
	//	urlStr, cert.Issuer.CommonName, cert.Subject.CommonName,
	//	cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))

	return remainingDays, isExpired, nil
}
