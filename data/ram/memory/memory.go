package memory

import (
	"context"
	"sync"
	"time"

	"github.com/code-payments/ocp-server/solana/vm"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type store struct {
	mu      sync.Mutex
	records []*ram.Record
	last    uint64
}

// New returns a new in memory-backed ram.Store
func New() ram.Store {
	return &store{}
}

// Save implements ram.Store.Save
func (s *store) Save(_ context.Context, data *ram.Record) error {
	if err := data.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.last++

	if item := s.find(data); item != nil {
		if item.Slot >= data.Slot {
			return ram.ErrStaleState
		}

		item.IsAllocated = data.IsAllocated
		item.Address = data.Address
		item.Type = data.Type
		item.Data = data.Data
		item.Slot = data.Slot
		item.IsSlotAdvanced = data.IsSlotAdvanced
		item.LastUpdatedAt = time.Now()

		item.CopyTo(data)
	} else {
		data.Id = s.last
		data.LastUpdatedAt = time.Now()

		cloned := data.Clone()
		s.records = append(s.records, &cloned)
	}

	return nil
}

// GetAllMemoryAccounts implements ram.Store.GetAllMemoryAccounts
func (s *store) GetAllMemoryAccounts(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.records) == 0 {
		return nil, ram.ErrAccountNotFound
	}

	uniqueMemoryAccounts := make(map[string]any)
	for _, item := range s.records {
		uniqueMemoryAccounts[item.MemoryAccount] = struct{}{}
	}

	var res []string
	for memoryAccount := range uniqueMemoryAccounts {
		res = append(res, memoryAccount)
	}

	return res, nil
}

// GetAllByMemoryAccount implements ram.Store.GetAllByMemoryAccount
func (s *store) GetAllByMemoryAccount(_ context.Context, memoryAccount string) ([]*ram.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.findByMemoryAccount(memoryAccount)
	if len(items) == 0 {
		return nil, ram.ErrItemNotFound
	}
	return cloneRecords(items), nil
}

// GetAllVirtualAccountsByAddressAndType implements ram.Store.GetAllVirtualAccountsByAddressAndType
func (s *store) GetAllVirtualAccountsByAddressAndType(_ context.Context, vm, address string, accountType vm.VirtualAccountType) ([]*ram.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.findByVmAddressAndAccountType(vm, address, accountType)
	if len(items) == 0 {
		return nil, ram.ErrItemNotFound
	}
	return cloneRecords(items), nil
}

func (s *store) find(data *ram.Record) *ram.Record {
	for _, item := range s.records {
		if item.Id == data.Id {
			return item
		}

		if data.MemoryAccount == item.MemoryAccount && data.Index == item.Index {
			return item
		}
	}
	return nil
}

func (s *store) findByMemoryAccount(memoryAccount string) []*ram.Record {
	var res []*ram.Record
	for _, item := range s.records {
		if item.MemoryAccount == memoryAccount {
			res = append(res, item)
		}
	}
	return res
}

func (s *store) findByVmAddressAndAccountType(vm, address string, accountType vm.VirtualAccountType) []*ram.Record {
	var res []*ram.Record
	for _, item := range s.records {
		if !item.IsAllocated {
			continue
		}

		if item.Vm == vm && *item.Address == address && *item.Type == accountType {
			res = append(res, item)
		}
	}
	return res
}

func cloneRecords(items []*ram.Record) []*ram.Record {
	res := make([]*ram.Record, len(items))
	for i, item := range items {
		cloned := item.Clone()
		res[i] = &cloned
	}
	return res
}

func (s *store) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = nil
	s.last = 0
}
