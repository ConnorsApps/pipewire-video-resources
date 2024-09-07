package reconciler

import (
	"github.com/rs/zerolog/log"
)

func (r *Reconciler) ensureSpeakersOn() {
	if r.recordSpeakerState == speakerOn {
		return
	}

	if err := r.hass.TurnSwitchOn(r.Hass.SpeakersDeviceID); err != nil {
		return
	}

	r.recordSpeakerState = speakerOn
}

func (r *Reconciler) ensureSpeakersOff() {
	if r.recordSpeakerState == speakerOff {
		return
	}

	if err := r.hass.TurnSwitchOff(r.Hass.SpeakersDeviceID); err != nil {
		return
	}

	r.recordSpeakerState = speakerOff
}

func (r *Reconciler) Reconcile() {
	if r.voltNode == nil {
		log.Info().Msg("(Volt) - Offline")
		r.ensureSpeakersOff()
	} else {
		log.Info().Msg("(Volt) - Online")
		r.ensureSpeakersOn()
	}
}
