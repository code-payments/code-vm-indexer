package geyser

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"sync"
	"time"

	"github.com/code-payments/ocp-server/retry"
	"github.com/code-payments/ocp-server/solana"
	"github.com/code-payments/ocp-server/solana/vm"
	"github.com/mr-tron/base58"
	"go.uber.org/zap"

	geyserpb "github.com/code-payments/code-vm-indexer/generated/geyser/v1"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

const (
	confirmationsToFinalization = 32
)

type cachedVirtualAccount struct {
	IsInitialized  bool
	Index          int
	Type           vm.VirtualAccountType
	Address        string
	State          []byte
	Slot           uint64
	IsSlotAdvanced bool
}

type MemoryAccountWithDataUpdateHandler struct {
	log *zap.Logger

	solanaClient solana.Client

	ramStore ram.Store

	observableVmAccounts map[string]any

	cachedMemoryAccountStateMu sync.RWMutex
	cachedMemoryAccountState   map[string]map[int]*cachedVirtualAccount

	lastSuccessfulSlotUpdateMu sync.RWMutex
	lastSuccessfulSlotUpdate   map[string]uint64

	highestQueuedSlotUpdateMu sync.RWMutex
	highestQueuedSlotUpdate   map[string]uint64

	backupWorkerInterval time.Duration
}

// NewMemoryAccountWithDataUpdateHandler returns a new ProgramAccountUpdateHandler
// for observing and persisting state changes to a MemoryAccountWithData account
func NewMemoryAccountWithDataUpdateHandler(log *zap.Logger, solanaClient solana.Client, ramStore ram.Store, backupWorkerInterval time.Duration, vmsToObserve ...string) ProgramAccountUpdateHandler {
	observableVmAccounts := make(map[string]any)
	for _, vm := range vmsToObserve {
		observableVmAccounts[vm] = struct{}{}
	}

	return &MemoryAccountWithDataUpdateHandler{
		log:                      log,
		solanaClient:             solanaClient,
		ramStore:                 ramStore,
		observableVmAccounts:     observableVmAccounts,
		cachedMemoryAccountState: make(map[string]map[int]*cachedVirtualAccount),
		lastSuccessfulSlotUpdate: make(map[string]uint64),
		highestQueuedSlotUpdate:  make(map[string]uint64),
		backupWorkerInterval:     backupWorkerInterval,
	}
}

// Handle implements geyser.ProgramAccountUpdateHandler.Handle
func (h *MemoryAccountWithDataUpdateHandler) Handle(ctx context.Context, update *geyserpb.SubscribeUpdateAccount) error {
	// Simply fetch finalized account state as the safest option
	var finalizedData []byte
	var finalizedSlot uint64
	var err error
	_, err = retry.Retry(
		func() error {
			finalizedData, finalizedSlot, err = h.solanaClient.GetAccountDataAfterBlock(update.Account.Pubkey, update.Slot)
			return err
		},
		waitForFinalizationRetryStrategies...,
	)
	if err != nil {
		return err
	}

	var state vm.MemoryAccountWithData
	if err := state.Unmarshal(finalizedData); err != nil {
		return nil
	}
	return h.onStateObserved(ctx, update.Account.Pubkey, finalizedSlot, &state)
}

// RunBackupWorker implements geyser.ProgramAccountUpdateHandler.RunBackupWorker
func (h *MemoryAccountWithDataUpdateHandler) RunBackupWorker(ctx context.Context) error {
	return h.backupWorker(ctx)
}

