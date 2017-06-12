package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	etcdRoot = "/sensu.io"
)

func getMutatorsPath(name string) string {
	return fmt.Sprintf("%s/mutators/%s", etcdRoot, name)
}

func getEventsPath(args ...string) string {
	return fmt.Sprintf("%s/events/%s", etcdRoot, strings.Join(args, "/"))
}

// Store is an implementation of the sensu-go/backend/store.Store iface.
type etcdStore struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *Etcd
}

// Mutators
func (s *etcdStore) GetMutators() ([]*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Mutator{}, nil
	}

	mutatorsArray := make([]*types.Mutator, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		mutator := &types.Mutator{}
		err = json.Unmarshal(kv.Value, mutator)
		if err != nil {
			return nil, err
		}
		mutatorsArray[i] = mutator
	}

	return mutatorsArray, nil
}

func (s *etcdStore) GetMutatorByName(name string) (*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	mutatorBytes := resp.Kvs[0].Value
	mutator := &types.Mutator{}
	if err := json.Unmarshal(mutatorBytes, mutator); err != nil {
		return nil, err
	}

	return mutator, nil
}

func (s *etcdStore) DeleteMutatorByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getMutatorsPath(name))
	return err
}

func (s *etcdStore) UpdateMutator(mutator *types.Mutator) error {
	if err := mutator.Validate(); err != nil {
		return err
	}

	mutatorBytes, err := json.Marshal(mutator)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getMutatorsPath(mutator.Name), string(mutatorBytes))
	if err != nil {
		return err
	}

	return nil
}

// Events

func (s *etcdStore) GetEvents() ([]*types.Event, error) {
	resp, err := s.kvc.Get(context.Background(), getEventsPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return []*types.Event{}, nil
	}

	eventsArray := make([]*types.Event, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		event := &types.Event{}
		err = json.Unmarshal(kv.Value, event)
		if err != nil {
			return nil, err
		}
		eventsArray[i] = event
	}

	return eventsArray, nil
}

func (s *etcdStore) GetEventsByEntity(entityID string) ([]*types.Event, error) {
	resp, err := s.kvc.Get(context.Background(), getEventsPath(entityID), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	eventsArray := make([]*types.Event, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		event := &types.Event{}
		err = json.Unmarshal(kv.Value, event)
		if err != nil {
			return nil, err
		}
		eventsArray[i] = event
	}

	return eventsArray, nil
}

func (s *etcdStore) GetEventByEntityCheck(entityID, checkID string) (*types.Event, error) {
	if entityID == "" {
		return nil, errors.New("entity id is required")
	}

	if checkID == "" {
		return nil, errors.New("check id is required")
	}

	resp, err := s.kvc.Get(context.Background(), getEventsPath(entityID, checkID), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	eventBytes := resp.Kvs[0].Value
	event := &types.Event{}
	if err := json.Unmarshal(eventBytes, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *etcdStore) UpdateEvent(event *types.Event) error {
	if event.Check == nil {
		return errors.New("event has no check")
	}

	// TODO(Simon): We should also validate event.Entity since we also use
	// some properties of Entity below, such as ID
	if err := event.Check.Validate(); err != nil {
		return err
	}

	// update the history
	// marshal the new event and store it.
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	entityID := event.Entity.ID
	checkID := event.Check.Config.Name

	_, err = s.kvc.Put(context.TODO(), getEventsPath(entityID, checkID), string(eventBytes))
	if err != nil {
		return err
	}

	return nil
}

func (s *etcdStore) DeleteEventByEntityCheck(entityID, checkID string) error {
	if entityID == "" {
		return errors.New("entity id is required")
	}

	if checkID == "" {
		return errors.New("check id is required")
	}

	_, err := s.kvc.Delete(context.TODO(), getEventsPath(entityID, checkID))
	return err
}

// NewStore ...
func (e *Etcd) NewStore() (store.Store, error) {
	c, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	store := &etcdStore{
		etcd:   e,
		client: c,
		kvc:    clientv3.NewKV(c),
	}
	return store, nil
}
