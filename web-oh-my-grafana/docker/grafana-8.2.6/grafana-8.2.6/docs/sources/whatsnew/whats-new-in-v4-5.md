+++
title = "What's new in Grafana v4.5"
description = "Feature and improvement highlights for Grafana v4.5"
keywords = ["grafana", "new", "documentation", "4.5", "release notes"]
aliases = ["/docs/grafana/latest/guides/whats-new-in-v4-5/"]
weight = -12
[_build]
list = false
+++

# What's new in Grafana v4.5

## Highlights

### New prometheus query editor

The new query editor has full syntax highlighting. As well as auto complete for metrics, functions, and range vectors. There are also integrated function docs right from the query editor!

{{< figure src="/static/img/docs/v45/prometheus_query_editor_still.png" class="docs-image--block" animated-gif="/static/img/docs/v45/prometheus_query_editor.gif" >}}

### Elasticsearch: Add ad-hoc filters from the table panel

{{< figure src="/static/img/docs/v45/elastic_ad_hoc_filters.png" class="docs-image--block" >}}

### Table cell links!

Create column styles that turn cells into links that use the value in the cell (or other row values) to generate a URL to another dashboard or system:
![](/static/img/docs/v45/table_links.jpg)

### Query Inspector

Query Inspector is a new feature that shows query requests and responses. This can be helpful if a graph is not shown or shows something very different than what you expected.
For more information about query inspector, refer to [using grafanas query inspector to troubleshoot issues](https://community.grafana.com/t/using-grafanas-query-inspector-to-troubleshoot-issues/2630).
![](/static/img/docs/v45/query_inspector.png)

## Changelog

### New Features

- **Table panel**: Render cell values as links that can have an URL template that uses variables from current table row. [#3754](https://github.com/grafana/grafana/issues/3754)
- **Elasticsearch**: Add ad hoc filters directly by clicking values in table panel [#8052](https://github.com/grafana/grafana/issues/8052).
- **MySQL**: New rich query editor with syntax highlighting
- **Prometheus**: New rich query editor with syntax highlighting, metric and range auto complete and integrated function docs. [#5117](https://github.com/grafana/grafana/issues/5117)

### Enhancements

- **GitHub OAuth**: Support for GitHub organizations with 100+ teams. [#8846](https://github.com/grafana/grafana/issues/8846), thx [@skwashd](https://github.com/skwashd)
- **Graphite**: Calls to Graphite API /metrics/find now include panel or dashboard time range (from and until) in most cases, [#8055](https://github.com/grafana/grafana/issues/8055)
- **Graphite**: Added new graphite 1.0 functions, available if you set version to 1.0.x in data source settings. New Functions: mapSeries, reduceSeries, isNonNull, groupByNodes, offsetToZero, grep, weightedAverage, removeEmptySeries, aggregateLine, averageOutsidePercentile, delay, exponentialMovingAverage, fallbackSeries, integralByInterval, interpolate, invert, linearRegression, movingMin, movingMax, movingSum, multiplySeriesWithWildcards, pow, powSeries, removeBetweenPercentile, squareRoot, timeSlice, closes [#8261](https://github.com/grafana/grafana/issues/8261)
- **Elasticsearch**: Ad-hoc filters now use query phrase match filters instead of term filters, works on non keyword/raw fields [#9095](https://github.com/grafana/grafana/issues/9095).

### Breaking change

- **InfluxDB/Elasticsearch**: The panel and data source option named "Group by time interval" is now named "Min time interval" and does now always define a lower limit for the auto group by time. Without having to use `>` prefix (that prefix still works). This should in theory have close to zero actual impact on existing dashboards. It does mean that if you used this setting to define a hard group by time interval of, say "1d", if you zoomed to a time range wide enough the time range could increase above the "1d" range as the setting is now always considered a lower limit.

This option is now renamed (and moved to Options sub section above your queries):
![image|519x120](upload://ySjHOVpavV6yk9LHQxL9nq2HIsT.png)

Data source selection and options and help are now above your metric queries.
![image|690x179](upload://5kNDxKgMz1BycOKgG3iWYLsEVXv.png)

### Minor Changes

- **InfluxDB**: Change time range filter for absolute time ranges to be inclusive instead of exclusive [#8319](https://github.com/grafana/grafana/issues/8319), thx [@Oxydros](https://github.com/Oxydros)
- **InfluxDB**: Added parenthesis around tag filters in queries [#9131](https://github.com/grafana/grafana/pull/9131)

## Bug Fixes

- **Modals**: Maintain scroll position after opening/leaving modal [#8800](https://github.com/grafana/grafana/issues/8800)
- **Templating**: You cannot select data source variables as data source for other template variables [#7510](https://github.com/grafana/grafana/issues/7510)
