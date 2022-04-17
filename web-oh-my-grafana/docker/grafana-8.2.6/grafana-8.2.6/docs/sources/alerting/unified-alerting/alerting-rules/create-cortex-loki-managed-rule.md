+++
title = "Create Cortex or Loki managed alert rule"
description = "Create Cortex or Loki managed alerting rule"
keywords = ["grafana", "alerting", "guide", "rules", "create"]
weight = 400
+++

# Create a Cortex or Loki managed alerting rule

Grafana allows you to create alerting rules for an external Cortex or Loki instance.

## Before you begin

For Cortex and Loki data sources to work with Grafana 8.0 alerting, enable the ruler API by configuring their respective services.

**Loki** - The `local` rule storage type, default for the Loki data source, supports only viewing of rules. To edit rules, configure one of the other rule storage types.

**Cortex** - When configuring a Grafana Prometheus data source to point to Cortex, use the [legacy `/api/prom` prefix](https://cortexmetrics.io/docs/api/#path-prefixes), not `/prometheus`. The Prometheus data source supports both Cortex and Prometheus, and Grafana expects that both the [Query API](https://cortexmetrics.io/docs/api/#querier--query-frontend) and [Ruler API](https://cortexmetrics.io/docs/api/#ruler) are under the same URL. You cannot provide a separate URL for the Ruler API.

> **Note:** If you do not want to manage alerting rules for a particular Loki or Prometheus data source, go to its settings and clear the **Manage alerts via Alerting UI** checkbox.

## Add a Cortex or Loki managed alerting rule

1. In the Grafana menu, click the **Alerting** (bell) icon to open the Alerting page listing existing alerts.
1. Click **New alert rule**.
1. In Step 1, add the rule name, type, and storage location.
   - In **Rule name**, add a descriptive name. This name is displayed in the alert rule list. It is also the `alertname` label for every alert instance that is created from this rule.
   - From the **Rule type** drop-down, select **Cortex / Loki managed alert**.
   - From the **Select data source** drop-down, select an external Prometheus, an external Loki, or a Grafana Cloud data source.
   - From the **Namespace** drop-down, select an existing rule namespace. Otherwise, click **Add new** and enter a name to create a new one. Namespaces can contain one or more rule groups and only have an organizational purpose. For more information, see [Cortex or Loki rule groups and namespaces]({{< relref "./edit-cortex-loki-namespace-group.md" >}}).
   - From the **Group** drop-down, select an existing group within the selected namespace. Otherwise, click **Add new** and enter a name to create a new one. Newly created rules are appended to the end of the group. Rules within a group are run sequentially at a regular interval, with the same evaluation time.
     {{< figure src="/static/img/docs/alerting/unified/rule-edit-cortex-alert-type-8-0.png" max-width="550px" caption="Alert details" >}}
1. In Step 2, add the query to evaluate.
   - Enter a PromQL or LogQL expression. The rule fires if the evaluation result has at least one series with a value that is greater than 0. An alert is created for each series.
     {{< figure src="/static/img/docs/alerting/unified/rule-edit-cortex-query-8-0.png" max-width="550px" caption="Alert details" >}}
1. In Step 3, add conditions.
   - In the **For** text box, specify the duration for which the condition must be true before an alert fires. If you specify `5m`, the condition must be true for 5 minutes before the alert fires.
     > **Note:** Once a condition is met, the alert goes into the `Pending` state. If the condition remains active for the duration specified, the alert transitions to the `Firing` state, else it reverts to the `Normal` state.
1. In Step 4, add additional metadata associated with the rule.
   - Add a description and summary to customize alert messages. Use the guidelines in [Annotations and labels for alerting]({{< relref "./alert-annotation-label.md" >}}).
   - Add Runbook URL, panel, dashboard, and alert IDs.
   - Add custom labels.
1. To evaluate the rule and see what alerts it would produce, click **Preview alerts**. It will display a list of alerts with state and value of for each one.
1. Click **Save** to save the rule or **Save and exit** to save the rule and go back to the Alerting page.
