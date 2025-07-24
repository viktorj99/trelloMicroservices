package repositories

import (
	"activity-history-service/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
)

type EventStoreRepository struct {
	client *esdb.Client
}

func NewEventStoreRepository(uri string) (*EventStoreRepository, error) {
	settings, err := esdb.ParseConnectionString(uri)
	if err != nil {
		return nil, fmt.Errorf("parse conn string failed: %w", err)
	}
	client, err := esdb.NewClient(settings)
	if err != nil {
		return nil, fmt.Errorf("create client failed: %w", err)
	}
	return &EventStoreRepository{client: client}, nil
}

func (r *EventStoreRepository) AppendToStream(streamName, eventType string, activity model.Activity) error {
	log.Printf("Appending activity to stream: %s\n", streamName)
	data, err := json.Marshal(activity)
	if err != nil {
		return err
	}
	event := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   eventType,
		Data:        data,
	}

	// Append to project stream
	_, err = r.client.AppendToStream(context.Background(), streamName, esdb.AppendToStreamOptions{}, event)
	if err != nil {
		return err
	}

	// Also append to user stream
	userStream := fmt.Sprintf("user-%s", activity.UserID)
	_, err = r.client.AppendToStream(context.Background(), userStream, esdb.AppendToStreamOptions{}, event)
	return err
}

func (r *EventStoreRepository) ReadStream(projectID string) ([]map[string]interface{}, error) {
	stream := fmt.Sprintf("project-%s", projectID)
	readResult, err := r.client.ReadStream(context.Background(), stream, esdb.ReadStreamOptions{Direction: esdb.Forwards}, 100)
	if err != nil {
		return nil, err
	}
	defer readResult.Close()

	var activities []map[string]interface{}
	for {
		event, err := readResult.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		var a map[string]interface{}
		if err := json.Unmarshal(event.Event.Data, &a); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *EventStoreRepository) ReadStreamByUser(userID string) ([]map[string]interface{}, error) {
	stream := fmt.Sprintf("user-%s", userID)
	readResult, err := r.client.ReadStream(context.Background(), stream, esdb.ReadStreamOptions{Direction: esdb.Forwards}, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to read stream %s: %w", stream, err)
	}
	defer readResult.Close()

	var activities []map[string]interface{}
	for {
		event, err := readResult.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading event from stream %s: %w", stream, err)
		}
		var a map[string]interface{}
		if err := json.Unmarshal(event.Event.Data, &a); err != nil {
			return nil, fmt.Errorf("error unmarshalling event data from stream %s: %w", stream, err)
		}
		activities = append(activities, a)
	}
	return activities, nil
}
