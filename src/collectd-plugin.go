package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/newrelic-experts/go-collectd/api"
	"github.com/newrelic-experts/go-collectd/network"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Port int `default:"25826" help:"Port to listen on"`
}

const (
	integrationName    = "com.newrelic.collectd-plugin"
	integrationVersion = "0.1.0"
)

var args argumentList

// CollectDReceiver implements the Writer interface. It wraps a channel to which metrics received are sent to
type CollectDReceiver struct {
	metricChannel chan []*api.ValueList
}

func populateInventory(inventory sdk.Inventory) error {
	// Insert here the logic of your integration to get the inventory data
	// Ex: inventory.SetItem("softwareVersion", "value", "1.0.1")
	// --
	return nil
}

func populateMetrics(integration *sdk.Integration) error {
	ListenAndWrite(integration)
	return nil
}

// ListenAndWrite starts the collectd receiver
func ListenAndWrite(integration *sdk.Integration) {
	var metricChannel = make(chan []*api.ValueList)
	writer := CollectDReceiver{metricChannel: metricChannel}

	srv := &network.Server{
		PasswordLookup: network.NewAuthFile("/etc/collectd/users"),
		Addr:           net.JoinHostPort("::", strconv.Itoa(args.Port)),
		Writer:         &writer,
	}

	go startInfraMetricProcessor(metricChannel, integration)

	// blocks
	log.Fatal(srv.ListenAndWrite(context.Background()))
}

func (receiver *CollectDReceiver) Write(_ context.Context, valueLists []*api.ValueList) error {
	receiver.metricChannel <- valueLists
	return nil
}

func startInfraMetricProcessor(metricChannel chan []*api.ValueList, integration *sdk.Integration) {
	for {
		valueLists := <-metricChannel
		//Create a new ms (metricset) for each Identifier.Plugin+PluginInstance'
		ms := integration.NewMetricSet("CollectdSample")
		for _, vl := range valueLists {
			//println(vl.Identifier.Plugin + " " + vl.Identifier.PluginInstance + "  " + vl.Identifier.Type + "  " + vl.Identifier.TypeInstance)
			metricName := ""
			if vl.Identifier.PluginInstance != "" {
				metricName = fmt.Sprintf("%s.%s.%s.%s", vl.Identifier.Plugin, vl.Identifier.PluginInstance, vl.Identifier.Type, vl.Identifier.TypeInstance)
			} else {
				metricName = fmt.Sprintf("%s.%s.%s", vl.Identifier.Plugin, vl.Identifier.Type, vl.Identifier.TypeInstance)
			}
			formatValues(vl, ms, metricName)
		}
		fatalIfErr(integration.Publish())
	}
}

func main() {
	integration, err := sdk.NewIntegration(integrationName, integrationVersion, &args)
	fatalIfErr(err)

	if args.All || args.Inventory {
		fatalIfErr(populateInventory(integration.Inventory))
	}

	if args.All || args.Metrics {
		fatalIfErr(populateMetrics(integration))
	}
	fatalIfErr(integration.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func formatValues(vl *api.ValueList, ms *metric.MetricSet, metricName string) {
	//formatTime(vl.Time)

	for _, val := range vl.Values {
		switch v := val.(type) {
		case api.Counter:
			//fields[i+1] = fmt.Sprintf("%d", v)
			ms.SetMetric(metricName, v, metric.GAUGE)
		case api.Gauge:
			//fields[i+1] = fmt.Sprintf("%.15g", v)
			ms.SetMetric(metricName, v, metric.GAUGE)
		case api.Derive:
			//fields[i+1] = fmt.Sprintf("%d", v)
			ms.SetMetric(metricName, v, metric.GAUGE)
		default:
			ms.SetMetric(metricName, v, metric.ATTRIBUTE)
			//return "", fmt.Errorf("unexpected type %T", v)
		}
	}

}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N"
	}

	return fmt.Sprintf("%.3f", float64(t.UnixNano())/1000000000.0)
}
