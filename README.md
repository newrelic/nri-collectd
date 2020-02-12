# New Relic Infra Integration for CollectD

Monitor and report metrics for configured CollectD devices

## Disclaimer

New Relic has open-sourced this integration to enable monitoring of this technology. This integration is provided AS-IS WITHOUT WARRANTY OR SUPPORT, although you can report issues and contribute to this integration via GitHub. Support for this integration is available with an [Expert Services subscription](newrelic.com/expertservices).

## Prerequsites

The CollectD client must be running on the same host as the infra agent on which the nr-collectd-plugin is installed.
The CollectD client must be configured to send metric out using the CollectD "network" plugin. Usually the collectd.conf file must be updated to add/uncomment this snippet.

```sh bash
    <Plugin network>
        Server "127.0.0.1"
    </Plugin>
```

At this time, customization of the network plugin is not supported. So no authetication or change of default UDP port.

## Configuration
* port - UDP port to listen for metrics on.
* dim - Reports output as [dimensional metrics](https://docs.newrelic.com/docs/data-ingest-apis/get-data-new-relic/metric-api/introduction-metric-api) [true | false]
* interval - Interval to report dimensional metrics, formatted in golang [time.Duration](https://golang.org/pkg/time/#Duration). **NOTE: Only used if dim is set to true, otherwise the interval is set within /var/db/newrelic-infra/custom-integrations/collectd-plugin-definition.yml**
* key - Insights Insert API Key - Only used if dim is set to true (required)

## Test the plugin binary from the command line

Before configuring this plugin with New Relic Infrastructure agent, test it by executing it from the command line. Run with the help option to learn about the command line arguments that can be passed to this plugin:

```sh bash
./nri-collectd -help
```

For example, use the pretty argument for a nice looking output. At this time, this plugin has no mandatory arguments, so passing zero arguments is fine.

```sh bash
./nri-collectd -pretty
```


## Installation

Create a copy of the sample configuration and edit as needed
```sh bash

cp collectd-plugin-config.yml.sample collectd-plugin-config

```

Install the CollectD plugin

```sh bash

cp nri-collectd collectd-plugin-definition.yml /var/db/newrelic-infra/custom-integrations/

cp collectd-plugin-config.yml  /etc/newrelic-infra/integrations.d/

```

Restart the infrastructure agent

```sh bash
sudo systemctl stop newrelic-infra | sudo service newrelic-infra stop

sudo systemctl start newrelic-infra | sudo service newrelic-infra start
```

## Compatibility

* Supported OS: Linux
* collectd-plugin versions: 1.0

## Dashboarding
