package header

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	BlockSize          = 512 // bytes
	MaxLenghtCategory  = 40
	MaxLenghtLabel     = 200
	CategoryCredential = "credential"

	maxLenghtVersion = 5
	version          = "v1"
	// ageHeaderPrefix is the magic string that marks the start of an age binary file header.
	//
	// Per the Age specification (https://age-encryption.org/v1):
	// "The textual file header wraps the file key for one or more recipients...
	// It starts with a version line...
	// The version line always starts with "age-encryption.org/", is followed by an arbitrary version string...
	// version-line = %s"age-encryption.org/" version LF"
	ageHeaderPrefix = "age-encryption.org/"
	paddingChar     = ' '
)

// IsCredential returns true if the header belongs to the credential category.
func (h *Header) IsCredential() bool {
	return h.Category == CategoryCredential
}

// Hash generates a deterministic hash of the header and the age identity.
//
// ageIdentity is the string representation of the age public key (recipient).
// This hash is used to generate unique filenames for encrypted content, ensuring
// that the same content encrypted for different identities results in different files.
func (h *Header) Hash(ageIdentity string) (string, error) {
	padded, err := h.Pad()
	if err != nil {
		return "", fmt.Errorf("failed to pad header for hashing: %w", err)
	}

	hashInput := append(padded, []byte(ageIdentity)...)
	sum := sha256.Sum256(hashInput)

	return fmt.Sprintf("%x", sum), nil
}

// Header represents the metadata of an encrypted file.
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

// Pad returns a serialized version of the header using LEFT padding
// to maintain backward compatibility with existing files.
// It returns an error if any field exceeds its maximum allowed byte length.
func (h *Header) Pad() ([]byte, error) {
    buf := new(bytes.Buffer)

    // 1. Version
    vBytes := []byte(version)
    padLen := maxLenghtVersion - len(vBytes)
    if padLen < 0 {
        return nil, fmt.Errorf("version constant exceeds maximum length of %d bytes", maxLenghtVersion)
    }
    buf.Write(bytes.Repeat([]byte{paddingChar}, padLen))
    buf.Write(vBytes)

    // 2. Category
    catBytes := []byte(h.Category)
    padLen = MaxLenghtCategory - len(catBytes)
    if padLen < 0 {
        return nil, fmt.Errorf("category exceeds maximum length of %d bytes", MaxLenghtCategory)
    }
    buf.Write(bytes.Repeat([]byte{paddingChar}, padLen))
    buf.Write(catBytes)

    // 3. Label
    labelBytes := []byte(h.Label)
    padLen = MaxLenghtLabel - len(labelBytes)
    if padLen < 0 {
        return nil, fmt.Errorf("label exceeds maximum length of %d bytes", MaxLenghtLabel)
    }
    buf.Write(bytes.Repeat([]byte{paddingChar}, padLen))
    buf.Write(labelBytes)

    // 4. Safety Check
    if buf.Len() > BlockSize {
        return nil, fmt.Errorf("internal error: padded header size %d exceeds BlockSize %d", buf.Len(), BlockSize)
    }

    return buf.Bytes(), nil
}

// Parse parses a serialized version of a header.
func Parse(h []byte) *Header {
    res := &Header{}

    // Slice strictly by byte offsets
    res.Version = string(bytes.TrimLeft(h[:maxLenghtVersion], string(paddingChar)))

    offset := maxLenghtVersion
    res.Category = string(bytes.TrimLeft(h[offset:offset+MaxLenghtCategory], string(paddingChar)))

    offset += MaxLenghtCategory
    res.Label = string(bytes.TrimLeft(h[offset:], string(paddingChar)))

    return res
}

// PadEncrypted fills the encrypted (with age) header up to BlockSize with paddingChar
// characters
func PadEncrypted(header []byte) ([]byte, error) {
	diff := BlockSize - len(header)
	pad := bytes.Repeat([]byte{paddingChar}, diff)
	padded := append(pad, header...)
	return padded, nil
}

// Unpad removes the filled characters of a encrypted header.
// It strictly verifies that all bytes preceding the age header prefix are valid padding characters.
func Unpad(header []byte) ([]byte, error) {

	idx := bytes.Index(header, []byte(ageHeaderPrefix))
	if idx == -1 {
		return nil, errors.New("could not unpad header, age prefix not found")
	}

	// Integrity check: Ensure skipped bytes are valid padding
	padding := header[:idx]
	if len(bytes.Trim(padding, string(paddingChar))) > 0 {
		return nil, errors.New("header corruption: non-padding bytes found before age prefix")
	}

	return header[idx:], nil
}
