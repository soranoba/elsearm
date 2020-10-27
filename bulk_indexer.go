package elsearm

import (
	"context"

	"github.com/elastic/go-elasticsearch/v7/esutil"
)

// BulkIndexer provides functions to bulk insert/update document in Elasticsearch.
type BulkIndexer struct {
	bulk esutil.BulkIndexer
	ctx  context.Context
}

// NewIndexer creates an Indexer.
func NewBulkIndexer(bulk esutil.BulkIndexer) *BulkIndexer {
	return &BulkIndexer{
		bulk: bulk,
		ctx:  context.Background(),
	}
}

// WithContext specifies a context to use and returns a new BulkIndexer.
func (indexer *BulkIndexer) WithContext(ctx context.Context) *BulkIndexer {
	return &BulkIndexer{
		bulk: indexer.bulk,
		ctx:  ctx,
	}
}

// CreateWithoutID create a document in index without DocumentID.
// Returns an error if the addition to the bulk indexer fails.
func (indexer *BulkIndexer) CreateWithoutID(model interface{}) error {
	assertModel(model)

	reader, err := DocumentBody(model)
	if err != nil {
		return err
	}

	return indexer.bulk.Add(indexer.ctx, esutil.BulkIndexerItem{
		Index:  IndexName(model),
		Action: "index",
		Body:   reader,
	})
}

// Update (or create) the document in index.
// Returns an error if the addition to the bulk indexer fails.
func (indexer *BulkIndexer) Update(model interface{}) error {
	assertModel(model)

	reader, err := DocumentBody(model)
	if err != nil {
		return err
	}

	return indexer.bulk.Add(indexer.ctx, esutil.BulkIndexerItem{
		Index:      IndexName(model),
		DocumentID: DocumentID(model),
		Action:     "index",
		Body:       reader,
	})
}

// Delete a document from Index.
// Returns an error if the addition to the bulk indexer fails.
func (indexer *BulkIndexer) Delete(model interface{}) error {
	assertModel(model)

	return indexer.bulk.Add(indexer.ctx, esutil.BulkIndexerItem{
		Index:      IndexName(model),
		DocumentID: DocumentID(model),
		Action:     "delete",
	})
}
