package storage

import (
	"github.com/go-redis/redis"
	"errors"
	"fmt"
)

// Initializers are a necessary evil in order to match dynamic configuration objects to client objects.
// The api is currently storage.NewStorage("memory", nil) or storage.NewStorage("redis", &redis.Options{})
//
// This can likely be improved upon. I don't want to use reflection.
var initializers = map[string]func(interface{}) IStorage {
	"redis": initRedis,
	"memory": initMemory,
}

// Interface for storage providers. I know the 'I' prefix isn't Golang convention but I prefer it.
type IStorage interface {
	Ping() error
	Create(name string, tokens int) error
	Take(bucketName string, tokens int) error
	Put(bucketName string, tokens int) error
	Count(bucketName string) (int, error)
}

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

	return &MemoryStorage{}, errors.New(fmt.Sprintf("Initialzer '%v' does not exist.", name))
}