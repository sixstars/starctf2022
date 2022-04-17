+++
title = "Enhanced LDAP Integration"
description = "Grafana Enhanced LDAP Integration Guide "
keywords = ["grafana", "configuration", "documentation", "ldap", "active directory", "enterprise"]
weight = 600
+++

# Enhanced LDAP integration

The enhanced LDAP integration adds additional functionality on top of the [LDAP integration]({{< relref "../auth/ldap.md" >}}) available in the open source edition of Grafana.

> Enhanced LDAP integration is only available in [Grafana Cloud Advanced](https://grafana.com/docs/grafana-cloud/) and in [Grafana Enterprise]({{< relref "../enterprise" >}}).

> Refer to [Fine-grained access control]({{< relref "../enterprise/access-control/_index.md" >}}) in Grafana Enterprise to understand how you can control access with fine-grained permissions.

## LDAP group synchronization for teams

{{< figure src="/static/img/docs/enterprise/team_members_ldap.png" class="docs-image--no-shadow docs-image--right" max-width= "600px" >}}

With enhanced LDAP integration, you can set up synchronization between LDAP groups and teams. This enables LDAP users that are members
of certain LDAP groups to automatically be added or removed as members to certain teams in Grafana.

Grafana keeps track of all synchronized users in teams, and you can see which users have been synchronized from LDAP in the team members list, see `LDAP` label in screenshot.
This mechanism allows Grafana to remove an existing synchronized user from a team when its LDAP group membership changes. This mechanism also allows you to manually add
a user as member of a team, and it will not be removed when the user signs in. This gives you flexibility to combine LDAP group memberships and Grafana team memberships.

[Learn more about team sync.]({{< relref "team-sync.md">}})

<div class="clearfix"></div>

## Active LDAP synchronization

In the open source version of Grafana, user data from LDAP is synchronized only during the login process when authenticating using LDAP.

With active LDAP synchronization, available in Grafana Enterprise v6.3+, you can configure Grafana to actively sync users with LDAP servers in the background. Only users that have logged into Grafana at least once are synchronized.

Users with updated role and team membership will need to refresh the page to get access to the new features.

Removed users are automatically logged out and their account disabled. These accounts are displayed in the Server Admin > Users page with a `disabled` label. Disabled users keep their custom permissions on dashboards, folders, and data sources, so if you add them back in your LDAP database, they have access to the application with the same custom permissions as before.

```bash
[auth.ldap]
...

# You can use the Cron syntax or several predefined schedulers -
# @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 0 1 1 *
# @monthly               | Run once a month, midnight, first of month | 0 0 0 1 * *
# @weekly                | Run once a week, midnight between Sat/Sun  | 0 0 0 * * 0
# @daily (or @midnight)  | Run once a day, midnight                   | 0 0 0 * * *
# @hourly                | Run once an hour, beginning of hour        | 0 0 * * * *
sync_cron = "0 0 1 * * *" # This is default value (At 1 am every day)
# This cron expression format uses 6 space-separated fields (including seconds), for example
# sync_cron = "* */10 * * * *"
# This will run the LDAP Synchronization every 10th minute, which is also the minimal interval between the Grafana sync times i.e. you cannot set it for every 9th minute

# You can also disable active LDAP synchronization
active_sync_enabled = true # enabled by default
```

Single bind configuration (as in the [Single bind example]({{< relref "../auth/ldap.md#single-bind-example">}})) is not supported with active LDAP synchronization because Grafana needs user information to perform LDAP searches.
