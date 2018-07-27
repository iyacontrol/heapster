// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clickhouse

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// ClickhouseConfig Clickhouse Config
type ClickhouseConfig struct {
	DSN string

	// select the current default database
	Database string
	// select the current default table
	Table string
	// auth credentials
	UserName string
	//  auth credentials
	Password string
	Host     string
	// comma separated list of single address host for load-balancing
	AltHosts string
	// timeout in second
	WriteTimeout string
	ReadTimeout  string
	//  random/in_order (default random)
	ConnectionOpenStrategy string
	// establish secure connection (default is false)
	Secure bool
	// skip certificate verification (default is true)
	SkipVerify bool

	BatchSize   int
	ClusterName string
	Concurrency int
}

func BuildConfig(uri *url.URL) (*ClickhouseConfig, error) {
	config := ClickhouseConfig{
		UserName:     "default",
		Password:     "",
		Host:         "localhost:9000",
		Database:     "k8s",
		Secure:       false,
		SkipVerify:   false,
		WriteTimeout: "20",
		ReadTimeout:  "10",
		ClusterName:  "default",
		BatchSize:    32768,
		Concurrency:  1,
	}

	if len(uri.Host) > 0 {
		config.Host = uri.Host
	}
	opts := uri.Query()
	if len(opts["user"]) >= 1 {
		config.UserName = opts["user"][0]
	}
	// TODO: use more secure way to pass the password.
	if len(opts["pw"]) >= 1 {
		config.Password = opts["pw"][0]
	}
	if len(opts["db"]) >= 1 {
		config.Database = opts["db"][0]
	}
	if len(opts["table"]) >= 1 {
		config.Table = opts["table"][0]
	}
	if len(opts["alt_hosts"]) >= 1 {
		config.AltHosts = opts["alt_hosts"][0]
	}
	if len(opts["connection_open_strategy"]) >= 1 {
		connectionOpenStrategy := opts["connection_open_strategy"][0]

		if connectionOpenStrategy == "random" || connectionOpenStrategy == "in_order" {
			config.ConnectionOpenStrategy = connectionOpenStrategy
		} else {
			return nil, errors.New("`connection_open_strategy` flag can only be random or in_order")
		}
	}

	if len(opts["secure"]) >= 1 {
		val, err := strconv.ParseBool(opts["secure"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `secure` flag - %v", err)
		}
		config.Secure = val
	}

	if len(opts["insecuressl"]) >= 1 {
		val, err := strconv.ParseBool(opts["insecuressl"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `insecuressl` flag - %v", err)
		}
		config.SkipVerify = val
	}

	if len(opts["batchsize"]) >= 1 {
		batchsize, err := strconv.Atoi(opts["batchsize"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `batchsize` flag - %v", err)
		}

		if batchsize <= 0 {
			return nil, errors.New("`batchsize` flag can only be positive")
		}

		config.BatchSize = batchsize
	}
	if len(opts["write_timeout"]) >= 1 {
		writeTimeout, err := strconv.Atoi(opts["write_timeout"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `write_timeout` flag - %v", err)
		}

		if writeTimeout <= 0 {
			return nil, errors.New("`write_timeout` flag can only be positive")
		}

		config.WriteTimeout = opts["write_timeout"][0]
	}
	if len(opts["read_timeout"]) >= 1 {
		readTimeout, err := strconv.Atoi(opts["read_timeout"][0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse `read_timeout` flag - %v", err)
		}

		if readTimeout <= 0 {
			return nil, errors.New("`read_timeout` flag can only be positive")
		}

		config.ReadTimeout = opts["read_timeout"][0]
	}

	if len(opts["cluster_name"]) >= 1 {
		config.ClusterName = opts["cluster_name"][0]
	}

	config.DSN = "tcp://" + config.Host + "?username=" + config.UserName + "&password=" + config.Password + "&database=" + config.Database + "&write_timeout=" + config.WriteTimeout + "&read_timeout=" + config.ReadTimeout

	if config.AltHosts != "" {
		config.DSN = config.DSN + "&alt_hosts=" + config.AltHosts
	}
	if config.ConnectionOpenStrategy != "" {
		config.DSN = config.DSN + "&connection_open_strategy=" + config.ConnectionOpenStrategy
	}

	return &config, nil
}
