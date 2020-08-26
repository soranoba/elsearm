package elsearm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

type User struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (u *User) GetDocumentID() string {
	return DefaultDocumentID(u)
}

func (u *User) GetIndexName() string {
	return DefaultIndexName(u)
}

func (u *User) GetDocumentBody() (io.Reader, error) {
	return DefaultDocumentBody(u)
}

func (u *User) ParseDocument(reader io.Reader) error {
	return DefaultParseDocument(u, reader)
}

func TestUserInterface(t *testing.T) {
	var _ CustomDocumentBodyModel = &User{}
	var _ CustomDocumentIdModel = &User{}
	var _ CustomIndexNameModel = &User{}
}

func TestDefaultValues(t *testing.T) {
	user := &User{ID: 1, Name: "Alice"}
	if user.GetDocumentID() != "1" {
		t.Errorf("invalid documentID: got = %s, expect = %s", user.GetDocumentID(), "1")
	}
	if user.GetIndexName() != "user" {
		t.Errorf("invalid indexName: got = %s, expect = %s", user.GetIndexName(), "User")
	}

	reader, err := user.GetDocumentBody()
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "{\"id\":1,\"name\":\"Alice\"}" {
		t.Errorf("invalid body: got = %s", string(b))
	}

	user = &User{}
	if err := user.ParseDocument(bytes.NewReader(b)); err != nil {
		t.Error(err)
	}
	if user.ID != 1 || user.Name != "Alice" {
		t.Errorf("failed to parse: got = %#v", user)
	}
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (t *Team) GetDocumentID() string {
	return t.Name
}

func (t *Team) GetIndexName() string {
	return "team"
}

func (t *Team) GetDocumentBody() (io.Reader, error) {
	return bytes.NewReader([]byte(fmt.Sprintf("%d.%s", t.ID, t.Name))), nil
}

func (t *Team) ParseDocument(reader io.Reader) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	parts := strings.Split(string(b), ".")
	if len(parts) != 2 {
		return errors.New("invalid document")
	}

	i, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	t.ID = int(i)
	t.Name = parts[1]
	return nil
}

func TestTeamInterface(t *testing.T) {
	var _ CustomDocumentBodyModel = &Team{}
	var _ CustomDocumentIdModel = &Team{}
	var _ CustomIndexNameModel = &Team{}
}

func TestCustomValues(t *testing.T) {
	team := &Team{ID: 1, Name: "Booooom"}
	if team.GetDocumentID() != "Booooom" {
		t.Errorf("invalid documentID: got = %s, expect = %s", team.GetDocumentID(), "Booooom")
	}
	if team.GetIndexName() != "team" {
		t.Errorf("invalid indexName: got = %s, expect = %s", team.GetIndexName(), "team")
	}

	reader, err := team.GetDocumentBody()
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Error(err)
	}
	if string(b) != "1.Booooom" {
		t.Errorf("invalid body: got = %s", string(b))
	}

	team = &Team{}
	if err := team.ParseDocument(bytes.NewReader(b)); err != nil {
		t.Error(err)
	}
	if team.ID != 1 || team.Name != "Booooom" {
		t.Errorf("failed to parse: got = %#v", team)
	}
}
