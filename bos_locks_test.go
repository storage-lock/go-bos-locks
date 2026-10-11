package bos_locks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalBosLockFactory_InitiallyInitialized(t *testing.T) {
	assert.NotNil(t, bosClientBeanFactory)
}
