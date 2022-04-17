+++
description = "Guide for getting started with Grafana and MS SQL Server"
keywords = ["grafana", "intro", "guide", "started", "SQL", "MSSQL"]
aliases = ["/docs/grafana/latest/guides/gettingstarted","/docs/grafana/latest/guides/getting_started"]
draft = true
weight = 400
+++

# Getting started with Grafana and MS SQL Server

Microsoft SQL Server is a popular relational database management system that is widely used in development and production environments. This topic walks you through the steps to create a series of dashboards in Grafana to display metrics from a MS SQL Server database. You can also configure the MS SQL Server data source on a [Grafana Cloud](https://grafana.com/docs/grafana-cloud/) instance without having to host Grafana yourself.

{{< docs/shared "getting-started/first-step.md" >}}

> **Note:** You must install Grafana 5.1+ in order to use the integrated MS SQL data source.

## Step 2. Download MS SQL Server

MS SQL Server can be installed on Windows or Linux operating systems and also on Docker containers. Refer to the [MS SQL Server downloads page](https://www.microsoft.com/en-us/sql-server/sql-server-downloads), for a complete list of all available options.

## Step 3. Install MS SQL Server

You can install MS SQL Server on the host running Grafana or on a remote server. To install the software from the [downloads page](https://www.microsoft.com/en-us/sql-server/sql-server-downloads), follow their setup prompts.

If you are on a Windows host but want to use Grafana and MS SQL data source on a Linux environment, refer to the [WSL to set up your Grafana development environment](https://grafana.com/blog/2021/03/03/.how-to-set-up-a-grafana-development-environment-on-a-windows-pc-using-wsl). This will allow you to leverage the resources available in [grafana/grafana](https://github.com/grafana/grafana) GitHub repository. Here you will find a collection of supported data sources, including MS SQL Server, along with test data and pre-configured dashboards for use.

## Step 4. Adding the MS SQL data source

To add MS SQL Server data source:

1. In the Grafana side menu, hover your cursor over the **Configuration** (gear) icon and then click **Data Sources**.
1. Filter by `mssql` and select the **Microsoft SQL Server** option.
1. Click **Add data source** in the top right header to open the configuration page.
1. Enter the information specified in the table below, then click **Save & Test**.

| Name       | Description                                                                                                           |
| ---------- | --------------------------------------------------------------------------------------------------------------------- |
| `Name`     | The data source name. This is how you refer to the data source in panels and queries.                                 |
| `Host`     | The IP address/hostname and optional port of your MS SQL instance. If port is omitted, the default 1433 will be used. |
| `Database` | Name of your MS SQL database.                                                                                         |
| `User`     | Database user's login/username.                                                                                       |
| `Password` | Database user's password.                                                                                             |

For installations from the [grafana/grafana](https://github.com/grafana/grafana/tree/main) repository, `gdev-mssql` data source is available. Once you add this data source, you can use the `Datasource tests - MSSQL` dashboard with three panels showing metrics generated from a test database.

<img src="/static/img/docs/getting-started/gdev-sql-dashboard.png" class="no-shadow" width="700px">

Optionally, play around this dashboard and customize it to:

- Create different panels.
- Change titles for panels.
- Change frequency of data polling.
- Change the period for which the data is displayed.
- Rearrange and resize panels.

## Step 5. Start building dashboards

Now that you have gained some idea of using the pre-packaged MS SQL data source and some test data, the next step is to setup your own instance of MS SQL Server database and data your development or sandbox area. In the previous steps, if you followed along the path of deploying your own instance of MS SQL Server, you are already on your way.

To fetch data from your own instance of MS SQL Server, add the data source using instructions in Step 4 of this topic. In Grafana [Explore]({{< relref "../explore/_index.md" >}}) build queries to experiment with the metrics you want to monitor.

Once you have a curated list of queries, create [dashboards]({{< relref "../dashboards/_index.md" >}}) to render metrics from the SQL Server database. For troubleshooting, user permissions, known issues, and query examples, refer to [Using Microsoft SQL Server in Grafana]({{< relref "../datasources/mssql.md" >}}).
