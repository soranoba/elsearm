package elsearm

import (
	"testing"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

var bulkIndexer *BulkIndexer

func init() {
	esClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic("failed to create elasticsearch client")
	}
	indexer = NewIndexer(esClient)
	bulk, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		NumWorkers:    1,
		Client:        esClient,
		FlushInterval: 50 * time.Millisecond,
	})
	bulkIndexer = NewBulkIndexer(bulk)
}

func TestBulkIndexerCreateWithoutID(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{Name: "Bob"}
	if err := bulkIndexer.CreateWithoutID(user); err != nil {
		t.Error(err)
	}
}

func TestBulkIndexerUpdate(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{ID: 1, Name: "Alice"}
	if err := bulkIndexer.Update(user); err != nil {
		t.Error(err)
	}
}
func TestBulkIndexerDelete(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{ID: 1, Name: "Alice"}
	if err := bulkIndexer.Update(user); err != nil {
		t.Error(err)
	}
	if err := bulkIndexer.Delete(user); err != nil {
		t.Error(err)
	}
	user = &User{ID: 1}
	if err := indexer.Get(user); err == nil {
		t.Errorf("Get should fail but succeeded")
	}
}
