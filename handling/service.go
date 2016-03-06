// Package handling provides the use-case for registering incidents. Used by
// views facing the people handling the cargo along its route.
package handling

import (
	"errors"
	"time"

	"github.com/marcusolsson/goddd/cargo"
	"github.com/marcusolsson/goddd/inspection"
	"github.com/marcusolsson/goddd/location"
	"github.com/marcusolsson/goddd/voyage"
)

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// EventHandler provides a means of subscribing to registered handling events.
type EventHandler interface {
	CargoWasHandled(cargo.HandlingEvent)
}

// Service provides handling operations.
type Service interface {
	// RegisterHandlingEvent registers a handling event in the system, and
	// notifies interested parties that a cargo has been handled.
	RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyageNumber voyage.Number,
		unLocode location.UNLocode, eventType cargo.HandlingEventType) error
}

type service struct {
	handlingEventRepository cargo.HandlingEventRepository
	handlingEventFactory    cargo.HandlingEventFactory
	handlingEventHandler    EventHandler
}

func (s *service) RegisterHandlingEvent(completionTime time.Time, trackingID cargo.TrackingID, voyage voyage.Number,
	loc location.UNLocode, eventType cargo.HandlingEventType) error {
	if completionTime.IsZero() || trackingID == "" || voyage == "" || loc == "" || eventType == cargo.NotHandled {
		return ErrInvalidArgument
	}

	e, err := s.handlingEventFactory.CreateHandlingEvent(time.Now(), completionTime, trackingID, voyage, loc, eventType)
	if err != nil {
		return err
	}

	s.handlingEventRepository.Store(e)
	s.handlingEventHandler.CargoWasHandled(e)

	return nil
}

// NewService creates a handling event service with necessary dependencies.
func NewService(r cargo.HandlingEventRepository, f cargo.HandlingEventFactory, h EventHandler) Service {
	return &service{
		handlingEventRepository: r,
		handlingEventFactory:    f,
		handlingEventHandler:    h,
	}
}

type handlingEventHandler struct {
	InspectionService inspection.Service
}

func (h *handlingEventHandler) CargoWasHandled(event cargo.HandlingEvent) {
	h.InspectionService.InspectCargo(event.TrackingID)
}

// NewEventHandler returns a new instance of a EventHandler.
func NewEventHandler(s inspection.Service) EventHandler {
	return &handlingEventHandler{
		InspectionService: s,
	}
}
