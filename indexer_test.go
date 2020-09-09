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
	if count != 0 {
		t.Errorf("invalid count: gots %d, wants %d", count, 0)
	}

	if err = indexer.CreateWithoutID(&User{Name: "Carol"}); err != nil {
		t.Error(err)
	}

	// NOTE: default refresh interval.
	time.Sleep(1 * time.Second)
	count, err = indexer.Count(&User{})
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Errorf("invalid count: gots %d, wants %d", count, 1)
	}
}

func TestIndexerSearch(t *testing.T) {
	_ = indexer.DeleteIndex(&User{})
	if err := indexer.CreateIndex(&User{}); err != nil {
		t.Error(err)
	}

	for _, name := range []string{"Alice", "Bob", "Carol"} {
		if err := indexer.CreateWithoutID(&User{Name: name}); err != nil {
			t.Error(err)
		}
	}

	// NOTE: default refresh interval.
	time.Sleep(1 * time.Second)

	var users []User
	if err := indexer.Search(&users); err != nil {
		t.Error(err)
	}
	if len(users) != 3 ||
		users[0].Name != "Alice" ||
		users[1].Name != "Bob" ||
		users[2].Name != "Carol" {
		t.Errorf("invalid result: got %#v", users)
	}

	var arrUsers [1]User
	if err := indexer.Search(&arrUsers); err != nil {
		t.Error(err)
	}
	if len(arrUsers) != 1 || arrUsers[0].Name != "Alice" {
		t.Errorf("invalid result: got %#v", arrUsers)
	}

	var ptrUsers []*User
	if err := indexer.Search(&ptrUsers); err != nil {
		t.Error(err)
	}
	if len(ptrUsers) != 3 ||
		ptrUsers[0].Name != "Alice" ||
		ptrUsers[1].Name != "Bob" ||
		ptrUsers[2].Name != "Carol" {
		t.Errorf("invalid result: got %#v", ptrUsers)
	}

	var ptrArrUsers [1]*User
	if err := indexer.Search(&ptrArrUsers); err != nil {
		t.Error(err)
	}
	if len(ptrArrUsers) != 1 || ptrArrUsers[0].Name != "Alice" {
		t.Errorf("invalid result: got %#v", ptrArrUsers)
	}
}
