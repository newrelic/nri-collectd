# New Relic Infra Integration for CollectD

Monitor and report metrics for configured CollectD devices


## Prerequsites

The CollectD client must be running on the same host as the infra agent on which the nr-collectd-plugin is installed.
The CollectD client must be configured to send metric out using the CollectD "network" plugin. Usually the collectd.conf file must be updated to add/uncomment this snippet.

```sh bash
    <Plugin network>
        Server "127.0.0.1"
    </Plugin>
```

At this time, customization of the network plugin is not supported. So no authetication or change of default UDP port.


## Test the plugin binary from the command line

Before configuring this plugin with New Relic Infrastructure agent, test it by executing it from the command line

Run with the help option to learn about the command line arguments that can be passed to this plugin

```sh bash
./bin/nr-collectd-plugin --help
```

For example, use the pretty argument for a nice looking output. At this time, this plugin has no mandatory arguments, so passing zero arguments is fine.

```sh bash
./bin/nr-collectd-plugin -pretty
```


## Installation

Install the CollectD plugin

```sh bash

cp -R bin /var/db/newrelic-infra/custom-integrations/

cp collectd-plugin-definition.yml /var/db/newrelic-infra/custom-integrations/

cp collectd-plugin-config.yml  /etc/newrelic-infra/integrations.d/

```

Restart the infrastructure agent

```sh bash
sudo systemctl stop newrelic-infra

sudo systemctl start newrelic-infra
```

## Compatibility

* Supported OS: Linux
* collectd-plugin versions: 1.0

## Dashboarding
