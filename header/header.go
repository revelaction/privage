package header

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

const (
	BlockSize          = 512 // bytes
	MaxLenghtCategory  = 40
	MaxLenghtLabel     = 200
	CategoryCredential = "credential"

	maxLenghtVersion = 5
	version          = "v1"
	ageHeaderPrefix  = "age-encryption.org/"
)

// A Header contains the filename and a category (the metadata) of a file.
//
// This metadata is serialized and encrypted at the start of a privage file.
type Header struct {
	Version  string
	Category string
	Label    string

	// Path of the privage file containing the header  
	Path string

	// Error when unpadding or decrypting the header of the  file
	Err error
}

func (h *Header) String() string {

	if h.Category == CategoryCredential {
		return fmt.Sprintf("📝 %s  🔖%s", h.Label, h.Category)
	}

	return fmt.Sprintf("💼 %s  🔖%s", h.Label, h.Category)
}

// Pad returns a serialized version of the header: string concatenation with
// padding
func (h *Header) Pad() []byte {
	versionFormatStr := "%" + strconv.Itoa(maxLenghtVersion) + "s" // "%20s"
	versionStr := fmt.Sprintf(versionFormatStr, version)
	catFormatStr := "%" + strconv.Itoa(MaxLenghtCategory) + "s" // "%20s"
	categoryStr := fmt.Sprintf(catFormatStr, h.Category)
	labelFormatStr := "%" + strconv.Itoa(MaxLenghtLabel) + "s" // "%50s"
	labelStr := fmt.Sprintf(labelFormatStr, h.Label)
	return []byte(versionStr + categoryStr + labelStr)
}

// Parse parses a serialized version of a header and creates a Header struct 
func Parse(h []byte) *Header {

	res := &Header{}
	res.Version = string(bytes.TrimLeft(h[:maxLenghtVersion], " "))
	res.Category = string(bytes.TrimLeft(h[maxLenghtVersion:maxLenghtVersion+MaxLenghtCategory], " "))
	res.Label = string(bytes.TrimLeft(h[maxLenghtVersion+MaxLenghtCategory:], " "))
	return res
}

// PadEncrypted fills the encrypted (with age) header up to BlockSize with 0x20
// characters
func PadEncrypted(header []byte) ([]byte, error) {
	diff := BlockSize - len(header)
	pad := bytes.Repeat([]byte{0x20}, diff)
	padded := append(pad, header...)
	return padded, nil
}

// Unpad removes the filled characters of a encrypted header.
func Unpad(header []byte) ([]byte, error) {

	idx := bytes.Index(header, []byte(ageHeaderPrefix))
	if -1 == idx {
		return nil, errors.New("Could not unpad header, age prefix not found.")
	}

	return header[idx:], nil
}
