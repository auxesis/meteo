# collectd plugins

All these plugins should run under Python 3.

To set up dependencies for these collectd plugins, run:

```
pip install -r requirements.txt
```

## `collectd_read_sma.py`

Uses `pysma` to scrape metrics from SMA inverters, and output them in the collectd plugin format.

| Argument     | Description | Example value  |
| ------------ | ----------- | -------------- |
| `--address`  | Where to find the SMA inverter on the network.           | `192.168.1.30`    |
| `--password` | Password to access the user account on the SMA inverter. | `abc123-#(`       |
| `--host`     | Host to report the metrics to collectd as coming from.   | `my.fqdn.example` |

Run with no arguments to see all available options.

Example collectd configuration:

```
LoadPlugin exec

<Plugin "exec">
  Exec "user:group" "/usr/bin/python3" "/opt/meteo/plugins/collectd/collectd_read_sma.py" "--address=192.168.1.16" "--password=S3cr3t" "--host=my.fqdn.example"
</Plugin>
```

## `collectd_read_bom.py`

Report Bureau of Meteorology metrics from a weather station in the collectd plugin format.

| Argument       | Description | Example value  |
| -------------- | ----------- | -------------- |
| `--host`       | Host to report the metrics to collectd as coming from.    | `my.fqdn.example` |
| `--area-id`    | BOM area id of the weather station.                       | `IDN60801`        |
| `--station-id` | BOM station id of the weather station.                    | `95682`           |

Run with no arguments to see all available options.

You can work out the area id and station id from looking at the URL of the latest weather observations for that weather station. For example, http://www.bom.gov.au/products/IDN60801/IDN60801.95682.shtml for the Mount Hope weather station has the area id of `IDN60801`, and the station id of `95682`.

Example collectd configuration:

```
LoadPlugin exec

<Plugin "exec">
  Exec "user:augroup" "/usr/bin/python3" "/opt/meteo/plugins/collectd/collectd_read_bom.py" "--host=mount-hope.example" "--area-id=IDN60801" "--station-id=95682"
</Plugin>
```

## `collectd_read_digitemp.py`

Output temperature reading from `digitemp_DS9097` in the collectd plugin format.

| Argument       | Description | Example value  |
| -------------- | ----------- | -------------- |
| `--host`       | Host to report the metrics to collectd as coming from.    | `my.fqdn.example` |
| `--plugin`     | Plugin to report the metrics as coming from.              | `bedroom_1`       |

Example collectd configuration:

```
LoadPlugin exec

<Plugin "exec">
  Exec "user:group" "/usr/bin/python3" "/opt/meteo/plugins/collectd/collectd_read_digitemp.py" "--host=my.fqdn.example" "--plugin=living_room"
</Plugin>
```
