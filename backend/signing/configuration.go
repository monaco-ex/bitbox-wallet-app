package signing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/shiftdevices/godbb/util/errp"
)

// Configuration models a signing configuration, which can be singlesig or multisig.
type Configuration struct {
	absoluteKeypath    AbsoluteKeypath
	extendedPublicKeys []*hdkeychain.ExtendedKey
	signingThreshold   int
}

// NewConfiguration creates a new configuration.
func NewConfiguration(
	absoluteKeypath AbsoluteKeypath,
	extendedPublicKeys []*hdkeychain.ExtendedKey,
	signingThreshold int,
) *Configuration {
	for _, extendedKey := range extendedPublicKeys {
		if extendedKey.IsPrivate() {
			panic("An extended key is private! Only extended public keys are accepted.")
		}
	}
	return &Configuration{
		absoluteKeypath:    absoluteKeypath,
		extendedPublicKeys: extendedPublicKeys,
		signingThreshold:   signingThreshold,
	}
}

// AbsoluteKeypath returns the configuration's keypath.
func (configuration *Configuration) AbsoluteKeypath() AbsoluteKeypath {
	return configuration.absoluteKeypath
}

// ExtendedPublicKeys returns the configuration's extended pubkeys.
func (configuration *Configuration) ExtendedPublicKeys() []*hdkeychain.ExtendedKey {
	return configuration.extendedPublicKeys
}

// SigningThreshold returns the signing threshold in case of a multisig config.
func (configuration *Configuration) SigningThreshold() int {
	return configuration.signingThreshold
}

// NumberOfSigners returns the number of signers (1 in single sig, N in a M/N multisig).
func (configuration *Configuration) NumberOfSigners() int {
	return len(configuration.extendedPublicKeys)
}

// Derive derives a subkeypath from the configuration's base absolute keypath.
func (configuration *Configuration) Derive(relativeKeypath RelativeKeypath) (*Configuration, error) {
	if relativeKeypath.Hardened() {
		return nil, errp.New("A configuration can only be derived with a non-hardened relative keypath.")
	}

	derivedPublicKeys := make([]*hdkeychain.ExtendedKey, configuration.NumberOfSigners())
	for index, extendedPublicKey := range configuration.extendedPublicKeys {
		derivedPublicKey, err := relativeKeypath.Derive(extendedPublicKey)
		if err != nil {
			return nil, err
		}
		derivedPublicKeys[index] = derivedPublicKey
	}
	return &Configuration{
		absoluteKeypath:    configuration.absoluteKeypath.Append(relativeKeypath),
		extendedPublicKeys: derivedPublicKeys,
		signingThreshold:   configuration.signingThreshold,
	}, nil
}

type configurationEncoding struct {
	Keypath   AbsoluteKeypath `json:"keypath"`
	Threshold int             `json:"threshold"`
	Xpubs     []string        `json:"xpubs"`
}

// MarshalJSON implements json.Marshaler.
func (configuration Configuration) MarshalJSON() ([]byte, error) {
	length := configuration.NumberOfSigners()
	xpubs := make([]string, length)
	for i := 0; i < length; i++ {
		xpubs[i] = configuration.extendedPublicKeys[i].String()
	}
	return json.Marshal(&configurationEncoding{
		Keypath:   configuration.absoluteKeypath,
		Threshold: configuration.signingThreshold,
		Xpubs:     xpubs,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (configuration *Configuration) UnmarshalJSON(bytes []byte) error {
	var encoding configurationEncoding
	if err := json.Unmarshal(bytes, &encoding); err != nil {
		return errp.Wrap(err, "Could not unmarshal a signing configuration.")
	}
	configuration.absoluteKeypath = encoding.Keypath
	configuration.signingThreshold = encoding.Threshold
	length := len(encoding.Xpubs)
	configuration.extendedPublicKeys = make([]*hdkeychain.ExtendedKey, length)
	for i := 0; i < length; i++ {
		var err error
		configuration.extendedPublicKeys[i], err = hdkeychain.NewKeyFromString(encoding.Xpubs[i])
		if err != nil {
			return errp.Wrap(err, "Could not read an extended public key.")
		}
	}
	return nil
}

// Hash returns a hash of the configuration.
func (configuration *Configuration) Hash() string {
	bytes, err := json.Marshal(configuration)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}