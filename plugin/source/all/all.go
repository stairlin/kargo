package all

import (
	// Import all plugins
	_ "github.com/stairlin/kargo/plugin/source/consul"
	_ "github.com/stairlin/kargo/plugin/source/couchbase"
	_ "github.com/stairlin/kargo/plugin/source/dir"
	_ "github.com/stairlin/kargo/plugin/source/influxdb"
	_ "github.com/stairlin/kargo/plugin/source/postgresql"
)
