//go:build noyubikey

package yubikey

import (
	"errors"

	"github.com/revelaction/privage/identity"
)

// New returns an error indicating that Yubikey support is disabled.
func New() (identity.Device, error) {
	return nil, errors.New("yubikey support is disabled in this build")
}
