package reconciler

import (
	pwmonitor "github.com/ConnorsApps/pipewire-monitor-go"
	"github.com/ConnorsApps/pipewire-video-resources/pipewire-listener/config"
	"github.com/ConnorsApps/pipewire-video-resources/pipewire-listener/hass"
	"github.com/rs/zerolog/log"
)

type pipewireNode struct {
	ID    int
	State pwmonitor.State
}

type Reconciler struct {
	*config.Config
	hass               *hass.Client
	voltNode           *pipewireNode
	recordSpeakerState speakerState
}

func New(config *config.Config) *Reconciler {
	return &Reconciler{
		Config: config,
		hass:   hass.New(config.Hass.URL, config.Hass.Token),
	}
}

type speakerState uint8

const (
	speakerStateUnknown speakerState = iota
	speakerOff
	speakerOn
)

func (s speakerState) String() string {
	switch s {
	case speakerStateUnknown:
		return "unknown"
	case speakerOff:
		return "off"
	case speakerOn:
		return "on"
	}
	return "unknown"
}

func (r *Reconciler) GetInitialState() {
	r.getRecordPiState()
	log.Info().Str("state", r.recordSpeakerState.String()).Msg("Speakers")
}

func (r *Reconciler) getRecordPiState() {
	state, err := r.hass.SwitchState(r.Hass.SpeakersEntityID)
	if err != nil {
		r.recordSpeakerState = speakerStateUnknown
		log.Error().Err(err).Msg("Unable to get record pi speaker state")
		return
	}
	switch state.State {
	case hass.SpeakerOn:
		r.recordSpeakerState = speakerOn
	case hass.SpeakerOff:
		r.recordSpeakerState = speakerOff
	}
}

// Returns true if the state was updated
func (r *Reconciler) UpdateState(e *pwmonitor.Event) bool {
	if e.IsRemovalEvent() {
		if r.voltNode != nil && e.ID == r.voltNode.ID {
			r.voltNode = nil
			return true
		}
		return false
	}

	nodeProps, err := e.NodeProps()
	if err != nil {
		log.Error().Err(err).Interface("event", e).Msg("Failed to get node")
		return false
	} else if nodeProps == nil {
		return false
	}

	switch nodeProps.Name {
	case r.Volt_AlsaOutput:
	default:
		return false
	}

	var state pwmonitor.State
	if e.Info != nil && e.Info.State != nil {
		state = *e.Info.State
	}

	r.voltNode = &pipewireNode{
		ID:    e.ID,
		State: state,
	}

	return true
}
