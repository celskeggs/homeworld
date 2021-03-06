package keygen

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/sipb/homeworld/platform/keysystem/keyserver/config"
	"github.com/sipb/homeworld/platform/keysystem/worldconfig"
	"github.com/sipb/homeworld/platform/util/certutil"
)

const AuthorityBits = 4096

func GenerateTLSSelfSignedCert(key *rsa.PrivateKey, name string) ([]byte, error) {
	issueat := time.Now()

	certTemplate := &x509.Certificate{
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,

		NotBefore: issueat,
		NotAfter:  time.Unix(issueat.Unix()+86400*1000000, 0), // one million days in the future

		Subject: pkix.Name{CommonName: "homeworld-authority-" + name},
	}

	return certutil.FinishCertificate(certTemplate, certTemplate, key.Public(), key)
}

func GenerateKeys(dir string) error {
	if info, err := os.Stat(dir); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New("expected authority directory, not authority file")
	}

	for _, authority := range worldconfig.ListAuthorities() {
		privkey, privkeybytes, err := certutil.GenerateRSA(AuthorityBits)
		if err != nil {
			return err
		}
		keyfile, certfile := authority.Filenames()
		err = ioutil.WriteFile(path.Join(dir, keyfile), privkeybytes, os.FileMode(0600))
		if err != nil {
			return err
		}
		switch authority.Type {
		case config.TLSAuthorityType:
			// self-signed cert
			cert, err := GenerateTLSSelfSignedCert(privkey, authority.Name)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(path.Join(dir, certfile), cert, os.FileMode(0644))
			if err != nil {
				return err
			}
		case config.SSHAuthorityType:
			// SSH authorities are just pubkeys
			pkey, err := ssh.NewPublicKey(privkey.Public())
			if err != nil {
				return err
			}
			pubkey := ssh.MarshalAuthorizedKey(pkey)
			err = ioutil.WriteFile(path.Join(dir, certfile), pubkey, os.FileMode(0644))
			if err != nil {
				return err
			}
		default:
			panic("invalid authority type in GenerateKeys")
		}
	}
	return nil
}
