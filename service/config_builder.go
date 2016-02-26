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

package service

import (
	"fmt"

	"github.com/pulcy/wormhole/haproxy"
	"github.com/pulcy/wormhole/service/backend"
)

var (
	globalOptions = []string{
		//"log global",
		"quiet",
	}
	defaultsOptions = []string{
		"mode tcp",
		"timeout connect 5000ms",
		"timeout client 50000ms",
		"timeout server 50000ms",
	}
)

// renderConfig creates a new haproxy configuration content.
func (s *Service) renderConfig(services backend.ServiceRegistrations) (string, error) {
	c := haproxy.NewConfig()
	c.Section("global").Add(globalOptions...)
	c.Section("defaults").Add(defaultsOptions...)

	// Create frontend for each service port
	for _, service := range services {
		frontEndSection := c.Section(fmt.Sprintf("frontend input-%d", service.ServicePort))
		frontEndSection.Add(fmt.Sprintf("bind *:%d", service.ServicePort))
		frontEndSection.Add(
			fmt.Sprintf("default_backend backend-%d", service.ServicePort),
		)
	}

	// Create backends
	for _, sr := range services {
		// Create backend
		backendSection := c.Section(fmt.Sprintf("backend backend-%d", sr.ServicePort))
		backendSection.Add(
			"balance roundrobin",
		)
		backendSection.Add("mode tcp")

		for i, instance := range sr.Instances {
			id := fmt.Sprintf("instance-%d-%d", sr.ServicePort, i)
			backendSection.Add(fmt.Sprintf("server %s %s:%d", id, instance.IP, instance.Port))
		}
	}

	// Render config
	return c.Render(), nil
}
