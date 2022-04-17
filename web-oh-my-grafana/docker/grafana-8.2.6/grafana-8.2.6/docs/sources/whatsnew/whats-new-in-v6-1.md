+++
title = "What's new in Grafana v6.1"
description = "Feature and improvement highlights for Grafana v6.1"
keywords = ["grafana", "new", "documentation", "6.1", "release notes"]
aliases = ["/docs/grafana/latest/guides/whats-new-in-v6-1/"]
weight = -20
[_build]
list = false
+++

# What's new in Grafana v6.1

## Highlights

### Ad hoc Filtering for Prometheus

{{< figure class="float-right"  max-width="30%" src="/static/img/docs/v61/prometheus-ad-hoc.gif" caption="Ad-hoc filters variable for Prometheus" >}}

The ad hoc filter feature allows you to create new key/value filters on the fly with autocomplete for both key and values. The filter condition is then automatically applied to all queries on the dashboard. This makes it easier to explore your data in a dashboard without changing queries and without having to add new template variables.

Other timeseries databases with label-based query languages have had this feature for a while. Recently Prometheus added support for fetching label names from their API and thanks to [Mitsuhiro Tanda](https://github.com/mtanda) implementing it in Grafana, the Prometheus data source finally supports ad hoc filtering.

Support for fetching a list of label names was released in Prometheus v2.6.0 so that is a requirement for this feature to work in Grafana.

### Permissions: Editors can own dashboards, folders and teams they create

When the dashboard folders feature and permissions system was released in Grafana 5.0, users with the editor role were not allowed to administrate dashboards, folders or teams. In the 6.1 release, we have added a configuration option that can change the default permissions so that editors are admins for any Dashboard, Folder or Team they create.

This feature also adds a new Team permission that can be assigned to any user with the editor or viewer role and enables that user to add other users to the Team.

We believe that this is more in line with the Grafana philosophy, as it will allow teams to be more self-organizing. This option will be made permanent if it gets positive feedback from the community so let us know what you think in the [issue on GitHub](https://github.com/grafana/grafana/issues/15590).

To turn this feature on add the following [configuration option](/administration/configuration/#editors-can-admin) to your Grafana ini file in the `users` section and then restart the Grafana server:

```ini
[users]
editors_can_admin = true
```

### List and revoke of user auth tokens in the API

As the first step of a feature to be able to list a user's signed in devices/sessions and to be able log out those devices from the Grafana UI, support has been added to the [API to list and revoke user authentication tokens](/http_api/admin/#auth-tokens-for-user).

### Minor Features and Fixes

This release contains a lot of small features and fixes:

- A new keyboard shortcut `d l` toggles all Graph legends in a dashboard.
- A small bug fix for Elasticsearch - template variables in the alias field now work properly.
- Some new capabilities have been added for data source plugins that will be of interest to plugin authors:
  - a new OAuth pass-through option.
  - it is now possible to add user details to requests sent to the dataproxy.
- Heatmap and Explore fixes.

Check out the [CHANGELOG.md](https://github.com/grafana/grafana/blob/master/CHANGELOG.md) file for a complete list of new features, changes, and bug fixes.

A huge thanks to our community for all the reported issues, bug fixes and feedback.
