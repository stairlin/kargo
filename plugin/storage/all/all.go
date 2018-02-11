package all

import (
	// Import all plugins
	_ "github.com/stairlin/kargo/plugin/storage/fs"
	_ "github.com/stairlin/kargo/plugin/storage/s3"
)
