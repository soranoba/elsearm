package elsearm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// ErrorResponse is an error response format of Elasticsearch.
type ErrorResponse struct {
	Status uint `json:"status"`
	Err    struct {
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		RootCause []struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"root_cause"`
		CausedBy struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"caused_by"`
	} `json:"error"`
}

func (err *ErrorResponse) Error() string {
	return err.Err.Reason
}

// SearchResponse is an response format of search API.
// ref: https://www.elastic.co/guide/en/elasticsearch/reference/current/search-search.html
type SearchResponse struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

//  SetResult copies the hit result to models.
func (res *SearchResponse) SetResult(models interface{}) error {
	if res == nil {
		return nil
	}

	v := reflect.ValueOf(models)
	if r, ok := models.(reflect.Value); ok {
		v = r
	}

	v = reflect.Indirect(v)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid model: %#v", models))
	}
	for i, hit := range res.Hits.Hits {
		if i >= v.Len() && v.Kind() == reflect.Slice {
			elem := reflect.New(v.Type().Elem()).Elem()
			v.Set(reflect.Append(v, elem))
		}

		if i < v.Len() {
			var aModel interface{}

			vv := v.Index(i)
			if vv.Kind() == reflect.Ptr {
				if vv.Type().Elem().Kind() == reflect.Struct {
					if vv.IsNil() {
						vv.Set(reflect.New(vv.Type().Elem()))
					}
					aModel = vv.Interface()
				}
			} else if vv.Kind() == reflect.Struct {
				aModel = vv.Addr().Interface()
			}

			if aModel == nil {
				panic(fmt.Sprintf("invalid model: %#v", models))
			}

			if err := ParseDocument(aModel, bytes.NewReader(hit.Source)); err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}
