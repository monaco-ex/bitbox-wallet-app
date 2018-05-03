package signing

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/shiftdevices/godbb/util/errp"
)

const (
	hardenedKeySymbol = "'"

	// Hardened denotes a hardened key derivation.
	Hardened = true

	// NonHardened denotes a non-hardened key derivation.
	NonHardened = false
)

// This type is called node because keyNode is marked by the linter as a misspelling.
type keyNode struct {
	index    uint32
	hardened bool
}

func (node keyNode) encode() string {
	suffix := ""
	if node.hardened {
		suffix = hardenedKeySymbol
	}
	return fmt.Sprint(node.index, suffix)
}

type keypath []keyNode

func newKeypath(input string) (keypath, error) {
	splits := strings.Split(input, "/")
	path := make(keypath, 0, len(splits))
	for _, split := range splits {
		split = strings.TrimSpace(split)
		if len(split) == 0 {
			continue
		}
		hardened := strings.HasSuffix(split, hardenedKeySymbol)
		if hardened {
			split = strings.TrimSpace(split[:len(split)-len(hardenedKeySymbol)])
		}
		index, err := strconv.Atoi(split)
		if err != nil {
			return nil, errp.Wrap(err, "A path node is not a number.")
		}
		if index < 0 {
			return nil, errp.New("A path index may not be negative.")
		}
		path = append(path, keyNode{uint32(index), hardened})
	}
	return path, nil
}

func (path keypath) encode() string {
	nodes := make([]string, len(path))
	for index, node := range path {
		nodes[index] = node.encode()
	}
	return strings.Join(nodes, "/")
}

func (path keypath) derive(extendedKey *hdkeychain.ExtendedKey) (*hdkeychain.ExtendedKey, error) {
	for _, node := range path {
		offset := uint32(0)
		if node.hardened {
			offset = hdkeychain.HardenedKeyStart
		}
		var err error
		extendedKey, err = extendedKey.Child(offset + node.index)
		if err != nil {
			return nil, err
		}
	}
	return extendedKey, nil
}

// RelativeKeypath models a relative keypath according to BIP32.
type RelativeKeypath keypath

// NewEmptyRelativeKeypath creates a new empty relative keypath.
func NewEmptyRelativeKeypath() RelativeKeypath {
	return RelativeKeypath{}
}

// NewRelativeKeypath creates a new relative keypath from a string like `1/2'/3`.
func NewRelativeKeypath(input string) (RelativeKeypath, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "m") {
		return nil, errp.New("A relative keypath may not start with 'm'.")
	}
	path, err := newKeypath(input)
	if err != nil {
		return nil, err
	}
	return RelativeKeypath(path), nil
}

// Encode encodes the relative keypath as a string.
func (relativeKeypath RelativeKeypath) Encode() string {
	return keypath(relativeKeypath).encode()
}

// Child appends the given node to this relative keypath.
func (relativeKeypath RelativeKeypath) Child(index uint32, hardened bool) RelativeKeypath {
	return append(relativeKeypath, keyNode{index, hardened})
}

// Hardened returns whether the keypath contains a hardened derivation.
func (relativeKeypath RelativeKeypath) Hardened() bool {
	for _, node := range relativeKeypath {
		if node.hardened {
			return true
		}
	}
	return false
}

// Derive derives the extended key at this path from the given extended key.
func (relativeKeypath RelativeKeypath) Derive(
	extendedKey *hdkeychain.ExtendedKey,
) (*hdkeychain.ExtendedKey, error) {
	return keypath(relativeKeypath).derive(extendedKey)
}

// AbsoluteKeypath models an absolute keypath according to BIP32.
type AbsoluteKeypath keypath

// NewEmptyAbsoluteKeypath creates a new empty absolute keypath.
func NewEmptyAbsoluteKeypath() AbsoluteKeypath {
	return AbsoluteKeypath{}
}

// NewAbsoluteKeypath creates a new absolute keypath from a string like `m/44'/1'`.
func NewAbsoluteKeypath(input string) (AbsoluteKeypath, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "m") {
		return nil, errp.New("An absolute keypath has to start with 'm'.")
	}
	input = input[strings.Index(input, "/")+1:]
	path, err := newKeypath(input)
	if err != nil {
		return nil, err
	}
	return AbsoluteKeypath(path), nil
}

// Encode encodes the absolute keypath as a string.
func (absoluteKeypath AbsoluteKeypath) Encode() string {
	return "m/" + keypath(absoluteKeypath).encode()
}

// Child appends the given node to this absolute keypath.
func (absoluteKeypath AbsoluteKeypath) Child(index uint32, hardened bool) AbsoluteKeypath {
	return append(absoluteKeypath, keyNode{index, hardened})
}

// Append appends a relative keypath to this absolute keypath.
func (absoluteKeypath AbsoluteKeypath) Append(suffix RelativeKeypath) AbsoluteKeypath {
	return append(absoluteKeypath, suffix...)
}

// Derive derives the extended key at this path from the given extended key.
func (absoluteKeypath AbsoluteKeypath) Derive(
	extendedKey *hdkeychain.ExtendedKey,
) (*hdkeychain.ExtendedKey, error) {
	return keypath(absoluteKeypath).derive(extendedKey)
}

// MarshalJSON implements json.Marshaler.
func (absoluteKeypath AbsoluteKeypath) MarshalJSON() ([]byte, error) {
	return json.Marshal(absoluteKeypath.Encode())
}

// UnmarshalJSON implements json.Unmarshaler.
func (absoluteKeypath *AbsoluteKeypath) UnmarshalJSON(bytes []byte) error {
	var input string
	if err := json.Unmarshal(bytes, &input); err != nil {
		return errp.Wrap(err, "Could not unmarshal an absolute keypath.")
	}
	var err error
	*absoluteKeypath, err = NewAbsoluteKeypath(input)
	return err
}