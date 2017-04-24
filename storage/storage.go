package distributed_token_bucket

import (
	"github.com/go-redis/redis"
	"errors"
	"fmt"
)

var (
	initializers = map[string]func(interface{}) IStorage {
		"redis": initRedis,
		"memory": initMemory,
	}
)

type (
	IStorage interface {
		Ping() error
		Create(name string, tokens int) error
		Take(bucketName string, tokens int) error
		Put(bucketName string, tokens int) error
		Count(bucketName string) (int, error)
	}
)

func initRedis(options interface{}) IStorage {
	client := redis.NewClient(options.(*redis.Options))
	return &RedisStorage{ client }
}

func initMemory(options interface{}) IStorage {
	return &MemoryStorage{}
}

func NewStorage(name string, options interface{}) (IStorage, error) {
	if initializer, exists := initializers[name]; exists {
		return initializer(options), nil
	}

	return &RedisStorage{}, errors.New(fmt.Sprintf("Initialzer '%v' does not exist.", name))
}