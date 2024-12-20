package internal

import (
	"context"
	"fmt"

	"github.com/algorandfoundation/algorun-tui/api"
)

type State string

const (
	FastCatchupState State = "FAST-CATCHUP"
	SyncingState     State = "SYNCING"
	StableState      State = "RUNNING"
)

// StatusModel represents a status response from algod.Status
type StatusModel struct {
	State       State
	Version     string
	Network     string
	Voting      bool
	NeedsUpdate bool
	LastRound   uint64 // Last recorded round
}

// String prints the last round value
func (m *StatusModel) String() string {
	return fmt.Sprintf("\nLastRound: %d\n", m.LastRound)
}
func (m *StatusModel) Update(lastRound int, catchupTime int, aquiredBlocks *int, upgradeNodeVote *bool) {
	m.LastRound = uint64(lastRound)
	if catchupTime > 0 {
		if aquiredBlocks != nil {
			m.State = FastCatchupState
		} else {
			m.State = SyncingState
		}
	} else {
		m.State = StableState
	}
	if upgradeNodeVote != nil {
		m.Voting = *upgradeNodeVote
	}
}

// Fetch handles algod.Status
func (m *StatusModel) Fetch(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg HttpPkgInterface) error {
	if m.Version == "" || m.Version == "N/A" {
		v, err := client.GetVersionWithResponse(ctx)
		if err != nil {
			return err
		}
		if v.StatusCode() != 200 {
			return fmt.Errorf("Status code %d: %s", v.StatusCode(), v.Status())
		}
		m.Network = v.JSON200.GenesisId
		m.Version = fmt.Sprintf("v%d.%d.%d-%s", v.JSON200.Build.Major, v.JSON200.Build.Minor, v.JSON200.Build.BuildNumber, v.JSON200.Build.Channel)
		currentRelease, err := GetGoAlgorandRelease(v.JSON200.Build.Channel, httpPkg)
		if err != nil {
			return err
		}
		if currentRelease != nil && m.Version != *currentRelease {
			m.NeedsUpdate = true
		} else {
			m.NeedsUpdate = false
		}
	}

	s, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return err
	}

	if s.StatusCode() != 200 {
		return fmt.Errorf("Status code %d: %s", s.StatusCode(), s.Status())
	}

	m.Update(s.JSON200.LastRound, s.JSON200.CatchupTime, s.JSON200.CatchpointAcquiredBlocks, s.JSON200.UpgradeNodeVote)
	return nil
}
