package rpc

import (
	"context"

	"github.com/code-payments/ocp-server/solana/vm"
	"github.com/mr-tron/base58"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	indexerpb "github.com/code-payments/code-vm-indexer/generated/indexer/v1"

	"github.com/code-payments/code-vm-indexer/data/ram"
)

type server struct {
	log *zap.Logger

	ramStore ram.Store

	indexerpb.UnimplementedIndexerServer
}

// NewServer returns a new indexerpb.IndexerServer implementation
func NewServer(log *zap.Logger, ramStore ram.Store) indexerpb.IndexerServer {
	return &server{
		log: log,

		ramStore: ramStore,
	}
}

// GetVirtualTimelockAccounts implements indexerpb.IndexerServer.GetVirtualTimelockAccounts
func (s *server) GetVirtualTimelockAccounts(ctx context.Context, req *indexerpb.GetVirtualTimelockAccountsRequest) (*indexerpb.GetVirtualTimelockAccountsResponse, error) {
	log := s.log.With(
		zap.String("method", "GetVirtualTimelockAccounts"),
		zap.String("vm", base58.Encode(req.VmAccount.Value)),
		zap.String("owner", base58.Encode(req.Owner.Value)),
	)

	records, err := s.ramStore.GetAllVirtualAccountsByVmAndAddressAndType(
		ctx,
		base58.Encode(req.VmAccount.Value),
		base58.Encode(req.Owner.Value),
		vm.VirtualAccountTypeTimelock,
	)
	if err == ram.ErrItemNotFound {
		return &indexerpb.GetVirtualTimelockAccountsResponse{
			Result: indexerpb.GetVirtualTimelockAccountsResponse_NOT_FOUND,
		}, nil
	} else if err != nil {
		log.Warn("failure querying db", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	okResp := &indexerpb.GetVirtualTimelockAccountsResponse{
		Result: indexerpb.GetVirtualTimelockAccountsResponse_OK,
		Items:  make([]*indexerpb.VirtualTimelockAccountWithStorageMetadata, len(records)),
	}

	for i, record := range records {
		decodedVmAddress, err := base58.Decode(record.Vm)
		if err != nil {
			log.Warn("invalid vm address", zap.Error(err))
			return nil, status.Error(codes.Internal, "")
		}

		decodedMemoryAccountAddress, err := base58.Decode(record.MemoryAccount)
		if err != nil {
			log.Warn("invalid memory account address", zap.Error(err))
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
			Slot:      record.Slot,
			VmAccount: &indexerpb.Address{Value: decodedVmAddress},
		}
	}

	return okResp, nil
}

// GetVirtualDurableNonce implements indexerpb.IndexerServer.GetVirtualDurableNonce
func (s *server) GetVirtualDurableNonce(ctx context.Context, req *indexerpb.GetVirtualDurableNonceRequest) (*indexerpb.GetVirtualDurableNonceResponse, error) {
	log := s.log.With(
		zap.String("method", "GetVirtualDurableNonce"),
		zap.String("vm", base58.Encode(req.VmAccount.Value)),
		zap.String("address", base58.Encode(req.Address.Value)),
	)

	records, err := s.ramStore.GetAllVirtualAccountsByVmAndAddressAndType(
		ctx,
		base58.Encode(req.VmAccount.Value),
		base58.Encode(req.Address.Value),
		vm.VirtualAccountTypeDurableNonce,
	)
	if err == ram.ErrItemNotFound {
		return &indexerpb.GetVirtualDurableNonceResponse{
			Result: indexerpb.GetVirtualDurableNonceResponse_NOT_FOUND,
		}, nil
	} else if err != nil {
		log.Warn("failure querying db", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	if len(records) > 1 {
		log.Warn("found multiple database records")
		return nil, status.Error(codes.Internal, "")
	}
	record := records[0]

	decodedVmAddress, err := base58.Decode(record.Vm)
	if err != nil {
		log.Warn("invalid vm address", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	decodedMemoryAccountAddress, err := base58.Decode(record.MemoryAccount)
	if err != nil {
		log.Warn("invalid memory account address", zap.Error(err))
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
				Value:   &indexerpb.Hash{Value: vdn.Value[:]},
			},
			Storage: &indexerpb.VirtualAccountStorage{
				Storage: &indexerpb.VirtualAccountStorage_Memory{
					Memory: &indexerpb.MemoryVirtualAccountStorage{
						Account: &indexerpb.Address{Value: decodedMemoryAccountAddress},
						Index:   uint32(record.Index),
					},
				},
			},
			Slot:      record.Slot,
			VmAccount: &indexerpb.Address{Value: decodedVmAddress},
		},
	}, nil
}

// SearchVirtualTimelockAccounts implements indexerpb.IndexerServer.SearchVirtualTimelockAccounts
func (s *server) SearchVirtualTimelockAccounts(ctx context.Context, req *indexerpb.SearchVirtualTimelockAccountsRequest) (*indexerpb.SearchVirtualTimelockAccountsResponse, error) {
	log := s.log.With(
		zap.String("method", "SearchVirtualTimelockAccounts"),
		zap.String("owner", base58.Encode(req.Owner.Value)),
	)

	records, err := s.ramStore.GetAllVirtualAccountsByAddressAndType(
		ctx,
		base58.Encode(req.Owner.Value),
		vm.VirtualAccountTypeTimelock,
	)
	if err == ram.ErrItemNotFound {
		return &indexerpb.SearchVirtualTimelockAccountsResponse{
			Result: indexerpb.SearchVirtualTimelockAccountsResponse_NOT_FOUND,
		}, nil
	} else if err != nil {
		log.Warn("failure querying db", zap.Error(err))
		return nil, status.Error(codes.Internal, "")
	}

	okResp := &indexerpb.SearchVirtualTimelockAccountsResponse{
		Result: indexerpb.SearchVirtualTimelockAccountsResponse_OK,
		Items:  make([]*indexerpb.VirtualTimelockAccountWithStorageMetadata, len(records)),
	}

	for i, record := range records {
		decodedVmAddress, err := base58.Decode(record.Vm)
		if err != nil {
			log.Warn("invalid vm address", zap.Error(err))
			return nil, status.Error(codes.Internal, "")
		}

		decodedMemoryAccountAddress, err := base58.Decode(record.MemoryAccount)
		if err != nil {
			log.Warn("invalid memory account address", zap.Error(err))
			return nil, status.Error(codes.Internal, "")
		}

		vta, ok := record.ToVirtualTimelockAccount()
		if !ok {
			log.Warn("invalid virtual timelock account data")
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
			Slot:      record.Slot,
			VmAccount: &indexerpb.Address{Value: decodedVmAddress},
		}
	}

	return okResp, nil
}
