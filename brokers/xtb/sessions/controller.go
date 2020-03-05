package sessions

import (
	"context"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type SessionController struct {
	*Session
	sync.Mutex
}

func New(ctx context.Context, natsConn *nats.Conn, topic, username, password string) (*SessionController, error) {
	s := &SessionController{}

	sessionID := uuid.New().String()
	log := logrus.WithField("session_id", sessionID)

	createSession := func() error {
		session, err := newSession(ctx, natsConn, topic, username, password, sessionID)
		if err != nil {
			return err
		}

		s.Session = session
		return nil
	}

	s.Lock()
	if err := createSession(); err != nil {
		s.Unlock()
		return nil, err
	}
	s.Unlock()

	go func() {
		const maxRetries = 3
		retries := 0
		for {
			<-s.finished

			log := log.WithField("retries", retries)
			s.Lock()

			if s.Session.startTime.After(s.Session.startTime.Add(time.Minute * 5)) {
				log.Debug("session kept alive for more than 5 minutes, resetting retries to 1")
				retries = 0
			}

			if retries >= maxRetries {
				log.Error("session has had more than max retries")
				s.Unlock()
				break
			}

			log.Debug("retrying session")
			retries++
			if err := createSession(); err != nil {
				log.WithError(err).Error("creating retry session")
				s.Unlock()
				break
			}

			s.Unlock()
		}
	}()

	return s, nil
}
