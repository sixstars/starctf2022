+++
title = "Permissions"
description = "Permissions"
keywords = ["grafana", "configuration", "documentation", "admin", "users", "datasources", "permissions"]
aliases = ["/docs/grafana/latest/permissions/overview/"]
weight = 50
+++

# Permissions

> Refer to [Fine-grained access Control]({{< relref "../enterprise/access-control/_index.md" >}}) in Grafana Enterprise for managing access with fine-grained permissions.

What you can do in Grafana is defined by the _permissions_ associated with your user account.

There are three types of permissions:

- Permissions granted as a Grafana Server Admin
- Permissions associated with your role in an organization
- Permissions granted to a specific folder or dashboard

You can be granted permissions based on:

- Grafana Server Admin status.
- Organization role (Admin, Editor, or Viewer).
- Folder or dashboard permissions assigned to your team (Admin, Editor, or Viewer).
- Folder or dashboard permissions assigned to your user account (Admin, Editor, or Viewer).
- (Grafana Enterprise) Data source permissions. For more information, refer to [Data source permissions]({{< relref "../enterprise/datasource_permissions.md" >}}) in [Grafana Enterprise]({{< relref "../enterprise" >}}).
- (Grafana Cloud) Grafana Cloud has additional roles. For more information, refer to [Grafana Cloud roles and permissions](/docs/grafana-cloud/cloud-portal/cloud-roles/).

If you are running Grafana Enterprise, you can grant access by using fine-grained roles and permissions, refer to [Fine-grained access Control]({{< relref "../enterprise/access-control/_index.md" >}}) for more information.

## Grafana Server Admin role

Grafana server administrators have the **Grafana Admin** flag enabled on their account. They can access the **Server Admin** menu and perform the following tasks:

- Manage users and permissions.
- Create, edit, and delete organizations.
- View server-wide settings that are set in the [Configuration]({{< relref "../administration/configuration.md" >}}) file.
- View Grafana server stats, including total users and active sessions.
- Upgrade the server to Grafana Enterprise.

> **Note:** This role does not exist in Grafana Cloud.

## Organization roles

Users can belong to one or more organizations. A user's organization membership is tied to a role that defines what the user is allowed to do in that organization. For more information, refer to [Organization roles]({{< relref "../permissions/organization_roles.md" >}}).

## Dashboard and folder permissions

Dashboard and folder permissions allow you to remove the default role based permissions for Editors and Viewers and assign permissions to specific users and teams. Learn more about [Dashboard and folder permissions]({{< relref "dashboard-folder-permissions.md" >}}).

## Data source permissions

Per default, a data source in an organization can be queried by any user in that organization. For example a user with `Viewer` role can still
issue any possible query to a data source, not just those queries that exist on dashboards he/she has access to.

Data source permissions allows you to change the default permissions for data sources and restrict query permissions to specific **Users** and **Teams**. For more information, refer to [Data source permissions]({{< relref "../enterprise/datasource_permissions.md" >}}) in [Grafana Enterprise]({{< relref "../enterprise" >}}).
