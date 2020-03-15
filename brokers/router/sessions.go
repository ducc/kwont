package router

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
)

type SessionFinder struct {
	sync.RWMutex
	// session ids to broker service addresses
	sessionAddresses map[string]string
}

func NewSessionFinder(ctx context.Context) *SessionFinder {
	return &SessionFinder{
		sessionAddresses: make(map[string]string),
	}
}

func (s *SessionFinder) SetServiceAddress(sessionID, serviceAddress string) {
	s.RLock()
	if s.GetServiceAddress(sessionID) == serviceAddress {
		s.RUnlock()
		return
	}
	s.RUnlock()

	s.Lock()
	defer s.Unlock()

	logrus.WithFields(logrus.Fields{
		"session_id":      sessionID,
		"service_address": serviceAddress,
	}).Info("adding new session service address")
	s.sessionAddresses[sessionID] = serviceAddress
}

func (s *SessionFinder) GetServiceAddress(sessionID string) string {
	return s.sessionAddresses[sessionID]
}

func (s *SessionFinder) GetSessionsForAddress(serviceAddress string) []string {
	s.RLock()
	defer s.RUnlock()

	sessionIDs := make([]string, 0)

	for sessionID, address := range s.sessionAddresses {
		if address == serviceAddress {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs
}

func (s *SessionFinder) RemoveSession(sessionID string) {
	s.Lock()
	defer s.Unlock()

	delete(s.sessionAddresses, sessionID)
}
