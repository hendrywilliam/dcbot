package voicemanager

import (
	"sync"

	"github.com/hendrywilliam/siren/internal/voice"
)

type GuildID = string

// VoiceManager tracks active Voice connections keyed by guild ID.
type VoiceManager struct {
	mu           sync.Mutex
	activeVoices map[GuildID]*voice.Voice
}

func NewVoiceManager() VoiceManager {
	return VoiceManager{
		activeVoices: make(map[GuildID]*voice.Voice),
	}
}

// Add inserts or replaces the Voice instance for the given guild.
func (vm *VoiceManager) Add(guildID GuildID, v *voice.Voice) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.activeVoices[guildID] = v
}

// Delete removes the Voice instance for the given guild.
func (vm *VoiceManager) Delete(guildID GuildID) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	delete(vm.activeVoices, guildID)
}

// Get returns the Voice instance for the given guild, or nil if none exists.
func (vm *VoiceManager) Get(guildID GuildID) *voice.Voice {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return vm.activeVoices[guildID]
}
