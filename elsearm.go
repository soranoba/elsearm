package elsearm

import "io"

// IndexName returns an index name of the model.
// By default, it returns converted to snake case the struct name of model.
func IndexName(model interface{}) string {
	indexName := (func() string {
		searchable, ok := model.(CustomIndexNameModel)
		if ok {
			return searchable.GetIndexName()
		}
		return DefaultIndexName(model)
	})()
	return globalConfig.IndexNamePrefix + indexName + globalConfig.IndexNameSuffix
}

// DocumentID returns a document id of the model.
// By default, it returns value of id or ID field in the model. Otherwise, it returns an empty string.
func DocumentID(model interface{}) string {
	searchable, ok := model.(CustomDocumentIdModel)
	if ok {
		return searchable.GetDocumentID()
	}
	return DefaultDocumentID(model)
}

// DocumentBody transforms the model into a data structure that is stored in Elasticsearch.
// By default, it execute json.Marshal.
func DocumentBody(model interface{}) (io.Reader, error) {
	searchable, ok := model.(CustomDocumentBodyModel)
	if ok {
		return searchable.GetDocumentBody()
	}
	return DefaultDocumentBody(model)
}

// MustDocumentBody is similar to DocumentBody.
// It will panic if the DocumentBody returns an error.
func MustDocumentBody(model interface{}) io.Reader {
	reader, err := DocumentBody(model)
	if err != nil {
		panic(err)
	}
	return reader
}

// ParseDocument parses and applies the value to the model.
// By default, it execute json.Unmarshal.
func ParseDocument(model interface{}, reader io.Reader) error {
	searchable, ok := model.(CustomDocumentBodyModel)
	if ok {
		return searchable.ParseDocument(reader)
	}
	return DefaultParseDocument(model, reader)
}
