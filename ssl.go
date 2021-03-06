package frontman

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const timeoutPortLookup = time.Second * 3

func certName(cert *x509.Certificate) string {
	return fmt.Sprintf("'%s' issued by %s", cert.Subject.CommonName, cert.Issuer.CommonName)
}

func (fm *Frontman) runSSLCheck(hostname string, port int, service string) (m MeasurementsMap, err error) {
	service = strings.ToLower(service)

	if net.ParseIP(hostname) != nil {
		hostname = ""
	}

	if port == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutPortLookup)
		defer cancel()

		if p, exists := defaultPortByService[service]; exists {
			port = p
		} else if p, lerr := net.DefaultResolver.LookupPort(ctx, "tcp", service); p > 0 {
			port = p
		} else if lerr != nil {
			err = fmt.Errorf("failed to auto-determine port for '%s': %s", service, lerr.Error())
			return
		}
	}

	prefix := fmt.Sprintf("net.tcp.ssl.%d.", port)

	m = MeasurementsMap{
		prefix + "success": 0,
	}

	addr := fmt.Sprintf("%s:%d", hostname, port)
	dialer := net.Dialer{Timeout: secToDuration(fm.Config.NetTCPTimeout)}
	connection, err := tls.DialWithDialer(
		&dialer,
		"tcp",
		addr,
		&tls.Config{ServerName: hostname},
	)
	if err != nil {
		logrus.Debugf("serviceCheck: SSL check %s for '%s' failed: %s", addr, hostname, err.Error())
		if strings.HasPrefix(err.Error(), "tls:") {
			err = fmt.Errorf("service doesn't support SSL")
		} else {
			err = fmt.Errorf(strings.TrimPrefix(err.Error(), "x509: "))
		}
		return
	}

	defer connection.Close()

	remainingValidity, firstCertToExpire := findCertRemainingValidity(connection.ConnectionState().VerifiedChains)
	m[prefix+"expiryDaysRemaining"] = remainingValidity

	if remainingValidity <= float64(fm.Config.SSLCertExpiryThreshold) {
		err = fmt.Errorf("certificate will expire soon: %s", certName(firstCertToExpire))
		return
	}

	m[prefix+"success"] = 1
	return
}

func findCertRemainingValidity(certChains [][]*x509.Certificate) (float64, *x509.Certificate) {
	var remainingValidity float64
	var firstToExpire *x509.Certificate

	// find chain with max remaining validity
	for _, chain := range certChains {
		chainRemainingValidity, c := findChainRemainingValidity(chain)
		if chainRemainingValidity > remainingValidity {
			remainingValidity = chainRemainingValidity
			firstToExpire = c
		}
	}
	return remainingValidity, firstToExpire
}

func findChainRemainingValidity(chain []*x509.Certificate) (float64, *x509.Certificate) {
	var min = math.MaxFloat64
	var firstToExpire *x509.Certificate

	// find cert that will expire first
	for _, cert := range chain {
		remainingValidity := time.Until(cert.NotAfter).Hours() / 24
		if remainingValidity < min {
			min = remainingValidity
			firstToExpire = cert
		}
	}
	return min, firstToExpire
}
