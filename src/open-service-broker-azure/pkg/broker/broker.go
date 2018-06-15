package broker

import (
	"context"
	"errors"
	"fmt"

	"open-service-broker-azure/pkg/api"
	"open-service-broker-azure/pkg/async"
	"open-service-broker-azure/pkg/http/filter"
	"open-service-broker-azure/pkg/service"
	"open-service-broker-azure/pkg/storage"
	log "github.com/Sirupsen/logrus"
)

type errAsyncEngineStopped struct {
	err error
}

func (e *errAsyncEngineStopped) Error() string {
	return fmt.Sprintf("async engine stopped: %s", e.err)
}

type errAPIServerStopped struct {
	err error
}

func (e *errAPIServerStopped) Error() string {
	return fmt.Sprintf("api server stopped: %s", e.err)
}

// Broker is an interface to be implemented by components that implement full
// OSB functionality.
type Broker interface {
	// Start starts all broker components (e.g. API server and async execution
	// engine) and blocks until one of those components returns or fails.
	Start(context.Context) error
}

type broker struct {
	store       storage.Store
	apiServer   api.Server
	asyncEngine async.Engine
	catalog     service.Catalog
}

// NewBroker returns a new Broker
func NewBroker(
	store storage.Store,
	asyncEngine async.Engine,
	filterChain filter.Filter,
	catalog service.Catalog,
) (Broker, error) {
	b := &broker{
		store:       store,
		asyncEngine: asyncEngine,
		catalog:     catalog,
	}

	err := b.asyncEngine.RegisterJob(
		"executeProvisioningStep",
		b.executeProvisioningStep,
	)
	if err != nil {
		return nil, errors.New(
			"error registering async job for executing provisioning steps",
		)
	}
	err = b.asyncEngine.RegisterJob("executeUpdatingStep", b.executeUpdatingStep)
	if err != nil {
		return nil, errors.New(
			"error registering async job for executing updating steps",
		)
	}
	err = b.asyncEngine.RegisterJob(
		"executeDeprovisioningStep",
		b.executeDeprovisioningStep,
	)
	if err != nil {
		return nil, errors.New(
			"error registering async job for executing deprovisioning steps",
		)
	}

	err = b.asyncEngine.RegisterJob("checkParentStatus", b.doCheckParentStatus)
	if err != nil {
		return nil, errors.New(
			"error registering async job for executing check of parent status",
		)
	}

	err = b.asyncEngine.RegisterJob(
		"checkChildrenStatuses",
		b.doCheckChildrenStatuses,
	)
	if err != nil {
		return nil, errors.New(
			"error registering async job for executing check of children " +
				"statuses",
		)
	}

	b.apiServer, err = api.NewServer(
		8080,
		b.store,
		b.asyncEngine,
		filterChain,
		b.catalog,
	)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Start starts all broker components (e.g. API server and async execution
// engine) and blocks until one of those components returns or fails.
func (b *broker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error)
	// Start async engine
	go func() {
		select {
		case errChan <- &errAsyncEngineStopped{err: b.asyncEngine.Run(ctx)}:
		case <-ctx.Done():
		}
	}()
	// Start api server
	go func() {
		select {
		case errChan <- &errAPIServerStopped{err: b.apiServer.Start(ctx)}:
		case <-ctx.Done():
		}
	}()
	select {
	case <-ctx.Done():
		log.Debug("context canceled; broker shutting down")
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
