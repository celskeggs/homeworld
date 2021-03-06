package keygen

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path"

	"github.com/sipb/homeworld/platform/keysystem/keyclient/actloop"
	"github.com/sipb/homeworld/platform/util/certutil"
	"github.com/sipb/homeworld/platform/util/fileutil"
)

const DefaultRSAKeyLength = 4096

func GenerateKey(keypath string, nac *actloop.NewActionContext) {
	if fileutil.Exists(keypath) {
		// nothing to do
	} else {
		// it's acceptable for the directory to not exist, because we'll just create it later
		info := fmt.Sprintf("generate key %s", keypath)
		err := keygen(keypath)
		if err != nil {
			nac.Errored(info, err)
		} else {
			nac.NotifyPerformed(info)
		}
	}
}

func keygen(keypath string) error {
	dirname := path.Dir(keypath)
	err := fileutil.EnsureIsFolder(dirname)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to prepare directory %s for generated key", dirname))
	}

	_, privateKeyPEM, err := certutil.GenerateRSA(DefaultRSAKeyLength)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to generate %d-bit RSA key for %s", DefaultRSAKeyLength, keypath))
	}
	err = fileutil.CreateFile(keypath, privateKeyPEM, os.FileMode(0600))
	if err != nil {
		return errors.Wrap(err, "failed to create file for generated key")
	}
	return nil
}
