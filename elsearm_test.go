package elsearm

import (
	"net/url"
	"testing"
)

type DateMathSupportInIndexNames struct {
}

func (m *DateMathSupportInIndexNames) GetIndexName() string {
	return url.QueryEscape("<my-index-{now/d}>,<my-index-{now/d-1d}>")
}

func TestIndexName(t *testing.T) {
	SetGlobalConfig(GlobalConfig{
		IndexNamePrefix: "prefix_",
		IndexNameSuffix: "_suffix",
	})
	defer SetGlobalConfig(GlobalConfig{})

	wantsName := url.QueryEscape("<prefix_my-index-{now/d}_suffix>,<prefix_my-index-{now/d-1d}_suffix>")
	name := IndexName(&DateMathSupportInIndexNames{})
	if name != wantsName {
		t.Errorf("invalid index name: gots %s, wants %s", name, wantsName)
	}
}
