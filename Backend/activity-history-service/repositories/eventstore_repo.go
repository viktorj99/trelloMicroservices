package repositories

import (
	"activity-history-service/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
)

type EventStoreRepository struct {
	client *esdb.Client
}

func NewEventStoreRepository(uri string) (*EventStoreRepository, error) {
	if !strings.Contains(uri, "?") {
		uri = fmt.Sprintf("%s?tls=false", uri)
	} else if !strings.Contains(uri, "tls=") {
		uri = fmt.Sprintf("%s&tls=false", uri)
	}

	fmt.Println("Connecting to EventStoreDB with URI:", uri)

	settings, err := esdb.ParseConnectionString(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	client, err := esdb.NewClient(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create EventStoreDB client: %w", err)
	}

	return &EventStoreRepository{client: client}, nil
}

func (r *EventStoreRepository) LogActivity(streamName, eventType string, activity model.Activity) error {
	eventData, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	event := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   eventType,
		Data:        eventData,
	}

	ctx := context.Background()
	_, err = r.client.AppendToStream(ctx, streamName, esdb.AppendToStreamOptions{}, event)
	return err
}

func (r *EventStoreRepository) GetActivitiesByProject(projectID string) ([]map[string]interface{}, error) {
	streamName := fmt.Sprintf("project-%s", projectID)
	fmt.Printf("Attempting to read stream: %s\n", streamName)

	maxCount := uint64(100)

	readResult, err := r.client.ReadStream(
		context.Background(),
		streamName,
		esdb.ReadStreamOptions{Direction: esdb.Forwards},
		maxCount,
	)
	if err != nil {
		fmt.Printf("Error reading stream %s: %v\n", streamName, err)
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
			fmt.Printf("Error processing stream %s: %v\n", streamName, err)
			return nil, err
		}
		fmt.Printf("Event found in stream %s: %v\n", streamName, event)

		var activity map[string]interface{}
		if err := json.Unmarshal(event.Event.Data, &activity); err != nil {
			fmt.Printf("Error unmarshaling event data: %v\n", err)
			return nil, err
		}
		activities = append(activities, activity)
	}
	fmt.Printf("Successfully read %d activities from stream %s\n", len(activities), streamName)

	return activities, nil
}
