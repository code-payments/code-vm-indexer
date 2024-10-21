package geyser

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"sync"
	"time"

	"github.com/code-payments/code-server/pkg/retry"
	"github.com/code-payments/code-server/pkg/solana"
	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"

	geyserpb "github.com/code-payments/code-vm-indexer/generated/geyser/v1"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type cachedVirtualAccount struct {
	IsInitialized bool
	Index         int
	Type          cvm.VirtualAccountType
	Address       string
	State         []byte
	Slot          uint64
}

type MemoryAccountWithDataUpdateHandler struct {
	log *logrus.Entry

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
func NewMemoryAccountWithDataUpdateHandler(solanaClient solana.Client, ramStore ram.Store, backupWorkerInterval time.Duration, vmsToObserve ...string) ProgramAccountUpdateHandler {
	observableVmAccounts := make(map[string]any)
	for _, vm := range vmsToObserve {
		observableVmAccounts[vm] = struct{}{}
	}

	return &MemoryAccountWithDataUpdateHandler{
		log:                      logrus.StandardLogger().WithField("type", "geyser/handler/memory"),
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
func (h *MemoryAccountWithDataUpdateHandler) Handle(ctx context.Context, update *geyserpb.AccountUpdate) error {
	// Simply fetch finalized account state as the safest option
	var finalizedData []byte
	var finalizedSlot uint64
	var err error
	_, err = retry.Retry(
		func() error {
			finalizedData, finalizedSlot, err = h.solanaClient.GetAccountDataAfterBlock(update.Pubkey, update.Slot)
			return err
		},
		waitForFinalizationRetryStrategies...,
	)
	if err != nil {
		return err
	}

	var state cvm.MemoryAccountWithData
	if err := state.Unmarshal(finalizedData); err != nil {
		return nil
	}
	return h.onStateObserved(ctx, update.Pubkey, finalizedSlot, &state)
}

// RunBackupWorker implements geyser.ProgramAccountUpdateHandler.RunBackupWorker
func (h *MemoryAccountWithDataUpdateHandler) RunBackupWorker(ctx context.Context) error {
	return h.backupWorker(ctx)
}

func (h *MemoryAccountWithDataUpdateHandler) backupWorker(ctx context.Context) error {
	log := h.log.WithField("method", "backupWorker")

	for {
		select {
		case <-time.After(h.backupWorkerInterval):
			addresses := make([]string, 0)

			h.cachedMemoryAccountStateMu.RLock()
			for address := range h.cachedMemoryAccountState {
				addresses = append(addresses, address)
			}
			h.cachedMemoryAccountStateMu.RUnlock()

			for _, address := range addresses {
				log := log.WithField("address", address)

				decodedAddress, err := base58.Decode(address)
				if err != nil {
					log.WithError(err).Warn("address is invalid")
					continue
				}

				h.lastSuccessfulSlotUpdateMu.RLock()
				minSlot := h.lastSuccessfulSlotUpdate[address]
				h.lastSuccessfulSlotUpdateMu.RUnlock()

				data, observedAtSlot, err := h.solanaClient.GetAccountDataAfterBlock(decodedAddress, minSlot)
				if err != nil {
					log.WithError(err).Warn("failure getting account data")
					continue
				}

				var state cvm.MemoryAccountWithData
				if err := state.Unmarshal(data); err != nil {
					log.WithError(err).Warn("invalid account data state")
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

func (h *MemoryAccountWithDataUpdateHandler) onStateObserved(ctx context.Context, address ed25519.PublicKey, observedAtSlot uint64, state *cvm.MemoryAccountWithData) error {
	log := h.log.WithField("method", "onStateObserved")

	base58VmAddress := base58.Encode(state.Vm)
	base58MemoryAccountAddress := base58.Encode(address)

	log = log.WithFields(logrus.Fields{
		"vm":      base58VmAddress,
		"address": base58MemoryAccountAddress,
	})

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
		case ram.ErrNotFound:
		default:
			log.WithError(err).Warn("failure loading memory account state from db")
			return err
		}

		h.cachedMemoryAccountState[base58MemoryAccountAddress] = cachedState
	}

	// Track delta changes to the memory account state to be persisted into the DB
	var dbUpdates []*cachedVirtualAccount
	for index, item := range state.Data.Items {
		log := log.WithField("index", index)

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

		isInitialized := item.IsAllocated()

		var base58VirtualAccountAddress string
		var newVirtualAccountState []byte
		var virtualAccountType cvm.VirtualAccountType
		if isInitialized {
			newVirtualAccountState, _ = state.Data.Read(index)
			virtualAccountType = cvm.VirtualAccountType(newVirtualAccountState[0])
			newVirtualAccountState = newVirtualAccountState[1:]

			switch virtualAccountType {
			case cvm.VirtualAccountTypeDurableNonce:
				var virtualAccountState cvm.VirtualDurableNonce
				if err := virtualAccountState.UnmarshalDirectly(newVirtualAccountState); err != nil {
					log.WithError(err).Warn("failure unmarshalling virtual durable nonce")
					return err
				}
				base58VirtualAccountAddress = base58.Encode(virtualAccountState.Address)
			case cvm.VirtualAccountTypeTimelock:
				var virtualAccountState cvm.VirtualTimelockAccount
				if err := virtualAccountState.UnmarshalDirectly(newVirtualAccountState); err != nil {
					log.WithError(err).Warn("failure unmarshalling virtual timelock account")
					return err
				}
				base58VirtualAccountAddress = base58.Encode(virtualAccountState.Owner)
			case cvm.VirtualAccountTypeRelay:
				var virtualAccountState cvm.VirtualRelayAccount
				if err := virtualAccountState.UnmarshalDirectly(newVirtualAccountState); err != nil {
					log.WithError(err).Warn("failure unmarshalling virtual relay account")
					return err
				}
				base58VirtualAccountAddress = base58.Encode(virtualAccountState.Address)
			default:
				// Changelog item, which isn't being tracked in the virtual accounts DB,
				// so treat it as an unititialized memory item
				isInitialized = false
			}
		}

		var dbUpdate *cachedVirtualAccount
		if isInitialized {
			if cachedVirtualAccountState.IsInitialized &&
				cachedVirtualAccountState.Address == base58VirtualAccountAddress &&
				bytes.Equal(cachedVirtualAccountState.State, newVirtualAccountState) &&
				observedAtSlot-cachedVirtualAccountState.Slot < 100 { // todo: configurable?
				continue
			}

			dbUpdate = &cachedVirtualAccount{
				IsInitialized: true,
				Index:         index,
				Type:          virtualAccountType,
				Address:       base58VirtualAccountAddress,
				State:         newVirtualAccountState,
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
	for _, dbUpdate := range dbUpdates {
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
				h.cachedMemoryAccountState[base58MemoryAccountAddress][dbUpdate.Index] = dbUpdate
			case ram.ErrStaleState:
				// Should never happen given current locking structure
			default:
				log.WithError(err).Warn("failure updating db record")
			}
		}(dbUpdate)
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
