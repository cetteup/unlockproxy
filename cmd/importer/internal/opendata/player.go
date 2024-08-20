package opendata

import (
	"fmt"
	"strconv"
	"strings"
)

type Player struct {
	PID  int
	Nick string
}

//goland:noinspection GoMixedReceiverTypes
func (p Player) String() string {
	return fmt.Sprintf("%d\t%s", p.PID, p.Nick)
}

//goland:noinspection GoMixedReceiverTypes
func (p *Player) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*p = Player{}
		return nil
	}

	elements := strings.Split(string(text), "\t")
	if len(elements) != 2 {
		return fmt.Errorf("expected 2 elements, got %d", len(elements))
	}

	pid, err2 := strconv.Atoi(elements[0])
	if err2 != nil {
		return fmt.Errorf("failed to parse pid: %w", err2)
	}

	*p = Player{
		PID:  pid,
		Nick: elements[1],
	}

	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (p Player) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}
