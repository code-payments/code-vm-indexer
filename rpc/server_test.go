package rpc

import (
	"context"
	"crypto/ed25519"
	"math/rand"
	"testing"

	"github.com/code-payments/code-server/pkg/code/common"
	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/code-payments/code-server/pkg/testutil"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	indexerpb "github.com/code-payments/code-vm-indexer/generated/indexer/v1"

	"github.com/code-payments/code-vm-indexer/data/ram"
	memory_ram_store "github.com/code-payments/code-vm-indexer/data/ram/memory"
)

func TestGetVirtualDurableNonce_HappyPath_Memory(t *testing.T) {
	env, cleanup := setup(t)
	defer cleanup()

	vmAccount := testutil.NewRandomAccount(t)

	vdn := generateVirtualDurableNonce()

	resp, err := env.client.GetVirtualDurableNonce(env.ctx, &indexerpb.GetVirtualDurableNonceRequest{
		VmAccount: &indexerpb.Address{Value: vmAccount.PublicKey().ToBytes()},
		Address:   &indexerpb.Address{Value: vdn.Address},
	})
	require.NoError(t, err)
	assert.Equal(t, indexerpb.GetVirtualDurableNonceResponse_NOT_FOUND, resp.Result)

	ramRecord := env.saveVirtualAccountToRamDb(t, vmAccount, cvm.VirtualAccountTypeDurableNonce, base58.Encode(vdn.Address), vdn.Marshal())

	resp, err = env.client.GetVirtualDurableNonce(env.ctx, &indexerpb.GetVirtualDurableNonceRequest{
		VmAccount: &indexerpb.Address{Value: vmAccount.PublicKey().ToBytes()},
		Address:   &indexerpb.Address{Value: vdn.Address},
	})
	require.NoError(t, err)

	assert.Equal(t, indexerpb.GetVirtualDurableNonceResponse_OK, resp.Result)

	assert.EqualValues(t, vdn.Address, resp.Item.Account.Address.Value)
	assert.EqualValues(t, vdn.Nonce[:], resp.Item.Account.Nonce.Value)

	memoryStorage := resp.Item.Storage.GetMemory()
	require.NotNil(t, memoryStorage)
	assert.Equal(t, ramRecord.MemoryAccount, base58.Encode(memoryStorage.Account.Value))
	assert.EqualValues(t, ramRecord.Index, memoryStorage.Index)

	assert.Equal(t, ramRecord.Slot, resp.Item.Slot)
}

func TestGetVirtualTimelockAccounts_HappyPath_Memory(t *testing.T) {
	env, cleanup := setup(t)
	defer cleanup()

	vmAccount := testutil.NewRandomAccount(t)
	ownerAccount := testutil.NewRandomAccount(t)

	resp, err := env.client.GetVirtualTimelockAccounts(env.ctx, &indexerpb.GetVirtualTimelockAccountsRequest{
		VmAccount: &indexerpb.Address{Value: vmAccount.PublicKey().ToBytes()},
		Owner:     &indexerpb.Address{Value: ownerAccount.PublicKey().ToBytes()},
	})
	require.NoError(t, err)
	assert.Equal(t, indexerpb.GetVirtualTimelockAccountsResponse_NOT_FOUND, resp.Result)

	var vtas []*cvm.VirtualTimelockAccount
	var ramRecords []*ram.Record
	for i := 0; i < 3; i++ {
		vta := generateVirtualTimelockAccount(ownerAccount.PublicKey().ToBytes())
		vtas = append(vtas, vta)

		ramRecord := env.saveVirtualAccountToRamDb(t, vmAccount, cvm.VirtualAccountTypeTimelock, base58.Encode(vta.Owner), vta.Marshal())
		ramRecords = append(ramRecords, ramRecord)
	}

	resp, err = env.client.GetVirtualTimelockAccounts(env.ctx, &indexerpb.GetVirtualTimelockAccountsRequest{
		VmAccount: &indexerpb.Address{Value: vmAccount.PublicKey().ToBytes()},
		Owner:     &indexerpb.Address{Value: ownerAccount.PublicKey().ToBytes()},
	})
	require.NoError(t, err)
	assert.Equal(t, indexerpb.GetVirtualTimelockAccountsResponse_OK, resp.Result)
	require.Len(t, resp.Items, len(vtas))

	for i := 0; i < len(vtas); i++ {
		vta := vtas[i]
		protoItem := resp.Items[i]
		protoVta := resp.Items[i].Account
		ramRecord := ramRecords[i]

		assert.EqualValues(t, vta.Owner, protoVta.Owner.Value)
		assert.EqualValues(t, vta.Nonce[:], protoVta.Nonce.Value)
		assert.EqualValues(t, vta.TokenBump, protoVta.TokenBump)
		assert.EqualValues(t, vta.UnlockBump, protoVta.UnlockBump)
		assert.EqualValues(t, vta.WithdrawBump, protoVta.WithdrawBump)
		assert.Equal(t, vta.Balance, protoVta.Balance)
		assert.EqualValues(t, vta.Bump, protoVta.Bump)

		memoryStorage := resp.Items[i].Storage.GetMemory()
		require.NotNil(t, memoryStorage)
		assert.Equal(t, ramRecord.MemoryAccount, base58.Encode(memoryStorage.Account.Value))
		assert.EqualValues(t, ramRecord.Index, memoryStorage.Index)

		assert.Equal(t, ramRecord.Slot, protoItem.Slot)
	}
}

