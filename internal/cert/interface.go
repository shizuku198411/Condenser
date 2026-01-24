package cert

type CertHandler interface {
	EnsureSelfSignedCert(certPath string, keyPath string, cfg CertConfig) error
}
