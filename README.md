# New Relic Infra Integration for CollectD

Monitor and report metrics for configured CollectD devices

## Disclaimer

New Relic has open-sourced this integration to enable monitoring of this technology. This integration is provided AS-IS WITHOUT WARRANTY OR SUPPORT, although you can report issues and contribute to this integration via GitHub. Support for this integration is available with an [Expert Services subscription](https://newrelic.com/expertservices).

## Prerequsites

The CollectD client must be running on the same host as the infra agent on which the nr-collectd-plugin is installed.
The CollectD client must be configured to send metric out using the CollectD "network" plugin. Usually the collectd.conf file must be updated to add/uncomment this snippet.

```sh bash
    <Plugin network>
        Server "127.0.0.1"
    </Plugin>
```

At this time, customization of the network plugin is not supported. So no authetication or change of default UDP port.

## Configuration explained

* port - UDP port to listen for metrics on.
* dim - Reports output as [dimensional metrics](https://docs.newrelic.com/docs/data-ingest-apis/get-data-new-relic/metric-api/introduction-metric-api) [true | false]
* interval - Interval to report dimensional metrics, formatted in golang [time.Duration](https://golang.org/pkg/time/#Duration). **NOTE: Only used if dim is set to true, otherwise the interval is set within /var/db/newrelic-infra/custom-integrations/collectd-plugin-definition.yml**
* key - Insights Insert API Key - Only used if dim is set to true (required)

## Sample collectd configuration file

```sh bash
integration_name: com.newrelic.collectd-plugin

instances:
  - name: sample-collectd
    command: metrics
    arguments:
      port: 25826
      key: "Your Insights Inserts API Key"
      #interval: "30s"
      dim: true
    labels:
      role: collectd
```

## Test the plugin binary from the command line

1. [Download](https://github.com/newrelic/nri-collectd/releases) the latest release of collectd plugin.
1. Unzip and change directory to the bin folder.
1. Before configuring this plugin with New Relic Infrastructure agent, test it by executing it from the command line. 
1. Run with the help option to learn about the command line arguments that can be passed to this plugin:

    ```sh bash
    ./nri-collectd -help
    ```

1. For example, use the pretty argument for a nice looking output. At this time, this plugin has no mandatory arguments, so passing zero arguments is fine.

    ```sh bash
    ./nri-collectd -pretty
    ```

## Install and Configure

1. Create collectd plugin config file

  ```sudo cp collectd-plugin-config.yml.sample collectd-plugin-config.yml```

1. Copy collectd plugin to integration folder

  ```sudo cp nri-collectd /var/db/newrelic-infra/custom-integrations/```

1. Copy collectd definition files to integration folder

  ```sudo cp collectd-plugin-definition.yml /var/db/newrelic-infra/custom-integrations/```

1. Copy collectd plugin config file integration folder

  ```sudo cp collectd-plugin-config.yml /etc/newrelic-infra/integrations.d/```

1. Note down port number used in the /etc/collectd/collectd.conf file network stanza. Use the same port number in the collectd-plugin-config.yml file

1. Refer **Sample collectd configuration file** section and edit collectd-plugin-config.yml in the folder /etc/newrelic-infra/integrations.d

1. To obtain Inserts key, go to your New Relic account ⇒ Insights ⇒ In the left navigation panel, click Manage data ⇒ From the top navigation, click API Keys ⇒ copy Inserts Key. 

1. Stop the infrastructure agent

  ```sudo systemctl stop newrelic-infra | sudo service newrelic-infra stop```

1. Start the infrastructure agent

  ```sudo systemctl start newrelic-infra | sudo service newrelic-infra start```

1. Check to see if nri-collectd plugin is running

  ```sudo ps -ef | grep nri-collectd```

1. You should start seeing metrics in the New Relic Insights table Metric

## Compatibility

* Supported OS: Linux
* collectd-plugin versions: 1.0

## Dashboarding

1. Run sample NRQLs to validate collectd metrics are flowing into New Relic

  ```SELECT count(*) FROM Metric SINCE 30 MINUTES AGO```

  ```FROM Metric SELECT uniques(Plugin)```

  ```FROM Metric SELECT uniques(metricName) WHERE Plugin ='cpu'```
  
  ```FROM Metric SELECT latest(apache%) FACET metricName SINCE 1 DAY AGO```

1. [Create Dashboards](https://docs.newrelic.com/docs/insights/use-insights-ui/manage-dashboards/create-edit-insights-dashboards) for your collectd metrics.
