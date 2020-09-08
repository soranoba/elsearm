package elsearm

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch/esapi"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

// Indexer provides functions to update/delete document in Elasticsearch.
type Indexer struct {
	client *elasticsearch.Client
	ctx    context.Context
}

type source struct {
	data []byte
}

func (s *source) UnmarshalJSON(data []byte) error {
	s.data = data
	return nil
}

// NewIndexer creates an Indexer.
func NewIndexer(client *elasticsearch.Client) *Indexer {
	return &Indexer{
		client: client,
		ctx:    context.Background(),
	}
}

// WithContext specifies a context to use and returns a new Indexer.
func (indexer *Indexer) WithContext(ctx context.Context) *Indexer {
	return &Indexer{
		client: indexer.client,
		ctx:    ctx,
	}
}

// CreateIndexIfNotExist creates an index, if it to save the model does not exist.
func (indexer *Indexer) CreateIndexIfNotExist(model interface{}, baseReqs ...*esapi.IndicesCreateRequest) error {
	if len(baseReqs) > 1 {
		panic("CreateIndexIfNotExist only accept one or two arguments")
	}

	createReq := &esapi.IndicesCreateRequest{}
	if len(baseReqs) == 1 {
		createReq = baseReqs[0]
	}

	if createReq.Index == "" {
		createReq.Index = IndexName(model)
	}

	existsReq := &esapi.IndicesExistsRequest{
		Index: []string{createReq.Index},
	}

	if err := indexer.Do(existsReq); err != nil {
		return indexer.CreateIndex(model, baseReqs...)
	}
	return nil
}

// CreateIndex creates an index that to save the model.
// If it already exists, it returns an error.
func (indexer *Indexer) CreateIndex(model interface{}, baseReqs ...*esapi.IndicesCreateRequest) error {
	if len(baseReqs) > 1 {
		panic("CreateIndex only accept one or two arguments")
	}

	createReq := &esapi.IndicesCreateRequest{}
	if len(baseReqs) == 1 {
		createReq = baseReqs[0]
	}

	if createReq.Index == "" {
		createReq.Index = IndexName(model)
	}

	return indexer.Do(createReq)
}

// DeleteIndex deletes an index that to save the model.
func (indexer *Indexer) DeleteIndex(model interface{}) error {
	deleteReq := &esapi.IndicesDeleteRequest{
		Index: []string{IndexName(model)},
	}
	return indexer.Do(deleteReq)
}

// Delete a document from Index.
func (indexer *Indexer) Delete(model interface{}, baseReqs ...*esapi.DeleteRequest) error {
	if model == nil {
		return nil
	}

	if len(baseReqs) > 1 {
		panic("Delete only accept one or two arguments")
	}

	deleteReq := &esapi.DeleteRequest{}
	if len(baseReqs) == 1 {
		deleteReq = baseReqs[0]
	}

	if deleteReq.Index == "" {
		deleteReq.Index = IndexName(model)
	}
	if deleteReq.DocumentID == "" {
		deleteReq.DocumentID = DocumentID(model)
	}

	return indexer.Do(deleteReq)
}

// Get a document from Index.
func (indexer *Indexer) Get(model interface{}, baseReqs ...*esapi.GetRequest) error {
	if model == nil {
		return nil
	}

	if len(baseReqs) > 1 {
		panic("Get only accept one or two arguments")
	}

	getReq := &esapi.GetRequest{}
	if len(baseReqs) == 1 {
		getReq = baseReqs[0]
	}

	if getReq.Index == "" {
		getReq.Index = IndexName(model)
	}
	if getReq.DocumentID == "" {
		getReq.DocumentID = DocumentID(model)
	}

	var result map[string]*source
	if err := indexer.Do(getReq, &result); err != nil {
		return err
	}

	s := result["_source"]
	if s == nil {
		s = &source{}
	}
	return ParseDocument(model, bytes.NewReader(s.data))
}

// CreateWithoutID create a document in index without DocumentID.
func (indexer *Indexer) CreateWithoutID(model interface{}, baseReqs ...*esapi.IndexRequest) error {
	if model == nil {
		return nil
	}

	if len(baseReqs) > 1 {
		panic("CreateWithID only accept one or two arguments")
	}

	indexReq := &esapi.IndexRequest{}
	if len(baseReqs) == 1 {
		indexReq = baseReqs[0]
	}

	if indexReq.Index == "" {
		indexReq.Index = IndexName(model)
	}
	if indexReq.Body == nil {
		reader, err := DocumentBody(model)
		if err != nil {
			return err
		}
		indexReq.Body = reader
	}

	return indexer.Do(indexReq)
}

// Update (or create) the document in index.
func (indexer *Indexer) Update(model interface{}, baseReqs ...*esapi.IndexRequest) error {
	if model == nil {
		return nil
	}

	if len(baseReqs) > 1 {
		panic("Update only accept one or two arguments")
	}

	indexReq := &esapi.IndexRequest{}
	if len(baseReqs) == 1 {
		indexReq = baseReqs[0]
	}

	if indexReq.Index == "" {
		indexReq.Index = IndexName(model)
	}
	if indexReq.DocumentID == "" {
		indexReq.DocumentID = DocumentID(model)
	}
	if indexReq.Body == nil {
		reader, err := DocumentBody(model)
		if err != nil {
			return err
		}
		indexReq.Body = reader
	}

	return indexer.Do(indexReq)
}

func (indexer *Indexer) Count(model interface{}, baseReqs ...*esapi.CountRequest) (int, error) {
	if model == nil {
		return 0, nil
	}

	if len(baseReqs) > 1 {
		panic("Count only accept one or two arguments")
	}

	countReq := &esapi.CountRequest{}
	if len(baseReqs) == 1 {
		countReq = baseReqs[0]
	}

	if countReq.Index == nil {
		countReq.Index = []string{IndexName(model)}
	}

	type CountResponse struct {
		Count int `json:"count"`
	}
	var res CountResponse
	if err := indexer.Do(countReq, &res); err != nil {
		return 0, err
	}
	return res.Count, nil
}

// Do execute the request.
// When models specified, it parses and set a model if succeeded.
func (indexer *Indexer) Do(req Request, models ...interface{}) error {
	if len(models) > 1 {
		panic("Do only accept one or two arguments")
	}

	res, err := req.Do(indexer.ctx, indexer.client)
	if err != nil {
		return err
	}

	if !res.IsError() && len(models) == 0 {
		return nil
	}

	var model interface{}
	if len(models) == 0 {
		model = &map[string]interface{}{}
	} else {
		model = models[0]
	}

	if err := indexer.handleResponse(model, res); err != nil {
		return err
	}
	return nil
}

func (indexer *Indexer) handleResponse(model interface{}, res *esapi.Response) error {
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.IsError() {
		var errRes ErrorResponse
		if err := json.Unmarshal(b, &errRes); err != nil {
			return err
		}
		return &errRes
	}

	if err := json.Unmarshal(b, model); err != nil {
		return err
	}
	return nil
}
