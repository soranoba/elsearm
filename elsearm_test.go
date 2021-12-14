package elsearm

import (
	"strings"
	"testing"
)

type DateMathSupportInIndexNames struct {
}

func (m *DateMathSupportInIndexNames) GetIndexName() string {
	return "<my-index-{now/d}>,<my-index-{now/d-1d}>"
}

func (m *DateMathSupportInIndexNames) GetIndexNames() []string {
	return []string{
		"<my-index-{now/d}>",
		"<my-index-{now/d-1d}>",
	}
}

func TestIndexName(t *testing.T) {
	SetGlobalConfig(GlobalConfig{
		IndexNamePrefix: "prefix_",
		IndexNameSuffix: "_suffix",
	})
	defer SetGlobalConfig(GlobalConfig{})

	wantsName := "<prefix_my-index-{now/d}_suffix>,<prefix_my-index-{now/d-1d}_suffix>"
	name := IndexName(&DateMathSupportInIndexNames{})
	if name != wantsName {
		t.Errorf("invalid index name: gots %s, wants %s", name, wantsName)
	}
}

func TestIndexNamesWithAffix(t *testing.T) {
	SetGlobalConfig(GlobalConfig{
		IndexNamePrefix: "prefix_",
		IndexNameSuffix: "_suffix",
	})
	defer SetGlobalConfig(GlobalConfig{})

	wantsNames := "<prefix_my-index-{now/d}_suffix>,<prefix_my-index-{now/d-1d}_suffix>"
	names := strings.Join(IndexNamesWithAffix((&DateMathSupportInIndexNames{}).GetIndexNames()), ",")
	if names != wantsNames {
		t.Errorf("invalid index name: gots %s, wants %s", names, wantsNames)
	}
}
