package certutil

import (
	"io/ioutil"
	"os"

	"github.com/blend/go-sdk/exception"
)

// KeyPair is an x509 pem key pair as strings.
type KeyPair struct {
	Cert     string `json:"cert,omitempty" yaml:"cert,omitempty"`
	CertPath string `json:"certPath,omitempty" yaml:"certPath,omitempty"`
	Key      string `json:"key,omitempty" yaml:"key,omitempty"`
	KeyPath  string `json:"keyPath,omitempty" yaml:"keyPath,omitempty"`
}

// IsZero returns if the key pair is set or not.
func (kp KeyPair) IsZero() bool {
	return kp.Cert == "" &&
		kp.Key == "" &&
		kp.CertPath == "" &&
		kp.KeyPath == ""
}

// CertBytes returns the key pair cert bytes.
func (kp KeyPair) CertBytes() ([]byte, error) {
	if kp.Cert != "" {
		return []byte(kp.Cert), nil
	}
	contents, err := ioutil.ReadFile(os.ExpandEnv(kp.CertPath))
	if err != nil {
		return nil, exception.New("error loading cert from path").WithInner(err).WithMessage(kp.CertPath)
	}
	return contents, nil
}

// KeyBytes returns the key pair key bytes.
func (kp KeyPair) KeyBytes() ([]byte, error) {
	if kp.Key != "" {
		return []byte(kp.Key), nil
	}
	contents, err := ioutil.ReadFile(os.ExpandEnv(kp.KeyPath))
	if err != nil {
		return nil, exception.New("error loading key from path").WithInner(err).WithMessage(kp.KeyPath)
	}
	return contents, nil
}

// String returns a string representation of the key pair.
func (kp KeyPair) String() (output string) {
	output = "[ "
	if kp.Cert != "" {
		output = output + "cert: <literal>"
	} else if kp.CertPath != "" {
		output = output + "cert: " + os.ExpandEnv(kp.CertPath)
	}
	if kp.Key != "" {
		output = output + ", key: <literal>"
	} else if kp.KeyPath != "" {
		output = output + ", key: " + os.ExpandEnv(kp.KeyPath)
	}
	output = output + " ]"
	return output
}
