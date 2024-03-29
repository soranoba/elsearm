package elsearm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
)

// CustomIndexNameModel is an interface to implement when customizing IndexName of model.
type CustomIndexNameModel interface {
	// GetIndexName returns an IndexName.
	GetIndexName() string
}

// CustomSearchIndexNameModel is an interface to implement when customizing IndexName of model when searching.
type CustomSearchIndexNameModel interface {
	// GetSearchIndexName returns an IndexName which is search target.
	GetSearchIndexName() []string
}

// CustomDocumentIdModel is an interface to implement when customizing DocumentID of model.
type CustomDocumentIdModel interface {
	// GetDocumentID returns an DocumentID. If the DocumentID is not exist, it returns an error.
	GetDocumentID() (string, error)
}

// CustomDocumentBodyModel is an interface to implement when customizing DocumentBody of model.
type CustomDocumentBodyModel interface {
	GetDocumentBody() (io.Reader, error)
	ParseDocument(io.Reader) error
}

//  AutomaticIDModel is an interface to implement when Elasticsearch automatically creates id of the model.
type AutomaticIDModel interface {
	CustomDocumentIdModel
	// SetDocumentID set the DocumentID. If it failed, it returns an error.
	SetDocumentID(id string) error
}

// DefaultIndexName returns a default IndexName.
func DefaultIndexName(model interface{}) string {
	if model == nil {
		return ""
	}
	return toSnake(reflectValue(model).Type().Name())
}

// DefaultDocumentID returns a default DocumentID.
func DefaultDocumentID(model interface{}) string {
	if model == nil {
		return ""
	}

	value := reflectValue(model)
	field, ok := value.Type().FieldByName("id")
	if !ok {
		field, ok = value.Type().FieldByName("ID")
	}
	if !ok {
		return ""
	}

	id := reflect.Indirect(value.FieldByName(field.Name))
	switch id.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		i = id.Convert(reflect.TypeOf(i)).Interface().(int64)
		return strconv.FormatInt(i, 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var i uint64
		i = id.Convert(reflect.TypeOf(i)).Interface().(uint64)
		return strconv.FormatUint(i, 10)
	case reflect.String:
		return id.Interface().(string)
	default:
		return ""
	}
}

// DefaultDocumentBody returns a default DocumentBody.
func DefaultDocumentBody(model interface{}) (io.Reader, error) {
	if model == nil {
		return nil, errors.New("empty document")
	}

	b, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// DefaultParseDocument parse the content of reader, and update the model.
func DefaultParseDocument(model interface{}, reader io.Reader) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, model); err != nil {
		return err
	}
	return nil
}

func reflectValue(model interface{}) reflect.Value {
	value := reflect.ValueOf(model)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value
}

func toSnake(str string) string {
	runes := []rune(str)
	var p int
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if c >= 'A' && c <= 'Z' {
			runes[i] = c - ('A' - 'a')
			if p+1 < i {
				tmp := append([]rune{'_'}, runes[i:]...)
				runes = append(runes[0:i], tmp...)
				i++
			}
			p = i
		}
	}
	return string(runes)
}
