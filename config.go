package elsearm

type GlobalConfig struct {
	// A prefix of index names. It is included in the return value of elsearm.IndexName.
	IndexNamePrefix string
	// A suffix of index names. It is included in the return value of elsearm.IndexName.
	IndexNameSuffix string
}

var (
	globalConfig GlobalConfig
)

// SetGlobalConfig sets the config that applies globally.
func SetGlobalConfig(cfg GlobalConfig) {
	globalConfig = cfg
}
