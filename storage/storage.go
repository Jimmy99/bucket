package storage

// Interface for storage providers. I know the 'I' prefix isn't Golang convention but I prefer it.
type Storage interface {
	Ping() error
	Create(name string, tokens int) error
	Take(bucketName string, tokens int) error
	TakeAll(bucketName string) (int, error)
	Set(bucketName string, tokens int) error
	Put(bucketName string, tokens int) error
	Count(bucketName string) (int, error)
}