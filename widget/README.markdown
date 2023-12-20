# Weather Widget

Serve weather information, for use with [Widget Construction Set](https://wd.gt/widget_construction_set.html).

- Displays temperature, humidity, wind gust, and rainfall.
- Changes metric display colours based on specified thresholds.
- Polls Prometheus periodically to get latest values.
- Uses the [widget.json](https://wd.gt/) format.

## Using

Add a config file in a toml format:

``` toml
id = "melbourne"
name = "Melbourne Weather"
description = "Weather measurements for Melbourne, VIC, 3000"
token = "s3cr3t"
widget_url      = "https://graphs.domain.example/grafana/d/abc123/weather"
prometheus_url = "https://username:password@graphs.domain.example/prometheus/"

[metrics.temperature]
display_unit = "°"
prometheus_query = "outdoor_temperature_celsius"
levels = { "base" = 0, "low" = 18, "medium" = 27, "high" = 33 }

[metrics.humidity]
display_unit = "%"
prometheus_query = "outdoor_humidity_percentage"
levels = { "base" = 0, "low" = 20, "medium" = 80, "high" = 90 }

[metrics.wind_gust]
display_unit     = " km/h"
prometheus_query = "outdoor_wind_speed_burst_kilometers_per_hour"

[metrics.rainfall]
display_unit     = "mm"
prometheus_query = "delta(outdoor_rain_millimetres[24h])"
```

Then run it:

```
./weather_widget -c config.toml
```

Finally, fetch the JSON:

```
curl -v http://localhost:10002/widgets/melbourne?token=s3cr3t
```

Use the same URL to add a widget in [Widget Construction Set](https://wd.gt/widget_construction_set.html).

## Developing

Run the tests:

```bash
make
```

Build binaries:

```bash
make build
```
