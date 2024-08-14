// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: indexer/model.proto

package indexer

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = ptypes.DynamicAny{}
)

// Validate checks the field values on Address with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Address) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetValue()) != 32 {
		return AddressValidationError{
			field:  "Value",
			reason: "value length must be 32 bytes",
		}
	}

	return nil
}

// AddressValidationError is the validation error returned by Address.Validate
// if the designated constraints aren't met.
type AddressValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AddressValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AddressValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AddressValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AddressValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AddressValidationError) ErrorName() string { return "AddressValidationError" }

// Error satisfies the builtin error interface
func (e AddressValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAddress.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AddressValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AddressValidationError{}

// Validate checks the field values on Hash with the rules defined in the proto
// definition for this message. If any rules are violated, an error is returned.
func (m *Hash) Validate() error {
	if m == nil {
		return nil
	}

	if len(m.GetValue()) != 32 {
		return HashValidationError{
			field:  "Value",
			reason: "value length must be 32 bytes",
		}
	}

	return nil
}

// HashValidationError is the validation error returned by Hash.Validate if the
// designated constraints aren't met.
type HashValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e HashValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e HashValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e HashValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e HashValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e HashValidationError) ErrorName() string { return "HashValidationError" }

// Error satisfies the builtin error interface
func (e HashValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sHash.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = HashValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = HashValidationError{}

// Validate checks the field values on Proof with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Proof) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetRoot() == nil {
		return ProofValidationError{
			field:  "Root",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetRoot()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ProofValidationError{
				field:  "Root",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if l := len(m.GetProof()); l < 1 || l > 63 {
		return ProofValidationError{
			field:  "Proof",
			reason: "value must contain between 1 and 63 items, inclusive",
		}
	}

	for idx, item := range m.GetProof() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ProofValidationError{
					field:  fmt.Sprintf("Proof[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if m.GetLeaf() == nil {
		return ProofValidationError{
			field:  "Leaf",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetLeaf()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ProofValidationError{
				field:  "Leaf",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for ForLeafIndex

	// no validation rules for UntilLeafIndex

	return nil
}

// ProofValidationError is the validation error returned by Proof.Validate if
// the designated constraints aren't met.
type ProofValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ProofValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ProofValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ProofValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ProofValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ProofValidationError) ErrorName() string { return "ProofValidationError" }

// Error satisfies the builtin error interface
func (e ProofValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sProof.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ProofValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ProofValidationError{}

// Validate checks the field values on VirtualTimelockAccount with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *VirtualTimelockAccount) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetOwner() == nil {
		return VirtualTimelockAccountValidationError{
			field:  "Owner",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetOwner()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualTimelockAccountValidationError{
				field:  "Owner",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetNonce() == nil {
		return VirtualTimelockAccountValidationError{
			field:  "Nonce",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetNonce()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualTimelockAccountValidationError{
				field:  "Nonce",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetTokenBump() > 255 {
		return VirtualTimelockAccountValidationError{
			field:  "TokenBump",
			reason: "value must be less than or equal to 255",
		}
	}

	if m.GetUnlockBump() > 255 {
		return VirtualTimelockAccountValidationError{
			field:  "UnlockBump",
			reason: "value must be less than or equal to 255",
		}
	}

	if m.GetWithdrawBump() > 255 {
		return VirtualTimelockAccountValidationError{
			field:  "WithdrawBump",
			reason: "value must be less than or equal to 255",
		}
	}

	// no validation rules for Balance

	if m.GetBump() > 255 {
		return VirtualTimelockAccountValidationError{
			field:  "Bump",
			reason: "value must be less than or equal to 255",
		}
	}

	return nil
}

// VirtualTimelockAccountValidationError is the validation error returned by
// VirtualTimelockAccount.Validate if the designated constraints aren't met.
type VirtualTimelockAccountValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualTimelockAccountValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualTimelockAccountValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualTimelockAccountValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualTimelockAccountValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualTimelockAccountValidationError) ErrorName() string {
	return "VirtualTimelockAccountValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualTimelockAccountValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualTimelockAccount.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualTimelockAccountValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualTimelockAccountValidationError{}

// Validate checks the field values on VirtualDurableNonce with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *VirtualDurableNonce) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAddress() == nil {
		return VirtualDurableNonceValidationError{
			field:  "Address",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAddress()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualDurableNonceValidationError{
				field:  "Address",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetNonce() == nil {
		return VirtualDurableNonceValidationError{
			field:  "Nonce",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetNonce()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualDurableNonceValidationError{
				field:  "Nonce",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// VirtualDurableNonceValidationError is the validation error returned by
// VirtualDurableNonce.Validate if the designated constraints aren't met.
type VirtualDurableNonceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualDurableNonceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualDurableNonceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualDurableNonceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualDurableNonceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualDurableNonceValidationError) ErrorName() string {
	return "VirtualDurableNonceValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualDurableNonceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualDurableNonce.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualDurableNonceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualDurableNonceValidationError{}

// Validate checks the field values on VirtualRelayAccount with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *VirtualRelayAccount) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAddress() == nil {
		return VirtualRelayAccountValidationError{
			field:  "Address",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAddress()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualRelayAccountValidationError{
				field:  "Address",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetCommitment() == nil {
		return VirtualRelayAccountValidationError{
			field:  "Commitment",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetCommitment()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualRelayAccountValidationError{
				field:  "Commitment",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetRecentRoot() == nil {
		return VirtualRelayAccountValidationError{
			field:  "RecentRoot",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetRecentRoot()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualRelayAccountValidationError{
				field:  "RecentRoot",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetDestination() == nil {
		return VirtualRelayAccountValidationError{
			field:  "Destination",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetDestination()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualRelayAccountValidationError{
				field:  "Destination",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// VirtualRelayAccountValidationError is the validation error returned by
// VirtualRelayAccount.Validate if the designated constraints aren't met.
type VirtualRelayAccountValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualRelayAccountValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualRelayAccountValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualRelayAccountValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualRelayAccountValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualRelayAccountValidationError) ErrorName() string {
	return "VirtualRelayAccountValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualRelayAccountValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualRelayAccount.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualRelayAccountValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualRelayAccountValidationError{}

// Validate checks the field values on VirtualAccountStorage with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *VirtualAccountStorage) Validate() error {
	if m == nil {
		return nil
	}

	switch m.Storage.(type) {

	case *VirtualAccountStorage_Memory:

		if v, ok := interface{}(m.GetMemory()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return VirtualAccountStorageValidationError{
					field:  "Memory",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *VirtualAccountStorage_Compressed:

		if v, ok := interface{}(m.GetCompressed()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return VirtualAccountStorageValidationError{
					field:  "Compressed",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	default:
		return VirtualAccountStorageValidationError{
			field:  "Storage",
			reason: "value is required",
		}

	}

	return nil
}

// VirtualAccountStorageValidationError is the validation error returned by
// VirtualAccountStorage.Validate if the designated constraints aren't met.
type VirtualAccountStorageValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualAccountStorageValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualAccountStorageValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualAccountStorageValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualAccountStorageValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualAccountStorageValidationError) ErrorName() string {
	return "VirtualAccountStorageValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualAccountStorageValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualAccountStorage.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualAccountStorageValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualAccountStorageValidationError{}

// Validate checks the field values on MemoryVirtualAccountStorage with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *MemoryVirtualAccountStorage) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAccount() == nil {
		return MemoryVirtualAccountStorageValidationError{
			field:  "Account",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAccount()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return MemoryVirtualAccountStorageValidationError{
				field:  "Account",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetIndex() > 65535 {
		return MemoryVirtualAccountStorageValidationError{
			field:  "Index",
			reason: "value must be less than or equal to 65535",
		}
	}

	return nil
}

// MemoryVirtualAccountStorageValidationError is the validation error returned
// by MemoryVirtualAccountStorage.Validate if the designated constraints
// aren't met.
type MemoryVirtualAccountStorageValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MemoryVirtualAccountStorageValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MemoryVirtualAccountStorageValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MemoryVirtualAccountStorageValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MemoryVirtualAccountStorageValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MemoryVirtualAccountStorageValidationError) ErrorName() string {
	return "MemoryVirtualAccountStorageValidationError"
}

// Error satisfies the builtin error interface
func (e MemoryVirtualAccountStorageValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMemoryVirtualAccountStorage.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MemoryVirtualAccountStorageValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MemoryVirtualAccountStorageValidationError{}

// Validate checks the field values on CompressedVirtualAccountStorage with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *CompressedVirtualAccountStorage) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAccount() == nil {
		return CompressedVirtualAccountStorageValidationError{
			field:  "Account",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAccount()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return CompressedVirtualAccountStorageValidationError{
				field:  "Account",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetProof() == nil {
		return CompressedVirtualAccountStorageValidationError{
			field:  "Proof",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetProof()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return CompressedVirtualAccountStorageValidationError{
				field:  "Proof",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// CompressedVirtualAccountStorageValidationError is the validation error
// returned by CompressedVirtualAccountStorage.Validate if the designated
// constraints aren't met.
type CompressedVirtualAccountStorageValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CompressedVirtualAccountStorageValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CompressedVirtualAccountStorageValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CompressedVirtualAccountStorageValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CompressedVirtualAccountStorageValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CompressedVirtualAccountStorageValidationError) ErrorName() string {
	return "CompressedVirtualAccountStorageValidationError"
}

// Error satisfies the builtin error interface
func (e CompressedVirtualAccountStorageValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCompressedVirtualAccountStorage.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CompressedVirtualAccountStorageValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CompressedVirtualAccountStorageValidationError{}

// Validate checks the field values on
// VirtualTimelockAccountWithStorageMetadata with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *VirtualTimelockAccountWithStorageMetadata) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAccount() == nil {
		return VirtualTimelockAccountWithStorageMetadataValidationError{
			field:  "Account",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAccount()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualTimelockAccountWithStorageMetadataValidationError{
				field:  "Account",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetStorage() == nil {
		return VirtualTimelockAccountWithStorageMetadataValidationError{
			field:  "Storage",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetStorage()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualTimelockAccountWithStorageMetadataValidationError{
				field:  "Storage",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Slot

	return nil
}

// VirtualTimelockAccountWithStorageMetadataValidationError is the validation
// error returned by VirtualTimelockAccountWithStorageMetadata.Validate if the
// designated constraints aren't met.
type VirtualTimelockAccountWithStorageMetadataValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualTimelockAccountWithStorageMetadataValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualTimelockAccountWithStorageMetadataValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualTimelockAccountWithStorageMetadataValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualTimelockAccountWithStorageMetadataValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualTimelockAccountWithStorageMetadataValidationError) ErrorName() string {
	return "VirtualTimelockAccountWithStorageMetadataValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualTimelockAccountWithStorageMetadataValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualTimelockAccountWithStorageMetadata.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualTimelockAccountWithStorageMetadataValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualTimelockAccountWithStorageMetadataValidationError{}

// Validate checks the field values on VirtualDurableNonceWithStorageMetadata
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *VirtualDurableNonceWithStorageMetadata) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAccount() == nil {
		return VirtualDurableNonceWithStorageMetadataValidationError{
			field:  "Account",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAccount()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualDurableNonceWithStorageMetadataValidationError{
				field:  "Account",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetStorage() == nil {
		return VirtualDurableNonceWithStorageMetadataValidationError{
			field:  "Storage",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetStorage()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualDurableNonceWithStorageMetadataValidationError{
				field:  "Storage",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Slot

	return nil
}

// VirtualDurableNonceWithStorageMetadataValidationError is the validation
// error returned by VirtualDurableNonceWithStorageMetadata.Validate if the
// designated constraints aren't met.
type VirtualDurableNonceWithStorageMetadataValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualDurableNonceWithStorageMetadataValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualDurableNonceWithStorageMetadataValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualDurableNonceWithStorageMetadataValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualDurableNonceWithStorageMetadataValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualDurableNonceWithStorageMetadataValidationError) ErrorName() string {
	return "VirtualDurableNonceWithStorageMetadataValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualDurableNonceWithStorageMetadataValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualDurableNonceWithStorageMetadata.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualDurableNonceWithStorageMetadataValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualDurableNonceWithStorageMetadataValidationError{}

// Validate checks the field values on VirtualRelayAccountWithStorageMetadata
// with the rules defined in the proto definition for this message. If any
// rules are violated, an error is returned.
func (m *VirtualRelayAccountWithStorageMetadata) Validate() error {
	if m == nil {
		return nil
	}

	if m.GetAccount() == nil {
		return VirtualRelayAccountWithStorageMetadataValidationError{
			field:  "Account",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetAccount()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualRelayAccountWithStorageMetadataValidationError{
				field:  "Account",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if m.GetStorage() == nil {
		return VirtualRelayAccountWithStorageMetadataValidationError{
			field:  "Storage",
			reason: "value is required",
		}
	}

	if v, ok := interface{}(m.GetStorage()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return VirtualRelayAccountWithStorageMetadataValidationError{
				field:  "Storage",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Slot

	return nil
}

// VirtualRelayAccountWithStorageMetadataValidationError is the validation
// error returned by VirtualRelayAccountWithStorageMetadata.Validate if the
// designated constraints aren't met.
type VirtualRelayAccountWithStorageMetadataValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e VirtualRelayAccountWithStorageMetadataValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e VirtualRelayAccountWithStorageMetadataValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e VirtualRelayAccountWithStorageMetadataValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e VirtualRelayAccountWithStorageMetadataValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e VirtualRelayAccountWithStorageMetadataValidationError) ErrorName() string {
	return "VirtualRelayAccountWithStorageMetadataValidationError"
}

// Error satisfies the builtin error interface
func (e VirtualRelayAccountWithStorageMetadataValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sVirtualRelayAccountWithStorageMetadata.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = VirtualRelayAccountWithStorageMetadataValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = VirtualRelayAccountWithStorageMetadataValidationError{}
