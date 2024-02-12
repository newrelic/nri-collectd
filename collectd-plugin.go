package main

import (
   "context"
   "fmt"
   "net"
   "os"
   "strconv"
   "time"

   "collectd.org/api"
   "collectd.org/network"
   sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
   "github.com/newrelic/infra-integrations-sdk/data/metric"
   "github.com/newrelic/infra-integrations-sdk/integration"
   "github.com/newrelic/infra-integrations-sdk/log"
   "github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
)

type argumentList struct {
   sdkArgs.DefaultArgumentList
   Port         int    `default:"25826" help:"Port to listen on"`
   Key          string `default:"" help:"Insights Insert Key required for sending dimensional metrics"`
   Interval     string `default:"15s" help:"Interval to harvest dimensional metrics"`
   Dim          bool   `default:"false" help:"Report output as dimensional metrics (true|false)"`
   MetricApiUrl string `default:"https://metric-api.newrelic.com/metric/v1" help:"Metric API endpoint to use"`
}

const (
   integrationName    = "com.newrelic.collectd-plugin"
   integrationVersion = "0.2.0"
)

var args argumentList

// CollectDReceiver implements the Writer interface. It wraps a channel to which metrics received are sent to
type CollectDReceiver struct {
   metricChannel chan *api.ValueList
}

// Write captures the valueLists from collectd and places them on the channel
func (receiver *CollectDReceiver) Write(_ context.Context, valueLists *api.ValueList) error {
   receiver.metricChannel <- valueLists
   return nil
}

// Stateful nri- this nri never exits
func main() {
   collectdIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
   fatalIfErr(err)

   if args.All() || args.Metrics {
      fatalIfErr(populateMetrics(collectdIntegration))
   }
   fatalIfErr(collectdIntegration.Publish())
}

func populateMetrics(integration *integration.Integration) error {
   ListenAndWrite(integration)
   return nil
}

// ListenAndWrite starts the collectd receiver, it never exits but listens constantly
func ListenAndWrite(integration *integration.Integration) {
   var metricChannel = make(chan *api.ValueList)
   writer := CollectDReceiver{metricChannel: metricChannel}

   srv := &network.Server{
      PasswordLookup: network.NewAuthFile("/etc/collectd/users"),
      Addr:           net.JoinHostPort("::", strconv.Itoa(args.Port)),
      Writer:         &writer,
   }

   if args.Dim {
      go startDimMetricsProcessor(metricChannel)
   } else {
      go startInfraMetricProcessor(metricChannel, integration)
   }

   // blocks
   log.Fatal(srv.ListenAndWrite(context.Background()))
}

func startDimMetricsProcessor(metricChannel chan *api.ValueList) {
   t, _ := time.ParseDuration(args.Interval)

   h, err := telemetry.NewHarvester(
      telemetry.ConfigAPIKey(args.Key),
      telemetry.ConfigHarvestPeriod(t),
      telemetry.ConfigBasicErrorLogger(os.Stdout),
      telemetry.ConfigMetricsURLOverride(args.MetricApiUrl),
   )
   if err != nil {
      log.Fatal(err)
   }

   for {
      valueLists := <-metricChannel
      vl := valueLists
      // for _, vl := range valueLists {
      // println(vl.Identifier.Plugin + " " + vl.Identifier.PluginInstance + "  " + vl.Identifier.Type + "  " + vl.Identifier.TypeInstance)
      metricName := fmt.Sprintf("%s.%s", vl.Identifier.Type, vl.Identifier.TypeInstance)
      recordDimensionalMetric(vl, h, metricName)
      // }
   }
}

