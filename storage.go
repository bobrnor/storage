package storage

import "github.com/mailhog/data"

// Storage represents a storage backend
type Storage interface {
	Store(m *data.Message) (string, error)
	List(start, limit int) (*data.Messages, error)
	Search(kind, query string, start, limit int) (*data.Messages, int, error)
	Count() int
	DeleteOne(id string) error
	DeleteAll() error
	Load(id string) (*data.Message, error)
}

type StorageWithNamespace interface {
	StoreWithNamespace(namespace string, m *data.Message) (string, error)
	ListWithNamespace(namespace string, start, limit int) (*data.Messages, error)
	SearchWithNamespace(namespace string, kind, query string, start, limit int) (*data.Messages, int, error)
	CountWithNamespace(namespace string) int
	DeleteOneWithNamespace(namespace string, id string) error
	DeleteAllWithNamespace(namespace string) error
	LoadWithNamespace(namespace string, id string) (*data.Message, error)
	ListNamespaces() ([]string, error)
}
