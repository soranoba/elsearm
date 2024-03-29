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

func TestIndexerCreateWithoutID_automaticId(t *testing.T) {
	if err := indexer.CreateIndexIfNotExist(&Organization{}); err != nil {
		t.Error(err)
	}
	org := &Organization{Name: "Doodle"}
	if err := indexer.CreateWithoutID(org); err != nil {
		t.Error(err)
	}
	if org.ID == nil {
		t.Errorf("failed to set the document id")
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
	if _, err := indexer.Search(&users); err != nil {
		t.Error(err)
	}
	if len(users) != 3 ||
		users[0].Name != "Alice" ||
		users[1].Name != "Bob" ||
		users[2].Name != "Carol" {
		t.Errorf("invalid result: got %#v", users)
	}

	var arrUsers [1]User
	if _, err := indexer.Search(&arrUsers); err != nil {
		t.Error(err)
	}
	if arrUsers[0].Name != "Alice" {
		t.Errorf("invalid result: got %#v", arrUsers)
	}

	var ptrUsers []*User
	if _, err := indexer.Search(&ptrUsers); err != nil {
		t.Error(err)
	}
	if len(ptrUsers) != 3 ||
		ptrUsers[0].Name != "Alice" ||
		ptrUsers[1].Name != "Bob" ||
		ptrUsers[2].Name != "Carol" {
		t.Errorf("invalid result: got %#v", ptrUsers)
	}

	var ptrArrUsers [1]*User
	if _, err := indexer.Search(&ptrArrUsers); err != nil {
		t.Error(err)
	}
	if ptrArrUsers[0].Name != "Alice" {
		t.Errorf("invalid result: got %#v", ptrArrUsers)
	}
}

func TestIndexerSearch_automaticId(t *testing.T) {
	_ = indexer.DeleteIndex(&Organization{})
	if err := indexer.CreateIndex(&Organization{}); err != nil {
		t.Error(err)
	}

	for _, name := range []string{"Alice", "Bob", "Carol"} {
		if err := indexer.CreateWithoutID(&Organization{Name: name}); err != nil {
			t.Error(err)
		}
	}

	// NOTE: default refresh interval.
	time.Sleep(1 * time.Second)

	var orgs []Organization
	if _, err := indexer.Search(&orgs); err != nil {
		t.Error(err)
	}
	if len(orgs) != 3 ||
		orgs[0].Name != "Alice" ||
		orgs[1].Name != "Bob" ||
		orgs[2].Name != "Carol" ||
		orgs[0].ID == nil ||
		orgs[1].ID == nil ||
		orgs[2].ID == nil {
		t.Errorf("invalid result: got %#v", orgs)
	}
}

func TestIndexerSearchNotFound(t *testing.T) {
	_ = indexer.DeleteIndex(&User{})
	if err := indexer.CreateIndex(&User{}); err != nil {
		t.Error(err)
	}

	var users []User
	if _, err := indexer.Search(&users); err != nil {
		t.Error(err)
	}
	if users == nil || len(users) != 0 {
		t.Errorf("invalid result: got %#v", users)
	}

	var arrUsers [1]User
	if _, err := indexer.Search(&arrUsers); err != nil {
		t.Error(err)
	}
	if arrUsers[0].Name != "" {
		t.Errorf("invalid result: got %#v", arrUsers)
	}

	var ptrUsers []*User
	if _, err := indexer.Search(&ptrUsers); err != nil {
		t.Error(err)
	}
	if ptrUsers == nil || len(ptrUsers) != 0 {
		t.Errorf("invalid result: got %#v", ptrUsers)
	}

	var ptrArrUsers [1]*User
	if _, err := indexer.Search(&ptrArrUsers); err != nil {
		t.Error(err)
	}
	if ptrArrUsers[0] != nil {
		t.Errorf("invalid result: got %#v", ptrArrUsers)
	}
}

func TestIndexerScroll(t *testing.T) {
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
	meta, err := indexer.Search(
		&users,
		indexer.Q.Search.WithSize(1),
		indexer.Q.Search.WithScroll(1*time.Minute),
	)
	if err != nil {
		t.Error(err)
	}
	if meta.ScrollID == "" {
		t.Errorf("scroll id is empty")
	}
	if users == nil || len(users) != 1 || users[0].Name != "Alice" {
		t.Errorf("invalid result: got %#v", users)
	}

	users = nil
	meta, err = indexer.Scroll(
		&users,
		indexer.Q.Scroll.WithScrollID(meta.ScrollID),
		indexer.Q.Scroll.WithScroll(1*time.Minute),
	)
	if err != nil {
		t.Error(err)
	}
	if meta.ScrollID == "" {
		t.Errorf("scroll id is empty")
	}
	if users == nil || len(users) != 1 || users[0].Name != "Bob" {
		t.Errorf("invalid result: got %#v", users)
	}

	users = nil
	meta, err = indexer.Scroll(
		&users,
		indexer.Q.Scroll.WithScrollID(meta.ScrollID),
		indexer.Q.Scroll.WithScroll(1*time.Minute),
	)
	if err != nil {
		t.Error(err)
	}
	if meta.ScrollID == "" {
		t.Errorf("scroll id is empty")
	}
	if users == nil || len(users) != 1 || users[0].Name != "Carol" {
		t.Errorf("invalid result: got %#v", users)
	}

	users = nil
	meta, err = indexer.Scroll(
		&users,
		indexer.Q.Scroll.WithScrollID(meta.ScrollID),
		indexer.Q.Scroll.WithScroll(1*time.Minute),
	)
	if err != nil {
		t.Error(err)
	}
	if meta.ScrollID == "" {
		t.Errorf("scroll id is empty")
	}
	if users == nil || len(users) != 0 {
		t.Errorf("invalid result: got %#v", users)
	}
}
