package bos_locks

import (
	"context"

	"github.com/baidubce/bce-sdk-go/services/bos"
	bos_storage "github.com/storage-lock/go-bos-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
)

// bosClientConnectionManager 包装已有的 bos.Client
type bosClientConnectionManager struct {
	client *bos.Client
}

func (m *bosClientConnectionManager) Name() string {
	return "bos-client-connection-manager"
}

func (m *bosClientConnectionManager) Take(ctx context.Context) (*bos.Client, error) {
	return m.client, nil
}

func (m *bosClientConnectionManager) Return(ctx context.Context, connection *bos.Client) error {
	return nil
}

func (m *bosClientConnectionManager) Shutdown(ctx context.Context) error {
	return nil
}

var bosClientBeanFactory = storage_lock_factory.NewStorageLockFactoryBeanFactory[*bos.Client, *bos.Client]()

// newBOSStorage 指向 bos_storage.NewBOSStorage，包级变量以便单测注入构造失败，
// 覆盖 GetOrInit initFunc 的错误分支。
var newBOSStorage = bos_storage.NewBOSStorage

// NewBOSLockByClient 从已有 BOS Client 创建分布式锁
func NewBOSLockByClient(ctx context.Context, client *bos.Client, bucketName string, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetBOSLockFactoryByClient(ctx, client, bucketName)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

// GetBOSLockFactoryByClient 获取基于 BOS Client 的锁工厂
func GetBOSLockFactoryByClient(ctx context.Context, client *bos.Client, bucketName string) (*storage_lock_factory.StorageLockFactory[*bos.Client], error) {
	return bosClientBeanFactory.GetOrInit(ctx, client, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*bos.Client], error) {
		cm := bos_storage.NewBOSConnectionManagerFromClient(client)
		options := bos_storage.NewBOSStorageOptions().
			SetConnectionManager(cm).
			SetBucketName(bucketName).
			SetTimeProvider(&bos_storage.LocalTimeProvider{})
		storage, err := newBOSStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		return storage_lock_factory.NewStorageLockFactory[*bos.Client](storage, nil), nil
	})
}
