package session_checker

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"time"
)

type checker struct {
	ds         protos.DataServiceClient
	router     protos.BrokerServiceClient
	brokerName protos.Broker_Name
}

func Run(ctx context.Context, ds protos.DataServiceClient, router protos.BrokerServiceClient, brokerName protos.Broker_Name, pollInterval time.Duration) {
	r := &checker{
		ds:         ds,
		router:     router,
		brokerName: brokerName,
	}

	r.pollUsers(ctx, pollInterval)
}

func (r *checker) pollUsers(ctx context.Context, pollInterval time.Duration) {
	for {
		r.findUsersToCheckSessions(ctx)
		time.Sleep(pollInterval)
	}
}

func (r *checker) findUsersToCheckSessions(ctx context.Context) {
	// todo request params
	users, err := r.ds.ListUsers(ctx, &protos.ListUsersRequest{})
	if err != nil {
		logrus.WithError(err).Error("listing users to check sessions")
		return
	}

	sessions, err := r.router.GetCurrentSessions(ctx, &protos.GetCurrentSessionsRequest{})
	if err != nil {
		logrus.WithError(err).Error("getting current sessions from router")
		return
	}

	for _, user := range users.Users {
		r.processUser(ctx, user, sessions.SessionId)
	}
}

func ContainsString(i []string, v string) bool {
	for _, s := range i {
		if s == v {
			return true
		}
	}
	return false
}

func (r *checker) processUser(ctx context.Context, user *protos.User, sessions []string) {
	for _, connection := range user.BrokerConnections {
		if connection == nil {
			continue
		}

		if connection.Broker != r.brokerName {
			continue
		}

		r.checkBrokerConnection(ctx, user, connection, sessions)
	}
}

func (r *checker) checkBrokerConnection(ctx context.Context, user *protos.User, connection *protos.User_BrokerConnection, sessions []string) {
	if ContainsString(sessions, connection.SessionId) {
		// the router already has this session id (infering that it cannot be empty)
		// so it's likely it is still active
		return
	}

	r.openBrokerConnection(ctx, user, connection)
}

func (r *checker) openBrokerConnection(ctx context.Context, user *protos.User, connection *protos.User_BrokerConnection) {
	log := logrus.WithField("user_id", user.Id)

	res, err := r.router.OpenSession(ctx, &protos.OpenSessionRequest{
		Username: connection.Username,
		Password: connection.Password,
	})
	if err != nil {
		log.WithError(err).Error("opening session for user")
		return
	}

	log.WithField("session_id", res.SessionId).Info("opened session")
	connection.SessionId = res.SessionId

	r.updateDataService(ctx, user)
}

func (r *checker) updateDataService(ctx context.Context, user *protos.User) {
	// todo this will probably cause a race condition between different broker checkers
	if _, err := r.ds.UpdateUser(ctx, &protos.UpdateUserRequest{
		User: user,
	}); err != nil {
		logrus.WithError(err).Error("updating user")
	}
}
