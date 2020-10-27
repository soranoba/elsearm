package elsearm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// Indexer provides functions to update/delete document in Elasticsearch.
type Indexer struct {
	Q      *esapi.API
	client *elasticsearch.Client
	ctx    context.Context
}

// SearchResult is the metadata of the search result.
type SearchResult struct {
	// A search context
	ScrollID string
	// Total number of hits
	Total int
	// Accuracy of the total. The value is `eq` or `gte`.
	// ref: https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html
	TotalAccuracy string
}

type source struct {
	data json.RawMessage
}

func (s *source) UnmarshalJSON(data []byte) error {
	s.data = data
	return nil
}

// NewIndexer creates an Indexer.
func NewIndexer(client *elasticsearch.Client) *Indexer {
	return &Indexer{
		Q:      client.API,
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
	assertModel(model)

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
	assertModel(model)

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
	assertModel(model)

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
	assertModel(model)

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
	assertModel(model)

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
	assertModel(model)

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
	assertModel(model)

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
	assertModel(model)

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

// Search documents in the index, and set results to the model.
func (indexer *Indexer) Search(models interface{}, reqFuncs ...func(*esapi.SearchRequest)) (*SearchResult, error) {
	v := reflect.ValueOf(models)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("invalid model: %#v", models))
	}

	v = reflect.Indirect(v)
	var size *int
	if v.Kind() == reflect.Struct {
		s := 1
		size = &s
		v = reflect.ValueOf(&[...]interface{}{models}).Elem()
	} else if v.Kind() == reflect.Array {
		s := v.Len()
		size = &s
	} else if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid model: %#v", models))
	}

	t := v.Type().Elem()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	searchReq := &esapi.SearchRequest{
		Index: SearchIndexName(reflect.New(t).Interface()),
		Size:  size,
	}
	for _, f := range reqFuncs {
		f(searchReq)
	}

	var res SearchResponse
	if err := indexer.Do(searchReq, &res); err != nil {
		return nil, err
	}

	if err := res.SetResult(v); err != nil {
		return nil, err
	}

	return &SearchResult{
		ScrollID:      res.ScrollID,
		Total:         res.Hits.Total.Value,
		TotalAccuracy: res.Hits.Total.Relation,
	}, nil
}

// Scroll the search results, and set results to the model.
func (indexer *Indexer) Scroll(model interface{}, reqFuncs ...func(*esapi.ScrollRequest)) (*SearchResult, error) {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("invalid model: %#v", model))
	}

	v = reflect.Indirect(v)
	if v.Kind() == reflect.Struct {
		v = reflect.ValueOf(&[...]interface{}{model}).Elem()
	} else if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid model: %#v", model))
	}

	scrollReq := &esapi.ScrollRequest{}
	for _, f := range reqFuncs {
		f(scrollReq)
	}

	var res SearchResponse
	if err := indexer.Do(scrollReq, &res); err != nil {
		return nil, err
	}

	if err := res.SetResult(v); err != nil {
		return nil, err
	}

	return &SearchResult{
		ScrollID:      res.ScrollID,
		Total:         res.Hits.Total.Value,
		TotalAccuracy: res.Hits.Total.Relation,
	}, nil
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

func assertModel(model interface{}) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		return
	}
	panic(fmt.Sprintf("invalid model: %#v", model))
}
