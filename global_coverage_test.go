package bos_locks

import (
	"context"
	"errors"
	"testing"

	"github.com/baidubce/bce-sdk-go/services/bos"
	bos_storage "github.com/storage-lock/go-bos-storage"
	"github.com/stretchr/testify/assert"
)

func testBosClient(t *testing.T) *bos.Client {
	t.Helper()
	client, err := bos.NewClient("test-ak", "test-sk", "http://127.0.0.1:18080")
	assert.Nil(t, err)
	assert.NotNil(t, client)
	return client
}

func TestBosClientConnectionManager(t *testing.T) {
	client := testBosClient(t)
	m := &bosClientConnectionManager{client: client}
	assert.Equal(t, "bos-client-connection-manager", m.Name())

	ctx := context.Background()
	got, err := m.Take(ctx)
	assert.Nil(t, err)
	assert.Equal(t, client, got)
	assert.Nil(t, m.Return(ctx, got))
	assert.Nil(t, m.Shutdown(ctx))
}

func TestGetBOSLockFactoryByClient(t *testing.T) {
	client := testBosClient(t)
	ctx := context.Background()

	// BOS 存储构造本身不探活（Init 只 Take 预设 client），本地即可成功
	factory, err := GetBOSLockFactoryByClient(ctx, client, "test-bucket")
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中（同一 client 返回同一工厂）
	factory2, err := GetBOSLockFactoryByClient(ctx, client, "test-bucket")
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	// BOS 存储未声明 CapabilityCAS，建锁应报能力缺失（诚实声明，不管生产锁）
	_, err = NewBOSLockByClient(ctx, client, "test-bucket", "test-bos-lock")
	assert.NotNil(t, err)
}

func TestGetBOSLockFactoryByClient_StorageError(t *testing.T) {
	old := newBOSStorage
	newBOSStorage = func(ctx context.Context, options *bos_storage.BOSStorageOptions) (*bos_storage.BOSStorage, error) {
		return nil, errors.New("boom")
	}
	defer func() { newBOSStorage = old }()

	client := testBosClient(t)
	ctx := context.Background()

	_, err := GetBOSLockFactoryByClient(ctx, client, "test-bucket")
	assert.NotNil(t, err)
	_, err = NewBOSLockByClient(ctx, client, "test-bucket", "x")
	assert.NotNil(t, err)
}
