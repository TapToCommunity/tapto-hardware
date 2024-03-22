package state

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

const (
	ReaderTypePN532   = "PN532"
	ReaderTypeACR122U = "ACR122U"
	ReaderTypeUnknown = "Unknown"
)

type Token struct {
	Type     string
	UID      string
	Text     string
	ScanTime time.Time
}

type State struct {
	mu              sync.RWMutex
	updateHook      *func(st *State)
	readerConnected bool
	readerType      string
	activeCard      Token
	lastScanned     Token
	stopService     bool
	disableLauncher bool
	writeRequest    string
	dbLoadTime      time.Time
	uidMap          map[string]string
	textMap         map[string]string
}

func (s *State) SetUpdateHook(hook *func(st *State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateHook = hook
}

func (s *State) SetActiveCard(card Token) {
	s.mu.Lock()

	if s.activeCard == card {
		log.Debug().Msg("ignoring duplicate card")
		s.mu.Unlock()
		return
	}

	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}

	s.mu.Unlock()

	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) GetActiveCard() Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeCard
}

func (s *State) GetLastScanned() Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastScanned
}

func (s *State) StopService() {
	s.mu.Lock()
	s.stopService = true
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) ShouldStopService() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopService
}

func (s *State) DisableLauncher() {
	s.mu.Lock()
	s.disableLauncher = true
	if _, err := os.Create(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot create disable launch file: %s", err)
	}
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) EnableLauncher() {
	s.mu.Lock()
	s.disableLauncher = false
	if err := os.Remove(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot remove disable launch file: %s", err)
	}
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) IsLauncherDisabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.disableLauncher
}

func (s *State) GetDB() (map[string]string, map[string]string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.uidMap, s.textMap
}

func (s *State) GetDBLoadTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbLoadTime
}

func (s *State) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	s.dbLoadTime = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) SetReaderConnected(rt string) {
	s.mu.Lock()
	s.readerConnected = true
	s.readerType = rt
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) SetReaderDisconnected() {
	s.mu.Lock()
	s.readerConnected = false
	s.readerType = ""
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) GetReaderStatus() (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readerConnected, s.readerType
}

func (s *State) SetWriteRequest(req string) {
	s.mu.Lock()
	s.writeRequest = req
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) GetWriteRequest() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writeRequest
}
