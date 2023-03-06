package fcm

import (
	"OpenIM/internal/push/offlinepush"
	"OpenIM/pkg/common/db/cache"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Push(t *testing.T) {
	var redis cache.Model
	offlinePusher := NewClient(redis)
	err := offlinePusher.Push(context.Background(), []string{"userID1"}, "test", "test", &offlinepush.Opts{})
	assert.Nil(t, err)
}