+++
title = "Explore"
keywords = ["explore", "loki", "logs"]
aliases = ["/docs/grafana/latest/features/explore/"]
weight = 90
+++

# Explore

Grafana's dashboard UI is all about building dashboards for visualization. Explore strips away the dashboard and panel options so that you can focus on the query. It helps you iterate until you have a working query and then think about building a dashboard.

> Refer to [Fine-grained access control]({{< relref "../enterprise/access-control/_index.md" >}}) in Grafana Enterprise to understand how you can control access with fine-grained permissions.

If you just want to explore your data and do not want to create a dashboard, then Explore makes this much easier. If your data source supports graph and table data, then Explore shows the results both as a graph and a table. This allows you to see trends in the data and more details at the same time. See also:

- [Query management in Explore]({{< relref "query-management.md" >}})
- [Logs integration in Explore]({{< relref "logs-integration.md" >}})
- [Trace integration in Explore]({{< relref "trace-integration.md" >}})
- [Inspector in Explore]({{< relref "explore-inspector.md" >}})

## Start exploring

> Refer to [Fine-grained access Control]({{< relref "../enterprise/access-control/_index.md" >}}) in Grafana Enterprise to understand how you can manage Explore with fine-grained permissions.

In order to access Explore, you must have an editor or an administrator role, unless the [viewers_can_edit option]({{< relref "../administration/configuration/#viewers_can_edit" >}}) is enabled. Refer to [Organization roles]({{< relref "../permissions/organization_roles.md" >}}) for more information on what each role has access to.

To access Explore:

1. Click on the Explore icon on the menu bar.

   {{< figure src="/static/img/docs/explore/access-explore-7-4.png" max-width= "650px" caption="Screenshot of the new Explore Icon" >}}

   An empty Explore tab opens.

   Alternately to start with an existing query in a panel, choose the Explore option from the Panel menu. This opens an Explore tab with the query from the panel and allows you to tweak or iterate in the query outside of your dashboard.

{{< figure src="/static/img/docs/explore/panel_dropdown-7-4.png" class="docs-image--no-shadow" max-width= "650px" caption="Screenshot of the new Explore option in the panel menu" >}}

1. Choose your data source from the dropdown in the top left. [Prometheus](https://grafana.com/oss/prometheus/) has a custom Explore implementation, the other data sources use their standard query editor.
1. In the query field, write your query to explore your data. There are three buttons beside the query field, a clear button (X), an add query button (+) and the remove query button (-). Just like the normal query editor, you can add and remove multiple queries.

## Split and compare

The split view provides an easy way to compare graphs and tables side-by-side or to look at related data together on one page.

To open the split view:

1. Click the split button to duplicate the current query and split the page into two side-by-side queries.

It is possible to select another data source for the new query which for example, allows you to compare the same query for two different servers or to compare the staging environment to the production environment.

{{< figure src="/static/img/docs/explore/explore_split-7-4.png" max-width= "950px" caption="Screenshot of Explore option in the panel menu" >}}

In split view, timepickers for both panels can be linked (if you change one, the other gets changed as well) by clicking on one of the time-sync buttons attached to the timepickers. Linking of timepickers helps with keeping the start and the end times of the split view queries in sync. It ensures that you’re looking at the same time interval in both split panels.

To close the newly created query, click on the Close Split button.

## Navigate between Explore and a dashboard

To help accelerate workflows that involve regularly switching from Explore to a dashboard and vice-versa, Grafana provides you with the ability to return to the origin dashboard after navigating to Explore from the panel's dropdown.

After you've navigated to Explore, you should notice a "Back" button in the Explore toolbar. Simply click it to return to the origin dashboard. To bring changes you make in Explore back to the dashboard, click the arrow next to the button to reveal a "Return to panel with changes" menu item.

{{< figure src="/static/img/docs/explore/explore_return_dropdown-7-4.png" class="docs-image--no-shadow" max-width= "400px" caption="Screenshot of the expanded explore return dropdown" >}}

## Share shortened link

> **Note:** Available in Grafana 7.3 and later versions.

The Share shortened link capability allows you to create smaller and simpler URLs of the format /goto/:uid instead of using longer URLs with query parameters. To create a shortened link to the executed query, click the **Share** option in the Explore toolbar. A shortened link that is never used will automatically get deleted after seven (7) days.
