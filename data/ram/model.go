package ram

import (
	"errors"
	"time"

	"github.com/code-payments/code-server/pkg/solana/cvm"
)

type Record struct {
	Id uint64

	Vm string

	MemoryAccount string
	Index         uint16
	IsAllocated   bool

	Address *string
	Type    *cvm.VirtualAccountType
	Data    []byte

	Slot uint64

	LastUpdatedAt time.Time
}

func (r *Record) Validate() error {
	if len(r.Vm) == 0 {
		return errors.New("vm is required")
	}

	if len(r.MemoryAccount) == 0 {
		return errors.New("memory account is required")
	}

	if r.IsAllocated {
		if r.Address == nil || len(*r.Address) == 0 {
			return errors.New("address is required for allocated memory")
		}

		if r.Type == nil {
			return errors.New("memory item type is required for allocated memory")
		}
		switch *r.Type {
		case cvm.VirtualAccountTypeDurableNonce, cvm.VirtualAccountTypeTimelock, cvm.VirtualAccountTypeRelay:
		default:
			return errors.New("invalid memory item type")
		}

		if len(r.Data) == 0 {
			return errors.New("data is required for allocated memory")
		}
	} else {
		if r.Address != nil {
			return errors.New("address cannot be set for unallocated memory")
		}

		if r.Type != nil {
			return errors.New("memory item type cannot be set for unallocated memory")
		}

		if len(r.Data) > 0 {
			return errors.New("data cannot be set for unallocated memory")
		}
	}

	if r.Slot == 0 {
		return errors.New("slot is required")
	}

	return nil
}

func (r *Record) Clone() Record {
	return Record{
		Id: r.Id,

		Vm: r.Vm,

		MemoryAccount: r.MemoryAccount,
		Index:         r.Index,
		IsAllocated:   r.IsAllocated,

		Address: r.Address,
		Type:    r.Type,
		Data:    r.Data,

		Slot: r.Slot,

		LastUpdatedAt: r.LastUpdatedAt,
	}
}

func (r *Record) CopyTo(dst *Record) {
	dst.Id = r.Id

	dst.Vm = r.Vm

	dst.MemoryAccount = r.MemoryAccount
	dst.Index = r.Index
	dst.IsAllocated = r.IsAllocated

	dst.Address = r.Address
	dst.Type = r.Type
	dst.Data = r.Data

	dst.Slot = r.Slot

	dst.LastUpdatedAt = r.LastUpdatedAt
}

func (r *Record) ToVirtualDurableNonce() (*cvm.VirtualDurableNonce, bool) {
	if !r.IsAllocated || *r.Type != cvm.VirtualAccountTypeDurableNonce {
		return nil, false
	}

	var vdn cvm.VirtualDurableNonce
	err := vdn.UnmarshalDirectly(r.Data)
	if err != nil {
		return nil, false
	}
	return &vdn, true
}

func (r *Record) ToVirtualTimelockAccount() (*cvm.VirtualTimelockAccount, bool) {
	if !r.IsAllocated || *r.Type != cvm.VirtualAccountTypeTimelock {
		return nil, false
	}

	var vta cvm.VirtualTimelockAccount
	err := vta.UnmarshalDirectly(r.Data)
	if err != nil {
		return nil, false
	}
	return &vta, true
}

func (r *Record) ToVirtualRelayAccount() (*cvm.VirtualRelayAccount, bool) {
	if !r.IsAllocated || *r.Type != cvm.VirtualAccountTypeRelay {
		return nil, false
	}

	var vra cvm.VirtualRelayAccount
	err := vra.UnmarshalDirectly(r.Data)
	if err != nil {
		return nil, false
	}
	return &vra, true
}
