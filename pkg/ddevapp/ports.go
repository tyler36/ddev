package ddevapp

import (
	"github.com/ddev/ddev/pkg/nodeps"
	"github.com/ddev/ddev/pkg/util"
)

// GetInternalPort returns the internal exposed port (as a string) for the given service.
// This can be used to find a given port for docker-compose manifests,
// or for automated testing.
func GetInternalPort(app *DdevApp, service string) string {
	switch service {
	case "db":
		if app.Database.Type == nodeps.Postgres {
			return "5432"
		}
		return "3306"
	case "mailpit":
		return "8025"
	case "web":
		return "80"
	case "xhgui":
		return "80"
	}

	util.Failed("Could not find port for service %s", service)
	return ""
}
