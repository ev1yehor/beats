// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

//go:build !aix

package azureeventhub

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/go-autorest/autorest/azure"

	v2 "github.com/elastic/beats/v7/filebeat/input/v2"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/feature"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/go-concert/unison"
)

const (
	eventHubConnector        = ";EntityPath="
	expandEventListFromField = "records"
	inputName                = "azure-eventhub"
)

var environments = map[string]azure.Environment{
	azure.ChinaCloud.ResourceManagerEndpoint:        azure.ChinaCloud,
	azure.GermanCloud.ResourceManagerEndpoint:       azure.GermanCloud,
	azure.PublicCloud.ResourceManagerEndpoint:       azure.PublicCloud,
	azure.USGovernmentCloud.ResourceManagerEndpoint: azure.USGovernmentCloud,
}

// Plugin returns the Azure Event Hub input plugin.
//
// Required register the plugin loader for the
// input API v2.
func Plugin(log *logp.Logger) v2.Plugin {
	return v2.Plugin{
		Name:       inputName,
		Stability:  feature.Stable,
		Deprecated: false,
		Info:       "Collect logs from Azure Event Hub",
		Manager: &eventHubInputManager{
			log: log,
		},
	}
}

// eventHubInputManager is the manager for the Azure Event Hub input.
//
// It is responsible for creating new instances of the input, according
// to the configuration provided.
type eventHubInputManager struct {
	log *logp.Logger
}

func (m *eventHubInputManager) Init(unison.Group) error {
	return nil
}

func (m *eventHubInputManager) Create(cfg *conf.C) (v2.Input, error) {
	var config azureInputConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("reading %s input config: %w", inputName, err)
	}

	return newEventHubInputV1(config, m.log)
}

func createPipelineClient(pipeline beat.Pipeline) (beat.Client, error) {
	return pipeline.ConnectWith(beat.ClientConfig{
		Processing: beat.ProcessingConfig{
			// This input only produces events with basic types so normalization
			// is not required.
			EventNormalization: to.Ptr(false),
		},
	})
}

// Strip connection string to remove sensitive information
// A connection string should look like this:
// Endpoint=sb://dummynamespace.servicebus.windows.net/;SharedAccessKeyName=DummyAccessKeyName;SharedAccessKey=5dOntTRytoC24opYThisAsit3is2B+OGY1US/fuL3ly=
// This code will remove everything after ';' so key information is stripped
func stripConnectionString(c string) string {
	if parts := strings.SplitN(c, ";", 2); len(parts) == 2 {
		return parts[0]
	}

	// We actually expect the string to have the documented format
	// if we reach here something is wrong, so let's stay on the safe side
	return "(redacted)"
}
