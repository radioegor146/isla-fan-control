package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bougou/go-ipmi"
)

type Gooxi struct {
	client *ipmi.Client
}

const COMMAND_RETRIES = 5

func NewLocalGooxi() (*Gooxi, error) {
	client, err := ipmi.NewOpenClient()
	if err != nil {
		return nil, err
	}
	return &Gooxi{
		client: client,
	}, nil
}

func (g *Gooxi) sendRawCommandWithRetries(ctx context.Context, netFn ipmi.NetFn, command byte, data []byte) ([]byte, error) {
	for i := 0; i < COMMAND_RETRIES; i++ {
		err := g.client.Connect(ctx)
		if err != nil {
			slog.Warn("failed to send raw command, retrying", "try", i+1, "error", err)
			continue
		}
		response, err := g.client.RawCommand(ctx, netFn, command, data, "Raw command")
		if err != nil {
			slog.Warn("failed to send raw command, retrying", "try", i+1, "error", err)
			continue
		}
		err = g.client.Close(ctx)
		if err != nil {
			slog.Warn("failed to send raw command, retrying", "try", i+1, "error", err)
			continue
		}
		return response.Response, nil
	}
	return nil, errors.New("failed to send raw command")
}

func (g *Gooxi) SetFanMode(ctx context.Context, automatic bool) error {
	var commandByte byte
	if automatic {
		commandByte = 1
		slog.Info("changing fan control to automatic")
	} else {
		commandByte = 0
		slog.Info("changing fan control to manual")
	}
	_, err := g.sendRawCommandWithRetries(ctx, 0x2a, 0x62, []byte{commandByte})
	return err
}

func (g *Gooxi) SetFanSpeed(ctx context.Context, fan byte, speed byte) error {
	slog.Info("setting fan speed", "fanId", fan, "speed", speed)
	_, err := g.sendRawCommandWithRetries(ctx, 0x2a, 0x09, []byte{fan, speed})
	return err
}