func (h *MemoryAccountWithDataUpdateHandler) backupWorker(ctx context.Context) error {
	log := h.log.With(zap.String("method", "backupWorker"))

	for {
		select {
		case <-time.After(h.backupWorkerInterval):
			var addresses []string
			for vm := range h.observableVmAccounts {
				log := log.With(zap.String("vm", vm))

				addressesByVm, err := h.ramStore.GetAllMemoryAccounts(ctx, vm)
				switch err {
				case nil:
					addresses = append(addresses, addressesByVm...)
				case ram.ErrAccountNotFound:
				default:
					log.Warn("failure getting memory account addresses by vm", zap.Error(err))
				}
			}

			for _, address := range addresses {
				log := log.With(zap.String("address", address))

				decodedAddress, err := base58.Decode(address)
				if err != nil {
					log.Warn("address is invalid", zap.Error(err))
					continue
				}

				h.lastSuccessfulSlotUpdateMu.RLock()
				minSlot := h.lastSuccessfulSlotUpdate[address]
				h.lastSuccessfulSlotUpdateMu.RUnlock()

				data, observedAtSlot, err := h.solanaClient.GetAccountDataAfterBlock(decodedAddress, minSlot)
				if err != nil {
					log.Warn("failure getting account data", zap.Error(err))
					continue
				}

				var state vm.MemoryAccountWithData
				if err := state.Unmarshal(data); err != nil {
					log.Warn("invalid account data state", zap.Error(err))
					continue
				}

				err = h.onStateObserved(ctx, decodedAddress, observedAtSlot, &state)
				if err != nil {
					continue
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *MemoryAccountWithDataUpdateHandler) onStateObserved(ctx context.Context, address ed25519.PublicKey, observedAtSlot uint64, state *vm.MemoryAccountWithData) error {
	log := h.log.With(zap.String("method", "onStateObserved"))

	base58VmAddress := base58.Encode(state.Vm)
	base58MemoryAccountAddress := base58.Encode(address)

	log = log.With(
		zap.String("vm", base58VmAddress),
		zap.String("address", base58MemoryAccountAddress),
	)

	// Not a VM that is being observed
	if _, ok := h.observableVmAccounts[base58VmAddress]; !ok {
		return nil
	}

	// Check if the state is stale relative to the last successful update
	h.lastSuccessfulSlotUpdateMu.RLock()
	lastSuccessfulSlotUpdate := h.lastSuccessfulSlotUpdate[base58MemoryAccountAddress]
	if observedAtSlot <= lastSuccessfulSlotUpdate {
		h.lastSuccessfulSlotUpdateMu.RUnlock()
		return nil
	}
	h.lastSuccessfulSlotUpdateMu.RUnlock()

	// Check if there's an update for a higher slot
	h.highestQueuedSlotUpdateMu.Lock()
	highestQueuedSlotUpdate := h.highestQueuedSlotUpdate[base58MemoryAccountAddress]
	if observedAtSlot <= highestQueuedSlotUpdate {
		h.highestQueuedSlotUpdateMu.Unlock()
		return nil
	}
	h.highestQueuedSlotUpdate[base58MemoryAccountAddress] = observedAtSlot
	h.highestQueuedSlotUpdateMu.Unlock()

	h.cachedMemoryAccountStateMu.Lock()
	defer h.cachedMemoryAccountStateMu.Unlock()

	// Check (again after acquiring cached memory state mutex) if the state
	// is stale relative to the last successful update
	h.lastSuccessfulSlotUpdateMu.RLock()
	lastSuccessfulSlotUpdate = h.lastSuccessfulSlotUpdate[base58MemoryAccountAddress]
	if observedAtSlot <= lastSuccessfulSlotUpdate {
		h.lastSuccessfulSlotUpdateMu.RUnlock()
		return nil
	}
	h.lastSuccessfulSlotUpdateMu.RUnlock()

	// Check (again after acquiring cached memory state mutex) if there's an
	// update queued for a higher slot
	h.highestQueuedSlotUpdateMu.RLock()
	highestQueuedSlotUpdate = h.highestQueuedSlotUpdate[base58MemoryAccountAddress]
	if observedAtSlot < highestQueuedSlotUpdate {
		h.highestQueuedSlotUpdateMu.RUnlock()
		return nil
	}
	h.highestQueuedSlotUpdateMu.RUnlock()

	cachedState, ok := h.cachedMemoryAccountState[base58MemoryAccountAddress]
	if !ok {
		// Load entire memory account state from the DB into the cache
		cachedState = make(map[int]*cachedVirtualAccount)

		records, err := h.ramStore.GetAllByMemoryAccount(ctx, base58MemoryAccountAddress)
		switch err {
		case nil:
			for _, record := range records {
				cachedVirtualAccountState := &cachedVirtualAccount{
					IsInitialized: record.IsAllocated,
					Index:         int(record.Index),
					Slot:          record.Slot,
				}

				if cachedVirtualAccountState.IsInitialized {
					cachedVirtualAccountState.Type = *record.Type
					cachedVirtualAccountState.Address = *record.Address
					cachedVirtualAccountState.State = record.Data
				}

				cachedState[cachedVirtualAccountState.Index] = cachedVirtualAccountState
			}
		case ram.ErrItemNotFound:
		default:
			log.Warn("failure loading memory account state from db", zap.Error(err))
			return err
		}

		h.cachedMemoryAccountState[base58MemoryAccountAddress] = cachedState
	}

	// Track delta changes to the memory account state to be persisted into the DB
	var dbUpdates []*cachedVirtualAccount
	for index := range state.Data.State {
		log := log.With(zap.Int("index", index))

		cachedVirtualAccountState, ok := cachedState[index]
		if !ok {
			cachedVirtualAccountState = &cachedVirtualAccount{
				IsInitialized: false,
				Index:         index,
			}
		}

		if cachedVirtualAccountState.Slot >= observedAtSlot {
			continue
		}

		isInitialized := state.Data.IsAllocated(index)

		var base58VirtualAccountAddress string
		var newVirtualAccountState []byte
		var virtualAccountType vm.VirtualAccountType
		if isInitialized {
			newVirtualAccountState, _ = state.Data.Read(index)
			virtualAccountType = vm.VirtualAccountType(newVirtualAccountState[0])
			newVirtualAccountState = newVirtualAccountState[1:]

			switch virtualAccountType {
			case vm.VirtualAccountTypeDurableNonce:
				var virtualAccountState vm.VirtualDurableNonce
				if err := virtualAccountState.UnmarshalDirectly(newVirtualAccountState); err != nil {
					log.Warn("failure unmarshalling virtual durable nonce", zap.Error(err))
					return err
				}
				base58VirtualAccountAddress = base58.Encode(virtualAccountState.Address)
			case vm.VirtualAccountTypeTimelock:
				var virtualAccountState vm.VirtualTimelockAccount
				if err := virtualAccountState.UnmarshalDirectly(newVirtualAccountState); err != nil {
					log.Warn("failure unmarshalling virtual timelock account", zap.Error(err))
					return err
				}
				base58VirtualAccountAddress = base58.Encode(virtualAccountState.Owner)
			default:
				// Changelog item, which isn't being tracked in the virtual accounts DB,
				// so treat it as an unititialized memory item
				isInitialized = false
			}
		}

		var dbUpdate *cachedVirtualAccount
		if isInitialized {
			var isSlotAdvanced bool
			if cachedVirtualAccountState.IsInitialized &&
				cachedVirtualAccountState.Address == base58VirtualAccountAddress &&
				bytes.Equal(cachedVirtualAccountState.State, newVirtualAccountState) {

				if cachedVirtualAccountState.IsSlotAdvanced || observedAtSlot-cachedVirtualAccountState.Slot < 2*confirmationsToFinalization {
					continue
				}

				// Advance the slot sufficiently far past finalization if it hasn't
				// already been advanced. This is necessary for systems that rely on
				// finalized states via the RPC service.
				isSlotAdvanced = true
			}

			dbUpdate = &cachedVirtualAccount{
				IsInitialized:  true,
				Index:          index,
				Type:           virtualAccountType,
				Address:        base58VirtualAccountAddress,
				State:          newVirtualAccountState,
				IsSlotAdvanced: isSlotAdvanced,
			}
		} else {
			if !cachedVirtualAccountState.IsInitialized {
				continue
			}

			dbUpdate = &cachedVirtualAccount{
				IsInitialized: false,
				Index:         index,
			}
		}

		dbUpdate.Slot = observedAtSlot

		dbUpdates = append(dbUpdates, dbUpdate)
	}

	// Update the DB with the delta changes to the memory account state
	var wg sync.WaitGroup
	var cachedMemoryAccountStateUpdateMu sync.Mutex
	for i, dbUpdate := range dbUpdates {
		wg.Add(1)

		go func(dbUpdate *cachedVirtualAccount) {
			defer wg.Done()

			record := &ram.Record{
				Vm: base58VmAddress,

				MemoryAccount: base58MemoryAccountAddress,
				Index:         uint16(dbUpdate.Index),
				IsAllocated:   dbUpdate.IsInitialized,

				Slot: dbUpdate.Slot,

				LastUpdatedAt: time.Now(),
			}

			if dbUpdate.IsInitialized {
				record.Address = &dbUpdate.Address
				record.Type = &dbUpdate.Type
				record.Data = dbUpdate.State
			}

			err := h.ramStore.Save(ctx, record)
			switch err {
			case nil:
				cachedMemoryAccountStateUpdateMu.Lock()
				h.cachedMemoryAccountState[base58MemoryAccountAddress][dbUpdate.Index] = dbUpdate
				cachedMemoryAccountStateUpdateMu.Unlock()
			case ram.ErrStaleState:
				// Should never happen given current locking structure
			default:
				log.Warn("failure updating db record", zap.Error(err))
			}
		}(dbUpdate)

		if i%250 == 0 {
			time.Sleep(time.Second / 4)
		}
	}
	wg.Wait()

	h.lastSuccessfulSlotUpdateMu.Lock()
	lastSuccessfulSlotUpdate = h.lastSuccessfulSlotUpdate[base58MemoryAccountAddress]
	if observedAtSlot > lastSuccessfulSlotUpdate {
		h.lastSuccessfulSlotUpdate[base58MemoryAccountAddress] = observedAtSlot
	}
	h.lastSuccessfulSlotUpdateMu.Unlock()

	return nil
}
