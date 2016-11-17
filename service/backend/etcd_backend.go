// Copyright (c) 2016 Pulcy.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/coreos/etcd/client"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
)

const (
	recentWatchErrorsMax = 5
)

type etcdBackend struct {
	client            client.Client
	watcher           client.Watcher
	Logger            *logging.Logger
	serviceKey        string
	recentWatchErrors int
}

func NewEtcdBackend(logger *logging.Logger, endpoints []string, path string) (Backend, error) {
	cfg := client.Config{
		Transport: client.DefaultTransport,
		Endpoints: endpoints,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, maskAny(err)
	}
	keysAPI := client.NewKeysAPI(c)
	options := &client.WatcherOptions{
		Recursive: true,
	}
	watcher := keysAPI.Watcher(path, options)
	return &etcdBackend{
		client:     c,
		watcher:    watcher,
		serviceKey: path,
		Logger:     logger,
	}, nil
}

// Watch for changes on a path and return where there is a change.
func (eb *etcdBackend) Watch() error {
	if eb.watcher == nil || eb.recentWatchErrors > recentWatchErrorsMax {
		eb.recentWatchErrors = 0
		keysAPI := client.NewKeysAPI(eb.client)
		options := &client.WatcherOptions{
			Recursive: true,
		}
		eb.watcher = keysAPI.Watcher(eb.serviceKey, options)
	}
	_, err := eb.watcher.Next(context.Background())
	if err != nil {
		eb.recentWatchErrors++
		return maskAny(err)
	}
	eb.recentWatchErrors = 0
	return nil
}

// Load all registered instances for the configured service
func (eb *etcdBackend) Get() (ServiceRegistrations, error) {
	instanceMap, err := eb.readInstancesTree()
	if err != nil {
		return nil, maskAny(err)
	}
	list := ServiceRegistrations{}
	for port, instances := range instanceMap {
		list = append(list, ServiceRegistration{
			ServicePort: port,
			Instances:   instances,
		})
	}
	return list, nil
}

// Load all registered service instances
func (eb *etcdBackend) readInstancesTree() (map[int]ServiceInstances, error) {
	keysAPI := client.NewKeysAPI(eb.client)
	options := &client.GetOptions{
		Recursive: true,
		Sort:      false,
	}
	resp, err := keysAPI.Get(context.Background(), eb.serviceKey, options)
	if err != nil {
		return nil, maskAny(err)
	}
	result := make(map[int]ServiceInstances)
	if resp.Node == nil {
		return result, nil
	}
	for _, instanceNode := range resp.Node.Nodes {
		uniqueID := path.Base(instanceNode.Key)
		parts := strings.Split(uniqueID, ":")
		if len(parts) < 3 {
			eb.Logger.Warning("UniqueID malformed: '%s'", uniqueID)
			continue
		}
		port, err := strconv.Atoi(parts[2])
		if err != nil {
			eb.Logger.Warning("Failed to parse port: '%s'", parts[2])
			continue
		}
		instance, err := eb.parseServiceInstance(instanceNode.Value)
		if err != nil {
			eb.Logger.Warning("Failed to parse instance '%s': %#v", instanceNode.Value, err)
			continue
		}

		list, ok := result[port]
		if !ok {
			list = ServiceInstances{}
		}
		result[port] = append(list, instance)
	}

	return result, nil
}

// parseServiceInstance parses a string in the format of "<ip>':'<port>" into a ServiceInstance.
func (eb *etcdBackend) parseServiceInstance(s string) (ServiceInstance, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return ServiceInstance{}, maskAny(fmt.Errorf("Invalid service instance '%s'", s))
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return ServiceInstance{}, maskAny(fmt.Errorf("Invalid service instance port '%s' in '%s'", parts[1], s))
	}
	return ServiceInstance{
		IP:   parts[0],
		Port: port,
	}, nil
}
