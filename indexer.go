package elsearm

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch/esapi"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/pkg/errors"
)

// Indexer provides functions to update/delete document in Elasticsearch.
type Indexer struct {
	client *elasticsearch.Client
}

type errorResponse struct {
	Error struct {
		Reason string `json:"reason"`
	} `json:"error"`
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

	res, err := existsReq.Do(context.Background(), indexer.client)
	if err != nil {
		return err
	}

	if res.IsError() {
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

	res, err := createReq.Do(context.Background(), indexer.client)
	if err != nil {
		return err
	}

	_, err = indexer.handleResponse(res)
	return err
}

// DeleteIndex deletes an index that to save the model.
func (indexer *Indexer) DeleteIndex(model interface{}) error {
	deleteReq := &esapi.IndicesDeleteRequest{
		Index: []string{IndexName(model)},
	}

	res, err := deleteReq.Do(context.Background(), indexer.client)
	if err != nil {
		return err
	}

	if res.IsError() {
		_, err = indexer.handleResponse(res)
		return err
	}
	return nil
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

	res, err := deleteReq.Do(context.Background(), indexer.client)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	return nil
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

	res, err := getReq.Do(context.Background(), indexer.client)
	if err != nil {
		return err
	}

	m, err := indexer.handleResponse(res)
	if err != nil {
		return err
	}

	return ParseDocument(model, bytes.NewReader(m["_source"].data))
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

	res, err := indexReq.Do(context.Background(), indexer.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (indexer *Indexer) handleResponse(res *esapi.Response) (map[string]*source, error) {
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		var errRes errorResponse
		if err := json.Unmarshal(b, &errRes); err != nil {
			return nil, err
		}
		return nil, errors.New(errRes.Error.Reason)
	}

	var result map[string]*source
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}
