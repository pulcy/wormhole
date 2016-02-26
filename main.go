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

package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"

	"github.com/pulcy/wormhole/service"
	"github.com/pulcy/wormhole/service/backend"
)

const (
	projectName     = "wormhole"
	defaultLogLevel = "info"
)

var (
	projectVersion = "dev"
	projectBuild   = "dev"
)

var (
	cmdMain = &cobra.Command{
		Use:   projectName,
		Short: "Forward TCP requests to services",
		Run:   cmdMainRun,
	}
	log   = logging.MustGetLogger(projectName)
	flags struct {
		logLevel        string
		etcdAddr        string
		haproxyConfPath string
	}
)

func init() {
	logging.SetFormatter(logging.MustStringFormatter("[%{level:-5s}] %{message}"))

	cmdMain.Flags().StringVar(&flags.logLevel, "log-level", defaultLogLevel, "Log level (debug|info|warning|error)")
	cmdMain.Flags().StringVar(&flags.etcdAddr, "etcd-addr", "", "Address of etcd backend")
	cmdMain.Flags().StringVar(&flags.haproxyConfPath, "haproxy-conf", "/data/config/haproxy.cfg", "Path of haproxy config file")
}

func main() {
	cmdMain.Execute()
}

func cmdMainRun(cmd *cobra.Command, args []string) {
	// Parse arguments
	if flags.etcdAddr == "" {
		Exitf("Please specify --etcd-addr")
	}
	etcdUrl, err := url.Parse(flags.etcdAddr)
	if err != nil {
		Exitf("--etcd-addr '%s' is not valid: %#v", flags.etcdAddr, err)
	}

	// Set log level
	level, err := logging.LogLevel(flags.logLevel)
	if err != nil {
		Exitf("Invalid log-level '%s': %#v", flags.logLevel, err)
	}
	logging.SetLevel(level, projectName)

	// Prepare backend
	backend, err := backend.NewEtcdBackend(log, etcdUrl)
	if err != nil {
		Exitf("Failed to backend: %#v", err)
	}

	// Prepare service
	if flags.haproxyConfPath == "" {
		Exitf("Please specify --haproxy-conf")
	}
	service := service.NewService(service.ServiceConfig{
		HaproxyConfPath: flags.haproxyConfPath,
	}, service.ServiceDependencies{
		Logger:  log,
		Backend: backend,
	})

	service.Run()
}

func Exitf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	fmt.Printf(format, args...)
	os.Exit(1)
}

func def(envKey, defaultValue string) string {
	s := os.Getenv(envKey)
	if s == "" {
		s = defaultValue
	}
	return s
}
