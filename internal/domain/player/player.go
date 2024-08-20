package player

import (
	"time"

	"github.com/cetteup/unlockproxy/internal/domain/provider"
)

type Player struct {
	ID                       int
	PID                      int
	Nick                     string
	Country                  *string
	Provider                 provider.Provider
	Added                    time.Time
	Updated                  *time.Time
	LastSeen                 *time.Time
	CountryLastChecked       *time.Time
	OtherProjectsLastChecked *time.Time
	Errors                   int
}
