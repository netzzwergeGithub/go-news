package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/netzzwergeGithub/go-news/domain"
	"go.etcd.io/bbolt"
)

const articlesBucketName = "articles"

var _ domain.Storage = (*BoltStore)(nil)

type BoltStore struct {
	db *bbolt.DB
}

func NewBoltStore(dbPath string, readOnly bool) (*BoltStore, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout:  1 * time.Second,
		ReadOnly: readOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if !readOnly {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(articlesBucketName))
			return err
		})
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &BoltStore{db: db}, nil
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}
func (s *BoltStore) AddArticles(articles []*domain.Article) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(articlesBucketName))
		if bucket == nil {
			return fmt.Errorf("articles bucket not found")
		}

		for _, article := range articles {
			data, err := json.Marshal(article)
			if err != nil {
				return fmt.Errorf("failed to marshal article %s: %w",
					article.ID, err)
			}

			if err := bucket.Put([]byte(article.ID), data); err != nil {
				return fmt.Errorf("failed to store article %s: %w",
					article.ID, err)
			}
		}

		return nil
	})
}
