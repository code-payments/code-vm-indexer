package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/code-payments/code-server/pkg/solana/cvm"
	"github.com/code-payments/code-vm-indexer/data/ram"
)

func RunTests(t *testing.T, s ram.Store, teardown func()) {
	for _, tf := range []func(t *testing.T, s ram.Store){
		testRoundTrip,
		testGetAllByMemoryAccount,
		testGetAllVirtualAccountsByAddressAndType,
	} {
		tf(t, s)
		teardown()
	}
}

func testRoundTrip(t *testing.T, s ram.Store) {
	t.Run("testRoundTrip", func(t *testing.T) {
		ctx := context.Background()

		memoryAccount := "memory_account"
		address := "address"
		accountType := cvm.VirtualAccountTypeTimelock

		_, err := s.GetAllVirtualAccountsByAddressAndType(ctx, address, accountType)
		assert.Equal(t, ram.ErrNotFound, err)

		_, err = s.GetAllByMemoryAccount(ctx, memoryAccount)
		assert.Equal(t, ram.ErrNotFound, err)

		start := time.Now()

		expected := &ram.Record{
			Vm: "vm",

			MemoryAccount: memoryAccount,
			Index:         12345,
			IsAllocated:   true,

			Address: &address,
			Type:    &accountType,
			Data:    []byte("data"),

			Slot: 67890,
		}
		cloned := expected.Clone()

		require.NoError(t, s.Save(ctx, expected))

		assert.EqualValues(t, 1, expected.Id)
		assert.True(t, expected.LastUpdatedAt.After(start))

		actual, err := s.GetAllVirtualAccountsByAddressAndType(ctx, address, accountType)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		assertEquivalentRecords(t, &cloned, actual[0])

		actual, err = s.GetAllByMemoryAccount(ctx, memoryAccount)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		assertEquivalentRecords(t, &cloned, actual[0])

		expected.IsAllocated = false
		expected.Address = nil
		expected.Type = nil
		expected.Data = nil
		assert.Equal(t, ram.ErrStaleState, s.Save(ctx, expected))

		actual, err = s.GetAllVirtualAccountsByAddressAndType(ctx, address, accountType)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		assertEquivalentRecords(t, &cloned, actual[0])

		actual, err = s.GetAllByMemoryAccount(ctx, memoryAccount)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		assertEquivalentRecords(t, &cloned, actual[0])

		expected.Slot += 1
		cloned = expected.Clone()
		require.NoError(t, s.Save(ctx, expected))

		_, err = s.GetAllVirtualAccountsByAddressAndType(ctx, address, accountType)
		assert.Equal(t, ram.ErrNotFound, err)

		actual, err = s.GetAllByMemoryAccount(ctx, memoryAccount)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		assertEquivalentRecords(t, &cloned, actual[0])
	})
}

func testGetAllByMemoryAccount(t *testing.T, s ram.Store) {
	t.Run("testGetAllByMemoryAccount", func(t *testing.T) {
		ctx := context.Background()

		memoryAccount := "memory_account"

		var expected []*ram.Record
		for i := 0; i < 10; i++ {
			record := &ram.Record{
				Vm: "vm",

				MemoryAccount: memoryAccount,
				Index:         uint16(i),
				IsAllocated:   false,

				Slot: uint64(i + 1),
			}
			if i%2 == 0 {
				address := fmt.Sprintf("address%d", i)
				accountType := cvm.VirtualAccountTypeTimelock
				record.IsAllocated = true
				record.Address = &address
				record.Type = &accountType
				record.Data = []byte(fmt.Sprintf("data%d", i))
			}

			cloned := record.Clone()
			expected = append(expected, &cloned)

			require.NoError(t, s.Save(ctx, record))
		}

		actual, err := s.GetAllByMemoryAccount(ctx, memoryAccount)
		require.NoError(t, err)
		require.Len(t, actual, len(expected))
		for i, record := range actual {
			assertEquivalentRecords(t, record, expected[i])
		}
	})
}

func testGetAllVirtualAccountsByAddressAndType(t *testing.T, s ram.Store) {
	t.Run("testGetAllVirtualAccountsByAddressAndType", func(t *testing.T) {
		ctx := context.Background()

		addressToQuery := "address0"

		var expected []*ram.Record
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				address := fmt.Sprintf("address%d", i)
				accountType := cvm.VirtualAccountType(i)
				record := &ram.Record{
					Vm: "vm",

					MemoryAccount: "memory_account",
					Index:         uint16(3*i + j),
					IsAllocated:   true,

					Address: &address,
					Type:    &accountType,
					Data:    []byte(fmt.Sprintf("data%d", j)),

					Slot: 1,
				}

				if address == addressToQuery {
					cloned := record.Clone()
					expected = append(expected, &cloned)
				}

				require.NoError(t, s.Save(ctx, record))
			}
		}

		actual, err := s.GetAllVirtualAccountsByAddressAndType(ctx, addressToQuery, cvm.VirtualAccountTypeDurableNonce)
		require.NoError(t, err)
		require.Len(t, actual, len(expected))
		for i, record := range actual {
			assertEquivalentRecords(t, record, expected[i])
		}
	})
}

func assertEquivalentRecords(t *testing.T, obj1, obj2 *ram.Record) {
	assert.Equal(t, obj1.Vm, obj2.Vm)

	assert.Equal(t, obj1.MemoryAccount, obj2.MemoryAccount)
	assert.Equal(t, obj1.Index, obj2.Index)
	assert.Equal(t, obj1.IsAllocated, obj2.IsAllocated)

	assert.Equal(t, obj1.Address, obj2.Address)
	assert.Equal(t, obj1.Type, obj2.Type)
	assert.Equal(t, obj1.Data, obj2.Data)

	assert.Equal(t, obj1.Slot, obj2.Slot)
}
