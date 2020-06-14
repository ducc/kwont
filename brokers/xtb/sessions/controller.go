package sessions

import (
	"context"
	"github.com/google/uuid"
	"github.com/jpillora/backoff"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type SessionController struct {
	*Session
	sync.Mutex
}

func New(ctx context.Context, amqpChan *amqp.Channel, amqpQueue amqp.Queue, topic, username, password string) (*SessionController, error) {
	s := &SessionController{}

	sessionID := uuid.New().String()
	log := logrus.WithField("session_id", sessionID)

	createSession := func() error {
		session, err := newSession(ctx, amqpChan, amqpQueue, topic, username, password, sessionID)
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
		b := backoff.Backoff{
			Factor: 1.5,
			Jitter: false,
			Min:    0,
			Max:    time.Minute * 5,
		}

		for {
			<-s.finished

			dur := b.Duration()
			log.Debugf("session finished, sleeping for %v", dur)
			time.Sleep(dur)

			s.Lock()
			log.Debug("retrying session")
			if err := createSession(); err != nil {
				log.WithError(err).Error("creating retry session")
				s.Unlock()
				break
			}

			s.Unlock()

			ctx := context.Background()
			for _, symbolName := range s.GetTickSubscription() {
				if err := s.Session.AddTickSubscription(ctx, symbolName); err != nil {
					log.WithError(err).WithField("symbol_name", symbolName).Error("adding tick subscription")
					continue
				}
			}
		}
	}()

	return s, nil
}
