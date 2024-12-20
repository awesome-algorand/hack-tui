package internal

import (
	"context"
	"errors"
	"time"

	"github.com/algorandfoundation/algorun-tui/api"
)

type StateModel struct {
	// Models
	Status            StatusModel
	Metrics           MetricsModel
	Accounts          map[string]Account
	ParticipationKeys *[]api.ParticipationKey

	// Application State
	Admin bool

	// TODO: handle contexts instead of adding it to state
	Watching bool

	// RPC
	Client  api.ClientWithResponsesInterface
	Context context.Context
}

func (s *StateModel) waitAfterError(err error, cb func(model *StateModel, err error)) {
	if err != nil {
		s.Status.State = "DOWN"
		cb(nil, err)
		time.Sleep(time.Second * 3)
	}
}

// TODO: allow context to handle loop
func (s *StateModel) Watch(cb func(model *StateModel, err error), ctx context.Context, client api.ClientWithResponsesInterface) {
	s.Watching = true
	if s.Metrics.Window == 0 {
		s.Metrics.Window = 100
	}

	err := s.Status.Fetch(ctx, client, new(HttpPkg))
	if err != nil {
		cb(nil, err)
	}

	lastRound := s.Status.LastRound

	for {
		if !s.Watching {
			break
		}

		if s.Status.State == FastCatchupState {
			time.Sleep(time.Second * 10)
			err := s.Status.Fetch(ctx, client, new(HttpPkg))
			if err != nil {
				cb(nil, err)
			}
			continue
		}

		status, err := client.WaitForBlockWithResponse(ctx, int(lastRound))
		s.waitAfterError(err, cb)
		if err != nil {
			continue
		}
		if status.StatusCode() != 200 {
			s.waitAfterError(errors.New(status.Status()), cb)
			continue
		}

		s.Status.State = "Unknown"

		// Update Status
		s.Status.Update(status.JSON200.LastRound, status.JSON200.CatchupTime, status.JSON200.CatchpointAcquiredBlocks, status.JSON200.UpgradeNodeVote)

		// Fetch Keys
		s.UpdateKeys()

		if s.Status.State == SyncingState {
			lastRound = s.Status.LastRound
			cb(s, nil)
			continue
		}
		// Run Round Averages and RX/TX every 5 rounds
		if s.Status.LastRound%5 == 0 || (s.Status.LastRound > 100 && s.Metrics.RoundTime.Seconds() == 0) {
			bm, err := GetBlockMetrics(ctx, client, s.Status.LastRound, s.Metrics.Window)
			s.waitAfterError(err, cb)
			if err != nil {
				continue
			}
			s.Metrics.RoundTime = bm.AvgTime
			s.Metrics.TPS = bm.TPS
			s.UpdateMetricsFromRPC(ctx, client)
		}

		lastRound = s.Status.LastRound
		cb(s, nil)
	}
}

func (s *StateModel) Stop() {
	s.Watching = false
}

func (s *StateModel) UpdateMetricsFromRPC(ctx context.Context, client api.ClientWithResponsesInterface) {
	// Fetch RX/TX
	res, err := GetMetrics(ctx, client)
	if err != nil {
		s.Metrics.Enabled = false
	}
	if err == nil {
		s.Metrics.Enabled = true
		now := time.Now()
		diff := now.Sub(s.Metrics.LastTS)

		s.Metrics.TX = max(0, int(float64(res["algod_network_sent_bytes_total"]-s.Metrics.LastTX)/diff.Seconds()))
		s.Metrics.RX = max(0, int(float64(res["algod_network_received_bytes_total"]-s.Metrics.LastRX)/diff.Seconds()))

		s.Metrics.LastTS = now
		s.Metrics.LastTX = res["algod_network_sent_bytes_total"]
		s.Metrics.LastRX = res["algod_network_received_bytes_total"]
	}
}
func (s *StateModel) UpdateAccounts() error {
	var err error
	s.Accounts, err = AccountsFromState(s, new(Clock), s.Client)
	return err
}

func (s *StateModel) UpdateKeys() {
	var err error
	s.ParticipationKeys, err = GetPartKeys(s.Context, s.Client)
	if err != nil {
		s.Admin = false
	}
	if err == nil {
		s.Admin = true
		err = s.UpdateAccounts()
		if err != nil {
			// TODO: Handle error
		}
	}
}
