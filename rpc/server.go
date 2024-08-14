package rpc

import (
	"context"

	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	indexerpb "github.com/code-payments/code-vm-indexer/generated/indexer/v1"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type server struct {
	log *logrus.Entry

	ramStore ram.Store

	indexerpb.UnimplementedIndexerServer
}

// NewServer returns a new indexerpb.IndexerServer implementation
func NewServer(ramStore ram.Store) indexerpb.IndexerServer {
	return &server{
		log: logrus.StandardLogger().WithField("type", "rpc/indexer/server"),

		ramStore: ramStore,
	}
}

// GetVirtualTimelockAccounts implements indexerpb.IndexerServer.GetVirtualTimelockAccounts
func (s *server) GetVirtualTimelockAccounts(ctx context.Context, req *indexerpb.GetVirtualTimelockAccountsRequest) (*indexerpb.GetVirtualTimelockAccountsResponse, error) {
	log := s.log.WithFields(logrus.Fields{
		"method": "GetVirtualTimelockAccounts",
		"vm":     base58.Encode(req.VmAccount.Value),
		"owner":  base58.Encode(req.Owner.Value),
	})

	records, err := s.ramStore.GetAllVirtualAccountsByAddressAndType(ctx, base58.Encode(req.Owner.Value), cvm.VirtualAccountTypeTimelock)
	if err == ram.ErrNotFound {
		return &indexerpb.GetVirtualTimelockAccountsResponse{
			Result: indexerpb.GetVirtualTimelockAccountsResponse_NOT_FOUND,
		}, nil
	} else if err != nil {
		log.WithError(err).Warn("failure querying db")
		return nil, status.Error(codes.Internal, "")
	}

	okResp := &indexerpb.GetVirtualTimelockAccountsResponse{
		Result: indexerpb.GetVirtualTimelockAccountsResponse_OK,
		Items:  make([]*indexerpb.VirtualTimelockAccountWithStorageMetadata, len(records)),
	}

	for i, record := range records {
		decodedMemoryAccountAddress, err := base58.Decode(record.MemoryAccount)
		if err != nil {
			log.WithError(err).Warn("invalid memory account address")
			return nil, status.Error(codes.Internal, "")
		}

		vta, ok := record.ToVirtualTimelockAccount()
		if !ok {
			log.Warn("invalid virtual durable nonce data")
			return nil, status.Error(codes.Internal, "")
		}

		okResp.Items[i] = &indexerpb.VirtualTimelockAccountWithStorageMetadata{
			Account: &indexerpb.VirtualTimelockAccount{
				Owner:        &indexerpb.Address{Value: vta.Owner},
				Nonce:        &indexerpb.Hash{Value: vta.Nonce[:]},
				TokenBump:    uint32(vta.TokenBump),
				UnlockBump:   uint32(vta.UnlockBump),
				WithdrawBump: uint32(vta.WithdrawBump),
				Balance:      vta.Balance,
				Bump:         uint32(vta.Bump),
			},
			Storage: &indexerpb.VirtualAccountStorage{
				Storage: &indexerpb.VirtualAccountStorage_Memory{
					Memory: &indexerpb.MemoryVirtualAccountStorage{
						Account: &indexerpb.Address{Value: decodedMemoryAccountAddress},
						Index:   uint32(record.Index),
					},
				},
			},
			Slot: record.Slot,
		}
	}

	return okResp, nil
}

// GetVirtualDurableNonce implements indexerpb.IndexerServer.GetVirtualDurableNonce
func (s *server) GetVirtualDurableNonce(ctx context.Context, req *indexerpb.GetVirtualDurableNonceRequest) (*indexerpb.GetVirtualDurableNonceResponse, error) {
	log := s.log.WithFields(logrus.Fields{
		"method":  "GetVirtualDurableNonce",
		"vm":      base58.Encode(req.VmAccount.Value),
		"address": base58.Encode(req.Address.Value),
	})

	records, err := s.ramStore.GetAllVirtualAccountsByAddressAndType(ctx, base58.Encode(req.Address.Value), cvm.VirtualAccountTypeDurableNonce)
	if err == ram.ErrNotFound {
		return &indexerpb.GetVirtualDurableNonceResponse{
			Result: indexerpb.GetVirtualDurableNonceResponse_NOT_FOUND,
		}, nil
	} else if err != nil {
		log.WithError(err).Warn("failure querying db")
		return nil, status.Error(codes.Internal, "")
	}

	if len(records) > 1 {
		log.Warn("found multiple database records")
		return nil, status.Error(codes.Internal, "")
	}
	record := records[0]

	decodedMemoryAccountAddress, err := base58.Decode(record.MemoryAccount)
	if err != nil {
		log.WithError(err).Warn("invalid memory account address")
		return nil, status.Error(codes.Internal, "")
	}

	vdn, ok := record.ToVirtualDurableNonce()
	if !ok {
		log.Warn("invalid virtual durable nonce data")
		return nil, status.Error(codes.Internal, "")
	}

	return &indexerpb.GetVirtualDurableNonceResponse{
		Result: indexerpb.GetVirtualDurableNonceResponse_OK,
		Item: &indexerpb.VirtualDurableNonceWithStorageMetadata{
			Account: &indexerpb.VirtualDurableNonce{
				Address: &indexerpb.Address{Value: vdn.Address},
				Nonce:   &indexerpb.Hash{Value: vdn.Nonce[:]},
			},
			Storage: &indexerpb.VirtualAccountStorage{
				Storage: &indexerpb.VirtualAccountStorage_Memory{
					Memory: &indexerpb.MemoryVirtualAccountStorage{
						Account: &indexerpb.Address{Value: decodedMemoryAccountAddress},
						Index:   uint32(record.Index),
					},
				},
			},
			Slot: record.Slot,
		},
	}, nil
}

// GetVirtualDurableNonce implements indexerpb.IndexerServer.GetVirtualDurableNonce
func (s *server) GetVirtualRelayAccount(ctx context.Context, req *indexerpb.GetVirtualRelayAccountRequest) (*indexerpb.GetVirtualRelayAccountResponse, error) {
	log := s.log.WithFields(logrus.Fields{
		"method":  "GetVirtualRelayAccount",
		"vm":      base58.Encode(req.VmAccount.Value),
		"address": base58.Encode(req.Address.Value),
	})

	records, err := s.ramStore.GetAllVirtualAccountsByAddressAndType(ctx, base58.Encode(req.Address.Value), cvm.VirtualAccountTypeRelay)
	if err == ram.ErrNotFound {
		return &indexerpb.GetVirtualRelayAccountResponse{
			Result: indexerpb.GetVirtualRelayAccountResponse_NOT_FOUND,
		}, nil
	} else if err != nil {
		log.WithError(err).Warn("failure querying db")
		return nil, status.Error(codes.Internal, "")
	}

	if len(records) > 1 {
		log.Warn("found multiple database records")
		return nil, status.Error(codes.Internal, "")
	}
	record := records[0]

	decodedMemoryAccountAddress, err := base58.Decode(record.MemoryAccount)
	if err != nil {
		log.WithError(err).Warn("invalid memory account address")
		return nil, status.Error(codes.Internal, "")
	}

	vra, ok := record.ToVirtualRelayAccount()
	if !ok {
		log.Warn("invalid virtual relay account data")
		return nil, status.Error(codes.Internal, "")
	}

	return &indexerpb.GetVirtualRelayAccountResponse{
		Result: indexerpb.GetVirtualRelayAccountResponse_OK,
		Item: &indexerpb.VirtualRelayAccountWithStorageMetadata{
			Account: &indexerpb.VirtualRelayAccount{
				Address:     &indexerpb.Address{Value: vra.Address},
				Commitment:  &indexerpb.Hash{Value: vra.Commitment[:]},
				RecentRoot:  &indexerpb.Hash{Value: vra.RecentRoot[:]},
				Destination: &indexerpb.Address{Value: vra.Destination},
			},
			Storage: &indexerpb.VirtualAccountStorage{
				Storage: &indexerpb.VirtualAccountStorage_Memory{
					Memory: &indexerpb.MemoryVirtualAccountStorage{
						Account: &indexerpb.Address{Value: decodedMemoryAccountAddress},
						Index:   uint32(record.Index),
					},
				},
			},
			Slot: record.Slot,
		},
	}, nil
}
