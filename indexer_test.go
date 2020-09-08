package elsearm

import (
	"testing"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

var indexer *Indexer

func init() {
	esClient, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic("failed to create elasticsearch client")
	}
	indexer = NewIndexer(esClient)
}

func TestIndexerCreateIndexIfNotExist(t *testing.T) {
	_ = indexer.DeleteIndex(&User{})
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
}

func TestIndexerCreateIndex(t *testing.T) {
	_ = indexer.DeleteIndex(&User{})
	if err := indexer.CreateIndex(&User{}); err != nil {
		t.Error(err)
	}
	if err := indexer.CreateIndex(&User{}); err == nil {
		t.Errorf("CreateIndex should fail but succeeded")
	}
}

func TestIndexerCreateWithoutID(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{Name: "Bob"}
	if err := indexer.CreateWithoutID(user); err != nil {
		t.Error(err)
	}
}

func TestIndexerUpdate(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{ID: 1, Name: "Alice"}
	if err := indexer.Update(user); err != nil {
		t.Error(err)
	}
}

func TestIndexerGet(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{ID: 1, Name: "Alice"}
	if err := indexer.Update(user); err != nil {
		t.Error(err)
	}
	user = &User{ID: 1}
	if err := indexer.Get(user); err != nil {
		t.Error(err)
	}
	if user.ID != 1 || user.Name != "Alice" {
		t.Errorf("invalid response: got = %+v", user)
	}
}

func TestIndexerDelete(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&User{}); err != nil {
		t.Error(err)
	}
	user := &User{ID: 1, Name: "Alice"}
	if err := indexer.Update(user); err != nil {
		t.Error(err)
	}
	if err := indexer.Delete(user); err != nil {
		t.Error(err)
	}
	user = &User{ID: 1}
	if err := indexer.Get(user); err == nil {
		t.Errorf("Get should fail but succeeded")
	}
}

func TestIndexerCount(t *testing.T) {
	_ = indexer.DeleteIndex(&User{})
	if err := indexer.CreateIndex(&User{}); err != nil {
		t.Error(err)
	}

	count, err := indexer.Count(&User{})
	if err != nil {
		t.Error(err)
	}

	if err = indexer.CreateWithoutID(&User{Name: "Carol"}); err != nil {
		t.Error(err)
	}

	var nextCount = 0
	for i := 0; i < 10; i++ {
		nextCount, err := indexer.Count(&User{})
		if err != nil {
			t.Error(err)
		}
		if count+1 == nextCount {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Errorf("invalid count: count = %d, nextCount = %d", count, nextCount)
}
