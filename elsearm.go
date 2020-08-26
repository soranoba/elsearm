package elsearm

import "io"

type GlobalConfig struct {
	IndexNamePrefix string
	IndexNameSuffix string
}

var (
	globalConfig GlobalConfig
)

func SetGlobalConfig(cfg GlobalConfig) {
	globalConfig = cfg
}

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

func DocumentID(model interface{}) string {
	searchable, ok := model.(CustomDocumentIdModel)
	if ok {
		return searchable.GetDocumentID()
	}
	return DefaultDocumentID(model)
}

func DocumentBody(model interface{}) (io.Reader, error) {
	searchable, ok := model.(CustomDocumentBodyModel)
	if ok {
		return searchable.GetDocumentBody()
	}
	return DefaultDocumentBody(model)
}

func MustDocumentBody(model interface{}) io.Reader {
	reader, err := DocumentBody(model)
	if err != nil {
		panic(err)
	}
	return reader
}

func ParseDocument(model interface{}, reader io.Reader) error {
	searchable, ok := model.(CustomDocumentBodyModel)
	if ok {
		return searchable.ParseDocument(reader)
	}
	return DefaultParseDocument(model, reader)
}
