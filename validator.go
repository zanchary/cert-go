package certgo

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/Alonza0314/cert-go/model"
)

// ValidateConfig validates the completeness and legality of the configuration file
func ValidateConfig(cfg *model.CAConfig) error {
	if err := validateRootCert(cfg.CA.Root); err != nil {
		return fmt.Errorf("root certificate: %w", err)
	}

	if err := validateIntermediateCert(cfg.CA.Intermediate); err != nil {
		return fmt.Errorf("intermediate certificate: %w", err)
	}

	if err := validateServerCert(cfg.CA.Server); err != nil {
		return fmt.Errorf("server certificate: %w", err)
	}

	if err := validateClientCert(cfg.CA.Client); err != nil {
		return fmt.Errorf("client certificate: %w", err)
	}

	return nil
}

func validateRootCert(cert model.Certificate) error {
	if cert.Type != string(CertTypeRoot) {
		return ErrInvalidCertType
	}

	if !cert.IsCA {
		return NewCertError("root certificate validation", fmt.Errorf("root certificate must be a CA certificate"))
	}

	return validateCommonFields(cert)
}

func validateIntermediateCert(cert model.Certificate) error {
	if cert.Type != string(CertTypeIntermediate) {
		return ErrInvalidCertType
	}

	if !cert.IsCA {
		return NewCertError("intermediate certificate validation", fmt.Errorf("intermediate certificate must be a CA certificate"))
	}

	if cert.ParentCertPath == "" {
		return NewCertError("intermediate certificate validation", ErrMissingField)
	}

	return validateCommonFields(cert)
}

func validateServerCert(cert model.Certificate) error {
	if cert.Type != string(CertTypeServer) {
		return ErrInvalidCertType
	}

	if cert.IsCA {
		return NewCertError("server certificate validation", fmt.Errorf("server certificate should not be a CA certificate"))
	}

	if len(cert.DNSNames) == 0 && len(cert.IPAddresses) == 0 {
		return NewCertError("server certificate validation", fmt.Errorf("at least one DNS name or IP address is required"))
	}

	return validateCommonFields(cert)
}

func validateClientCert(cert model.Certificate) error {
	if cert.Type != string(CertTypeClient) {
		return ErrInvalidCertType
	}

	if cert.IsCA {
		return NewCertError("client certificate validation", fmt.Errorf("client certificate should not be a CA certificate"))
	}

	return validateCommonFields(cert)
}

func validateCommonFields(cert model.Certificate) error {
	if cert.CommonName == "" {
		return NewCertError("common fields validation", fmt.Errorf("%w: common name", ErrMissingField))
	}

	if cert.Organization == "" {
		return NewCertError("common fields validation", fmt.Errorf("%w: organization", ErrMissingField))
	}

	if cert.ValidityYears < 0 || cert.ValidityMonth < 0 || cert.ValidityDay < 0 {
		return NewCertError("common fields validation", ErrInvalidPeriod)
	}

	if cert.KeyFilePath == "" {
		return NewCertError("common fields validation", fmt.Errorf("%w: key file path", ErrMissingField))
	}

	if cert.CertFilePath == "" {
		return NewCertError("common fields validation", fmt.Errorf("%w: certificate file path", ErrMissingField))
	}

	// Validate file paths
	if err := validateFilePath(cert.KeyFilePath); err != nil {
		return WrapPathError(cert.KeyFilePath, err)
	}

	if err := validateFilePath(cert.CertFilePath); err != nil {
		return WrapPathError(cert.CertFilePath, err)
	}

	if cert.CsrFilePath != "" {
		if err := validateFilePath(cert.CsrFilePath); err != nil {
			return WrapPathError(cert.CsrFilePath, err)
		}
	}

	// Validate IP address format
	for _, ip := range cert.IPAddresses {
		if net.ParseIP(ip) == nil {
			return NewCertError("common fields validation", fmt.Errorf("%w: %s", ErrInvalidIP, ip))
		}
	}

	// Validate URI format
	for _, uri := range cert.URIs {
		if !strings.Contains(uri, ".") {
			return NewCertError("common fields validation", fmt.Errorf("%w: %s", ErrInvalidURI, uri))
		}
	}

	return nil
}

func validateFilePath(path string) error {
	// Check if path is empty
	if path == "" {
		return ErrInvalidPath
	}

	// Check if path contains invalid characters
	if strings.ContainsAny(path, "<>:\"|?*") {
		return fmt.Errorf("path contains invalid characters")
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute")
	}

	return nil
}