func recordDimensionalMetric(vl *api.ValueList, h *telemetry.Harvester, metricName string) {
   for _, val := range vl.Values {
      switch val.(type) {
      case api.Counter:
         newVal := val.(api.Counter)
         h.RecordMetric(telemetry.Count{
            Timestamp: time.Now(),
            Value:     float64(newVal),
            Name:      metricName,
            Attributes: map[string]interface{}{
               "Plugin":   vl.Plugin,
               "Entity":   vl.PluginInstance,
               "HostName": vl.Host,
            },
         })
      case api.Gauge:
         newVal := val.(api.Gauge)
         h.RecordMetric(telemetry.Gauge{
            Timestamp: time.Now(),
            Value:     float64(newVal),
            Name:      metricName,
            Attributes: map[string]interface{}{
               "Plugin":   vl.Plugin,
               "Entity":   vl.PluginInstance,
               "HostName": vl.Host,
            },
         })
      case api.Derive:
         newVal := val.(api.Derive)
         h.RecordMetric(telemetry.Count{
            Timestamp: time.Now(),
            Value:     float64(newVal),
            Name:      metricName,
            Attributes: map[string]interface{}{
               "Plugin":   vl.Plugin,
               "Entity":   vl.PluginInstance,
               "HostName": vl.Host,
            },
         })
      default:
         newVal := val.(api.Gauge)
         h.RecordMetric(telemetry.Gauge{
            Timestamp: time.Now(),
            Value:     float64(newVal),
            Name:      metricName,
            Attributes: map[string]interface{}{
               "Plugin":   vl.Plugin,
               "Entity":   vl.PluginInstance,
               "HostName": vl.Host,
            },
         })
      }
   }
}

func startInfraMetricProcessor(metricChannel chan *api.ValueList, integration *integration.Integration) {
   // entity := integration.LocalEntity()

   for {
      valueLists := <-metricChannel
      // Create a new ms (metricset) for each Identifier.Plugin+PluginInstance'
      entity := integration.LocalEntity()
      ms := entity.NewMetricSet("CollectdSample")
      vl := valueLists
      // for _, vl := range valueLists {
      // println(vl.Identifier.Plugin + " " + vl.Identifier.PluginInstance + "  " + vl.Identifier.Type + "  " + vl.Identifier.TypeInstance)
      metricName := ""
      if vl.Identifier.PluginInstance != "" {
         metricName = fmt.Sprintf("%s.%s.%s.%s", vl.Identifier.Plugin, vl.Identifier.PluginInstance, vl.Identifier.Type, vl.Identifier.TypeInstance)
      } else {
         metricName = fmt.Sprintf("%s.%s.%s", vl.Identifier.Plugin, vl.Identifier.Type, vl.Identifier.TypeInstance)
      }
      formatValues(vl, ms, metricName)
      // }
      fatalIfErr(integration.Publish())
   }
}

func fatalIfErr(err error) {
   if err != nil {
      log.Fatal(err)
   }
}

func formatValues(vl *api.ValueList, ms *metric.Set, metricName string) {
   // formatTime(vl.Time)

   for _, val := range vl.Values {
      switch v := val.(type) {
      case api.Counter:
         // fields[i+1] = fmt.Sprintf("%d", v)
         ms.SetMetric(metricName, v, metric.GAUGE)
         ms.SetMetric("Entity", vl.Host, metric.ATTRIBUTE)
      case api.Gauge:
         // fields[i+1] = fmt.Sprintf("%.15g", v)
         ms.SetMetric(metricName, v, metric.GAUGE)
         ms.SetMetric("Entity", vl.Host, metric.ATTRIBUTE)
      case api.Derive:
         // fields[i+1] = fmt.Sprintf("%d", v)
         ms.SetMetric(metricName, v, metric.GAUGE)
         ms.SetMetric("Entity", vl.Host, metric.ATTRIBUTE)
      default:
         ms.SetMetric(metricName, v, metric.ATTRIBUTE)
         ms.SetMetric("Entity", vl.Host, metric.ATTRIBUTE)
         // return "", fmt.Errorf("unexpected type %T", v)
      }
   }

}

func formatTime(t time.Time) string {
   if t.IsZero() {
      return "N"
   }

   return fmt.Sprintf("%.3f", float64(t.UnixNano())/1000000000.0)
}
