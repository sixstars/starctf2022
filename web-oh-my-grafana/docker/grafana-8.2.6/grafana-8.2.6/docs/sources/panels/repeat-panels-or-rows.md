+++
title = "Repeat panels or rows"
keywords = ["grafana", "templating", "documentation", "guide", "template", "variable", "repeat"]
aliases = ["/docs/grafana/latest/variables/repeat-panels-or-rows/"]
weight = 800
+++

# Repeat panels or rows

{{< docs/shared "panels/repeat-panels-intro.md" >}}

Template variables can be very useful to dynamically change your queries across a whole dashboard. If you want
Grafana to dynamically create new panels or rows based on what values you have selected, you can use the _Repeat_ feature.

## Grafana Play examples

You can see examples in the following dashboards:

- [Prometheus repeat](https://play.grafana.org/d/000000036/prometheus-repeat)
- [Repeated Rows Dashboard](https://play.grafana.org/d/000000153/repeat-rows)

## Repeating panels

If you have a variable with `Multi-value` or `Include all value` options enabled you can choose one panel and have Grafana repeat that panel
for every selected value. You find the _Repeat_ feature under the _General tab_ in panel edit mode.

The `direction` controls how the panels will be arranged.

By choosing `horizontal` the panels will be arranged side-by-side. Grafana will automatically adjust the width
of each repeated panel so that the whole row is filled. Currently, you cannot mix other panels on a row with a repeated
panel.

Set `Max per row` to tell grafana how many panels per row you want at most. It defaults to _4_ if you don't set anything.

By choosing `vertical` the panels will be arranged from top to bottom in a column. The width of the repeated panels will be the same as of the first panel (the original template) being repeated.

Only make changes to the first panel (the original template). To have the changes take effect on all panels you need to trigger a dynamic dashboard re-build.
You can do this by either changing the variable value (that is the basis for the repeat) or reload the dashboard.

> **Note:** Repeating panels require variables to have one or more items selected; you cannot repeat a panel zero times to hide it.

## Repeating rows

As seen above with the panels you can also repeat rows if you have variables set with `Multi-value` or
`Include all value` selection option.

To enable this feature you need to first add a new _Row_ using the _Add Panel_ menu. Then by hovering the row title and
clicking on the cog button, you will access the `Row Options` configuration panel. You can then select the variable
you want to repeat the row for.

It may be a good idea to use a variable in the row title as well.
