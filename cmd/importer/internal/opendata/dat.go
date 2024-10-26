package opendata

import (
	"bufio"
	"context"
	"errors"
	"os"
)

func LoadPlayersFromFile(ctx context.Context, name string, cb func(ctx context.Context, p Player) error) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}

		var p Player
		err2 := p.UnmarshalText(scanner.Bytes())
		if err2 != nil {
			return err2
		}

		err2 = cb(ctx, p)
		if err2 != nil {
			return err2
		}
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
