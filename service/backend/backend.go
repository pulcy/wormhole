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
	"sort"
	"strings"
)

type Backend interface {
	// Watch for changes in the backend and return where there is a change.
	Watch() error

	// Load all registered instances for the configured service
	Get() (ServiceRegistrations, error)
}

type ServiceRegistration struct {
	ServicePort int
	Instances   ServiceInstances
}

func (sr ServiceRegistration) FullString() string {
	return fmt.Sprintf("%d-%s",
		sr.ServicePort,
		sr.Instances.FullString())
}

type ServiceRegistrations []ServiceRegistration

func (list ServiceRegistrations) Sort() {
	sort.Sort(list)
	for _, sr := range list {
		sr.Instances.Sort()
	}
}

// Len is the number of elements in the collection.
func (list ServiceRegistrations) Len() int {
	return len(list)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (list ServiceRegistrations) Less(i, j int) bool {
	a := list[i].FullString()
	b := list[j].FullString()
	return strings.Compare(a, b) < 0
}

// Swap swaps the elements with indexes i and j.
func (list ServiceRegistrations) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

type ServiceInstance struct {
	IP   string // IP address to connect to to reach the service instance
	Port int    // Port to connect to to reach the service instance
}

func (si ServiceInstance) FullString() string {
	return fmt.Sprintf("%s-%d", si.IP, si.Port)
}

type ServiceInstances []ServiceInstance

func (list ServiceInstances) FullString() string {
	slist := []string{}
	for _, si := range list {
		slist = append(slist, si.FullString())
	}
	sort.Strings(slist)
	return "[" + strings.Join(slist, ",") + "]"
}

func (list ServiceInstances) Sort() {
	sort.Sort(list)
}

// Len is the number of elements in the collection.
func (list ServiceInstances) Len() int {
	return len(list)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (list ServiceInstances) Less(i, j int) bool {
	a := list[i].FullString()
	b := list[j].FullString()
	return strings.Compare(a, b) < 0
}

// Swap swaps the elements with indexes i and j.
func (list ServiceInstances) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
