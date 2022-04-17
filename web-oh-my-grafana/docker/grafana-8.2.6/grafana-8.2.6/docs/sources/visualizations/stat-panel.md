+++
title = "Stat"
description = "Stat panel documentation"
keywords = ["grafana", "docs", "stat panel"]
aliases = ["/docs/grafana/latest/features/panels/stat/", "/docs/grafana/latest/features/panels/singlestat/", "/docs/grafana/latest/reference/singlestat/", "/docs/grafana/latest/panels/visualizations/stat-panel/"]
weight = 900
+++

# Stat

The Stat panel visualization shows a one large stat value with an optional graph sparkline. You can control the background or value color using thresholds.

{{< figure src="/static/img/docs/v66/stat_panel_dark3.png" max-width="1025px" caption="Stat panel" >}}

> **Note:** This panel replaces the Singlestat panel, which was deprecated in Grafana 7.0 and removed in Grafana 8.0.

By default, the Stat panel displays one of the following:

- Just the value for a single series or field.
- Both the value and name for multiple series or fields.

You can use the **Text mode** to control whether the text is displayed or not.

Example screenshot:

{{< figure src="/static/img/docs/v71/stat-panel-text-modes.png" max-width="1025px" caption="Stat panel" >}}

## Automatic layout adjustment

The panel automatically adjusts the layout depending on available width and height in the dashboard. It automatically hides the graph (sparkline) if the panel becomes too small.

## Value options

Use the following options to refine how your visualization displays the value:

### Show

Choose how Grafana displays your data.

#### Calculate

Show a calculated value based on all rows.

- **Calculation -** Select a reducer function that Grafana will use to reduce many fields to a single value. For a list of available calculations, refer to [List of calculations]({{< relref "../panels/calculations-list.md" >}}).
- **Fields -** Select the fields display in the panel.

#### All values

Show a separate stat for every row. If you select this option, then you can also limit the number of rows to display.

- **Limit -** The maximum number of rows to display. Default is 5,000.
- **Fields -** Select the fields display in the panel.

## Stat styles

Style your visualization.

### Orientation

Choose a stacking direction.

- **Auto -** Grafana selects what it thinks is the best orientation.
- **Horizontal -** Bars stretch horizontally, left to right.
- **Vertical -** Bars stretch vertically, top to bottom.

### Text mode

You can use the Text mode option to control what text the panel renders. If the value is not important, only the name and color is, then change the **Text mode** to **Name**. The value will still be used to determine color and is displayed in a tooltip.

- **Auto -** If the data contains multiple series or fields, show both name and value.
- **Value -** Show only value, never name. Name is displayed in the hover tooltip instead.
- **Value and name -** Always show value and name.
- **Name -** Show name instead of value. Value is displayed in the hover tooltip.
- **None -** Show nothing (empty). Name and value are displayed in the hover tooltip.

### Color mode

Select a color mode.

- **Value -** Colors only the value and graph area.
- **Background -** Colors the background as well.

### Graph mode

Select a graph and splarkline mode.

- **None -** Hides the graph and only shows the value.
- **Area -** Shows the area graph below the value. This requires that your query returns a time column.

### Text alignment

Choose an alignment mode.

- **Auto -** If only a single value is shown (no repeat), then the value is centered. If multiple series or rows are shown, then the value is left-aligned.
- **Center -** Stat value is centered.

## Text size

Adjust the sizes of the gauge text.

- **Title -** Enter a numeric value for the gauge title size.
- **Value -** Enter a numeric value for the gauge value size.
