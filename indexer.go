package elsearm

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
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
func (indexer *Indexer) CreateIndexIfNotExist(model interface{}, reqFuncs ...func(*esapi.IndicesCreateRequest)) error {
	createReq := &esapi.IndicesCreateRequest{
		Index: IndexName(model),
	}
	for _, f := range reqFuncs {
		f(createReq)
	}

	existsReq := &esapi.IndicesExistsRequest{
		Index: []string{createReq.Index},
	}
	if err := indexer.Do(existsReq); err != nil {
		return indexer.CreateIndex(model, reqFuncs...)
	}
	return nil
}

// CreateIndex creates an index that to save the model.
// If it already exists, it returns an error.
func (indexer *Indexer) CreateIndex(model interface{}, reqFuncs ...func(*esapi.IndicesCreateRequest)) error {
	createReq := &esapi.IndicesCreateRequest{
		Index: IndexName(model),
	}
	for _, f := range reqFuncs {
		f(createReq)
	}
	return indexer.Do(createReq)
}

// DeleteIndex deletes an index that to save the model.
func (indexer *Indexer) DeleteIndex(model interface{}, reqFuncs ...func(*esapi.IndicesDeleteRequest)) error {
	deleteReq := &esapi.IndicesDeleteRequest{
		Index: []string{IndexName(model)},
	}
	for _, f := range reqFuncs {
		f(deleteReq)
	}
	return indexer.Do(deleteReq)
}

// Delete a document from Index.
func (indexer *Indexer) Delete(model interface{}, reqFuncs ...func(*esapi.DeleteRequest)) error {
	if model == nil {
		return nil
	}

	deleteReq := &esapi.DeleteRequest{
		Index:      IndexName(model),
		DocumentID: DocumentID(model),
	}
	for _, f := range reqFuncs {
		f(deleteReq)
	}

	return indexer.Do(deleteReq)
}

// Get a document from Index.
func (indexer *Indexer) Get(model interface{}, reqFuncs ...func(*esapi.GetRequest)) error {
	if model == nil {
		return nil
	}

	getReq := &esapi.GetRequest{
		Index:      IndexName(model),
		DocumentID: DocumentID(model),
	}
	for _, f := range reqFuncs {
		f(getReq)
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
func (indexer *Indexer) CreateWithoutID(model interface{}, reqFuncs ...func(*esapi.IndexRequest)) error {
	if model == nil {
		return nil
	}

	reader, err := DocumentBody(model)
	if err != nil {
		return err
	}

	indexReq := &esapi.IndexRequest{
		Index: IndexName(model),
		Body:  reader,
	}
	for _, f := range reqFuncs {
		f(indexReq)
	}

	return indexer.Do(indexReq)
}

// Update (or create) the document in index.
func (indexer *Indexer) Update(model interface{}, reqFuncs ...func(*esapi.IndexRequest)) error {
	if model == nil {
		return nil
	}

	reader, err := DocumentBody(model)
	if err != nil {
		return err
	}

	indexReq := &esapi.IndexRequest{
		Index:      IndexName(model),
		DocumentID: DocumentID(model),
		Body:       reader,
	}
	for _, f := range reqFuncs {
		f(indexReq)
	}

	return indexer.Do(indexReq)
}

// Count returns count of documents saved in index.
func (indexer *Indexer) Count(model interface{}, reqFuncs ...func(*esapi.CountRequest)) (int, error) {
	if model == nil {
		return 0, nil
	}

	countReq := &esapi.CountRequest{
		Index: []string{IndexName(model)},
	}
	for _, f := range reqFuncs {
		f(countReq)
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