func TestGetVirtualRelayAccount_HappyPath_Memory(t *testing.T) {
	env, cleanup := setup(t)
	defer cleanup()

	vmAccount := testutil.NewRandomAccount(t)

	vra := generateVirtualRelayAccount()

	resp, err := env.client.GetVirtualRelayAccount(env.ctx, &indexerpb.GetVirtualRelayAccountRequest{
		VmAccount: &indexerpb.Address{Value: vmAccount.PublicKey().ToBytes()},
		Address:   &indexerpb.Address{Value: vra.Target},
	})
	require.NoError(t, err)
	assert.Equal(t, indexerpb.GetVirtualRelayAccountResponse_NOT_FOUND, resp.Result)

	ramRecord := env.saveVirtualAccountToRamDb(t, vmAccount, cvm.VirtualAccountTypeRelay, base58.Encode(vra.Target), vra.Marshal())

	resp, err = env.client.GetVirtualRelayAccount(env.ctx, &indexerpb.GetVirtualRelayAccountRequest{
		VmAccount: &indexerpb.Address{Value: vmAccount.PublicKey().ToBytes()},
		Address:   &indexerpb.Address{Value: vra.Target},
	})
	require.NoError(t, err)

	assert.Equal(t, indexerpb.GetVirtualRelayAccountResponse_OK, resp.Result)

	assert.EqualValues(t, vra.Target, resp.Item.Account.Target.Value)
	assert.EqualValues(t, vra.Destination, resp.Item.Account.Destination.Value)

	memoryStorage := resp.Item.Storage.GetMemory()
	require.NotNil(t, memoryStorage)
	assert.Equal(t, ramRecord.MemoryAccount, base58.Encode(memoryStorage.Account.Value))
	assert.EqualValues(t, ramRecord.Index, memoryStorage.Index)

	assert.Equal(t, ramRecord.Slot, resp.Item.Slot)
}

type testEnv struct {
	ctx      context.Context
	client   indexerpb.IndexerClient
	ramStore ram.Store
}

func setup(t *testing.T) (env *testEnv, cleanup func()) {
	ctx := context.Background()

	conn, serv, err := testutil.NewServer()
	require.NoError(t, err)

	env = &testEnv{
		ctx:      ctx,
		client:   indexerpb.NewIndexerClient(conn),
		ramStore: memory_ram_store.New(),
	}

	serv.RegisterService(func(server *grpc.Server) {
		indexerpb.RegisterIndexerServer(server, NewServer(env.ramStore))
	})

	cleanup, err = serv.Serve()
	require.NoError(t, err)
	return env, cleanup
}

func (e *testEnv) saveVirtualAccountToRamDb(
	t *testing.T,
	vm *common.Account,
	accountType cvm.VirtualAccountType,
	address string,
	data []byte,
) *ram.Record {
	record := &ram.Record{
		Vm: vm.PublicKey().ToBase58(),

		MemoryAccount: testutil.NewRandomAccount(t).PublicKey().ToBase58(),
		Index:         uint16(rand.Uint32()),
		IsAllocated:   true,

		Address: &address,
		Type:    &accountType,
		Data:    data,

		Slot: 42,
	}

	require.NoError(t, e.ramStore.Save(e.ctx, record))

	return record
}

func generateVirtualDurableNonce() *cvm.VirtualDurableNonce {
	var address [32]byte
	var nonce [32]byte

	rand.Read(address[:])
	rand.Read(nonce[:])

	return &cvm.VirtualDurableNonce{
		Address: address[:],
		Nonce:   nonce,
	}
}

func generateVirtualTimelockAccount(owner ed25519.PublicKey) *cvm.VirtualTimelockAccount {
	var nonce [32]byte

	rand.Read(nonce[:])

	return &cvm.VirtualTimelockAccount{
		Owner: owner[:],
		Nonce: nonce,

		TokenBump:    255,
		UnlockBump:   254,
		WithdrawBump: 253,

		Balance: rand.Uint64(),
		Bump:    252,
	}
}

func generateVirtualRelayAccount() *cvm.VirtualRelayAccount {
	var address [32]byte
	var destination [32]byte

	rand.Read(address[:])
	rand.Read(destination[:])

	return &cvm.VirtualRelayAccount{
		Target:      address[:],
		Destination: destination[:],
	}
}
