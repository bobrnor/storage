package storage

import (
	"log"
	"time"

	"strconv"

	"encoding/json"

	"encoding/binary"

	"strings"

	"github.com/boltdb/bolt"
	"github.com/mailhog/data"
)

// BoltDB represents BoltDB backed storage backend
type BoltDB struct {
	db     *bolt.DB
	bucket []byte
}

// CreateBoltDB creates a BoltDB backed storage backend
func CreateBoltDB(path, bucket string) *BoltDB {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("Error opening BoltDB database: %s", err)
		return nil
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("Error creating BoltDB bucket: %s", err)
		return nil
	}

	return &BoltDB{
		db:     db,
		bucket: []byte(bucket),
	}
}

// Store stores a message in BoltDB and returns its storage ID
func (b *BoltDB) Store(m *data.Message) (string, error) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)

		id, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		m.ID = data.MessageID(strconv.FormatUint(id, 10))
		buf, err := json.Marshal(m)
		if err != nil {
			return err
		}

		return bucket.Put(itob(id), buf)
	})

	return string(m.ID), err
}

func itob(v uint64) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// Count returns the number of stored messages
func (b *BoltDB) Count() int {
	var count int
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		stats := bucket.Stats()
		count = stats.KeyN
		return nil
	})

	if err != nil {
		return 0
	}

	return count
}

// Search finds messages matching the query
func (b *BoltDB) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	var matchFunc func(data.Message) bool
	switch kind {
	case "to":
		matchFunc = func(m data.Message) bool {
			for _, to := range m.Raw.To {
				return strings.Contains(to, query)
			}
			return false
		}
	case "from":
		matchFunc = func(m data.Message) bool {
			return strings.Contains(m.Raw.From, query)
		}
	default:
		matchFunc = func(m data.Message) bool {
			return strings.Contains(m.Raw.Data, query)
		}
	}

	count := 0
	messages := data.Messages{}

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)

		c := bucket.Cursor()

		var i int
		var m data.Message

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if err := json.Unmarshal(v, &m); err != nil {
				return err
			}

			if matchFunc(m) {
				count++

				if count >= start && count < start+limit {
					messages = append(messages, m)
				}
			}

			i++
		}

		return nil
	})

	if err != nil {
		log.Printf("Error filtering messages: %s", err)
		return nil, 0, err
	}

	return &messages, count, nil
}

// List returns a list of messages by index
func (b *BoltDB) List(start int, limit int) (*data.Messages, error) {
	messages := data.Messages{}

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)

		c := bucket.Cursor()

		var i int
		var m data.Message

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i < start {
				i++
				continue
			}

			if err := json.Unmarshal(v, &m); err != nil {
				return err
			}

			messages = append(messages, m)

			if len(messages) >= limit {
				return nil
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error listing messages: %s", err)
		return nil, err
	}

	return &messages, nil
}

// DeleteOne deletes an individual message by storage ID
func (b *BoltDB) DeleteOne(id string) error {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		return bucket.Delete(itob(idUint))
	})
}

// DeleteAll deletes all messages stored in BoltDB
func (b *BoltDB) DeleteAll() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(b.bucket); err != nil {
			return err
		}

		_, err := tx.CreateBucket(b.bucket)
		return err
	})
}

// Load loads an individual message by storage ID
func (b *BoltDB) Load(id string) (*data.Message, error) {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	result := data.Message{}

	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucket)
		buf := bucket.Get(itob(idUint))

		return json.Unmarshal(buf, &result)
	})

	return &result, err
}
