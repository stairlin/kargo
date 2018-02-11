package all

import (
	// Import all plugins
	_ "github.com/stairlin/kargo/plugin/notification/pagerduty"
	_ "github.com/stairlin/kargo/plugin/notification/slack"
)
