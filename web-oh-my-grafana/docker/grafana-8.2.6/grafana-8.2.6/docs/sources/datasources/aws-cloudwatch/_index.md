+++
title = "AWS CloudWatch"
description = "Guide for using CloudWatch in Grafana"
keywords = ["grafana", "cloudwatch", "guide"]
aliases = ["/docs/grafana/latest/datasources/cloudwatch"]
weight = 200
+++

# AWS CloudWatch data source

Grafana ships with built-in support for CloudWatch. Add it as a data source, then you are ready to build dashboards or use Explore with CloudWatch metrics and CloudWatch Logs.

This topic describes queries, templates, variables, and other configuration specific to the CloudWatch data source. For instructions on how to add a data source to Grafana, refer to [Add a data source]({{< relref "../add-a-data-source.md" >}}). Only users with the organization admin role can add data sources.

> **Note:** If you are having issues setting up the data source and Grafana is returning undescriptive errors, then check the log file located in /var/log/grafana/grafana.log).

## Cloudwatch settings

To access data source settings, hover your mouse over the **Configuration** (gear) icon, then click **Data Sources**, and then click the AWS Cloudwatch data source.

| Name                       | Description                                                                                                             |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `Name`                     | The data source name. This is how you refer to the data source in panels and queries.                                   |
| `Default`                  | Default data source means that it will be pre-selected for new panels.                                                  |
| `Default Region`           | Used in query editor to set region (can be changed on per query basis)                                                  |
| `Custom Metrics namespace` | Specify the CloudWatch namespace of Custom metrics                                                                      |
| `Auth Provider`            | Specify the provider to get credentials.                                                                                |
| `Credentials` profile name | Specify the name of the profile to use (if you use `~/.aws/credentials` file), leave blank for default.                 |
| `Assume Role Arn`          | Specify the ARN of the role to assume                                                                                   |
| `External ID`              | If you are assuming a role in another account, that has been created with an external ID, specify the external ID here. |

### X-Ray trace links

Link an X-Ray data source in the "X-Ray trace link" section of the configuration page to automatically add links in your logs when the log contains `@xrayTraceId` field.

![Trace link configuration](/static/img/docs/cloudwatch/xray-trace-link-configuration-8-2.png 'Trace link configuration')

The data source select will contain only existing data source instances of type X-Ray so in order to use this feature you need to have existing X-Ray data source already configured, see [X-Ray docs](https://grafana.com/grafana/plugins/grafana-x-ray-datasource/) for details.

The X-Ray link will then appear in the log details section which is accessible by clicking on the log row either in Explore or in dashboard [Logs panel]({{< relref "../../visualizations/logs-panel.md" >}}). To log the `@xrayTraceId` in your logs see the [AWS X-Ray documentation](https://docs.amazonaws.cn/en_us/xray/latest/devguide/xray-services.html). To provide the field to Grafana your log queries also have to contain the `@xrayTraceId` field, for example using query `fields @message, @xrayTraceId`.

![Trace link in log details](/static/img/docs/cloudwatch/xray-link-log-details-8-2.png 'Trace link in log details')

## Authentication

For authentication options and configuration details, see [AWS authentication]({{< relref "aws-authentication.md" >}}) topic.

## IAM policies

Grafana needs permissions granted via IAM to be able to read CloudWatch metrics
and EC2 tags/instances/regions. You can attach these permissions to IAM roles and
utilize Grafana's built-in support for assuming roles.

Here is a minimal policy example:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowReadingMetricsFromCloudWatch",
      "Effect": "Allow",
      "Action": [
        "cloudwatch:DescribeAlarmsForMetric",
        "cloudwatch:DescribeAlarmHistory",
        "cloudwatch:DescribeAlarms",
        "cloudwatch:ListMetrics",
        "cloudwatch:GetMetricStatistics",
        "cloudwatch:GetMetricData"
      ],
      "Resource": "*"
    },
    {
      "Sid": "AllowReadingLogsFromCloudWatch",
      "Effect": "Allow",
      "Action": [
        "logs:DescribeLogGroups",
        "logs:GetLogGroupFields",
        "logs:StartQuery",
        "logs:StopQuery",
        "logs:GetQueryResults",
        "logs:GetLogEvents"
      ],
      "Resource": "*"
    },
    {
      "Sid": "AllowReadingTagsInstancesRegionsFromEC2",
      "Effect": "Allow",
      "Action": ["ec2:DescribeTags", "ec2:DescribeInstances", "ec2:DescribeRegions"],
      "Resource": "*"
    },
    {
      "Sid": "AllowReadingResourcesForTags",
      "Effect": "Allow",
      "Action": "tag:GetResources",
      "Resource": "*"
    }
  ]
}
```

## Using the Query Editor

The CloudWatch data source can query data from both CloudWatch metrics and CloudWatch Logs APIs, each with its own specialized query editor. You select which API you want to query with using the query mode switch on top of the editor.

{{< figure src="/static/img/docs/v70/cloudwatch-metrics-query-field.png" max-width="800px" class="docs-image--left" caption="CloudWatch metrics query field" >}}
{{< figure src="/static/img/docs/v70/cloudwatch-logs-query-field.png" max-width="800px" class="docs-image--right" caption="CloudWatch Logs query field" >}}

## Using the Metric Query Editor

To create a valid query, you need to specify the namespace, metric name and at least one statistic. If `Match Exact` is enabled, you also need to specify all the dimensions of the metric you’re querying, so that the [metric schema](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/search-expression-syntax.html) matches exactly. If `Match Exact` is off, you can specify any number of dimensions by which you’d like to filter. Up to 100 metrics matching your filter criteria will be returned.

### Dynamic queries using dimension wildcards

You can monitor a dynamic list of metrics by using the asterisk (\*) wildcard for one or more dimension values.

{{< figure src="/static/img/docs/v65/cloudwatch-dimension-wildcard.png" max-width="800px" class="docs-image--right" caption="CloudWatch dimension wildcard" >}}

In this example, the query returns all metrics in the namespace `AWS/EC2` with a metric name of `CPUUtilization` and ANY value for the `InstanceId` dimension are queried. This can help you monitor metrics for AWS resources, like EC2 instances or containers. For example, when new instances are created as part of an auto scaling event, they will automatically appear in the graph without needing to track the new instance IDs. This capability is currently limited to retrieving up to 100 metrics.

Click on `Show Query Preview` to see the search expression that is automatically built to support wildcards. To learn more about search expressions, visit the [CloudWatch documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/search-expression-syntax.html). By default, the search expression is defined in such a way that the queried metrics must match the defined dimension names exactly. This means that in the example only metrics with exactly one dimension with name ‘InstanceId’ will be returned.

You can untoggle `Match Exact` to include metrics that have other dimensions defined. Disabling `Match Exact` also creates a search expression even if you don’t use wildcards. We simply search for any metric that matches at least the namespace, metric name, and all defined dimensions.

### Multi-value template variables

When defining dimension values based on multi-valued template variables, a search expression is used to query for the matching metrics. This enables the use of multiple template variables in one query and also allows you to use template variables for queries that have the `Match Exact` option disabled.

Search expressions are currently limited to 1024 characters, so your query may fail if you have a long list of values. We recommend using the asterisk (`*`) wildcard instead of the `All` option if you want to query all metrics that have any value for a certain dimension name.

The use of multi-valued template variables is only supported for dimension values. Using multi-valued template variables for `Region`, `Namespace`, or `Metric Name` is not supported.

### Metric math expressions

You can create new time series metrics by operating on top of CloudWatch metrics using mathematical functions. Arithmetic operators, unary subtraction and other functions are supported and can be applied to CloudWatch metrics. More details on the available functions can be found on [AWS Metric Math](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html)

As an example, if you want to apply arithmetic operations on a metric, you can do it by giving an id (a unique string) to the raw metric as shown below. You can then use this id and apply arithmetic operations to it in the Expression field of the new metric.

Please note that in the case you use the expression field to reference another query, like `queryA * 2`, it will not be possible to create an alert rule based on that query.

### Period

A period is the length of time associated with a specific Amazon CloudWatch statistic. Periods are defined in numbers of seconds, and valid values for period are 1, 5, 10, 30, or any multiple of 60.

If the period field is left blank or set to `auto`, then it calculates automatically based on the time range and [cloudwatch's retention policy](https://aws.amazon.com/about-aws/whats-new/2016/11/cloudwatch-extends-metrics-retention-and-new-user-interface/). The formula used is `time range in seconds / 2000`, and then it snaps to the next higher value in an array of predefined periods `[60, 300, 900, 3600, 21600, 86400]` after removing periods based on retention. By clicking `Show Query Preview` in the query editor, you can see what period Grafana used.

### Deep linking from Grafana panels to the CloudWatch console

{{< figure src="/static/img/docs/v65/cloudwatch-deep-linking.png" max-width="500px" class="docs-image--right" caption="CloudWatch deep linking" >}}

Left clicking a time series in the panel shows a context menu with a link to `View in CloudWatch console`. Clicking that link will open a new tab that will take you to the CloudWatch console and display all the metrics for that query. If you're not currently logged in to the CloudWatch console, the link will forward you to the login page. The provided link is valid for any account but will only display the right metrics if you're logged in to the account that corresponds to the selected data source in Grafana.

This feature is not available for metrics that are based on metric math expressions.

## Using the Logs Query Editor

To query CloudWatch Logs, select the region and up to 20 log groups which you want to query. Use the main input area to write your query in [CloudWatch Logs Query Language](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_QuerySyntax.html)

You can also write queries returning time series data by using the [`stats` command](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_Insights-Visualizing-Log-Data.html). When making `stats` queries in Explore, you have to make sure you are in Metrics Explore mode.

{{< figure src="/static/img/docs/v70/explore-mode-switcher.png" max-width="500px" class="docs-image--right" caption="Explore mode switcher" >}}

To the right of the query input field is a CloudWatch Logs Insights link that opens the CloudWatch Logs Insights console with your query. You can continue exploration there if necessary.

{{< figure src="/static/img/docs/v70/cloudwatch-logs-deep-linking.png" max-width="500px" class="docs-image--right" caption="CloudWatch Logs deep linking" >}}

### Using template variables

The CloudWatch data source supports use of template variables in queries.
For an introduction to templating and template variables, refer to the [Templating]({{< relref "../../variables/_index.md" >}}) documentation.

### Deep linking from Grafana panels to the CloudWatch console

{{< figure src="/static/img/docs/v70/cloudwatch-logs-deep-linking.png" max-width="500px" class="docs-image--right" caption="CloudWatch Logs deep linking" >}}
If you'd like to view your query in the CloudWatch Logs Insights console, simply click the `CloudWatch Logs Insights` button next to the query editor.
If you're not currently logged in to the CloudWatch console, the link will forward you to the login page. The provided link is valid for any account but will only display the right metrics if you're logged in to the account that corresponds to the selected data source in Grafana.

### Alerting

Since CloudWatch Logs queries can return numeric data, for example through the use of the `stats` command, alerts are supported.
For more information on Grafana alerts, refer to [Alerting]({{< relref "../../alerting/_index.md" >}}) documentation.

## Curated dashboards

The updated CloudWatch data source ships with pre-configured dashboards for five of the most popular AWS services:

- Amazon Elastic Compute Cloud `Amazon EC2`,
- Amazon Elastic Block Store `Amazon EBS`,
- AWS Lambda `AWS Lambda`,
- Amazon CloudWatch Logs `Amazon CloudWatch Logs`, and
- Amazon Relational Database Service `Amazon RDS`.

To import the pre-configured dashboards, go to the configuration page of your CloudWatch data source and click on the `Dashboards` tab. Click `Import` for the dashboard you would like to use. To customize the dashboard, we recommend saving the dashboard under a different name, because otherwise the dashboard will be overwritten when a new version of the dashboard is released.

{{< figure src="/static/img/docs/v65/cloudwatch-dashboard-import.png" caption="CloudWatch dashboard import" >}}

## Templated queries

Instead of hard-coding server, application, and sensor names in your metric queries, you can use variables. The variables are listed as dropdown select boxes at the top of the dashboard. These dropdowns make it easy to change the display of data in your dashboard.

For an introduction to templating and template variables, refer to the [Templating]({{< relref "../../variables/_index.md" >}}) documentation.

### Query variable

The CloudWatch data source provides the following queries that you can specify in the `Query` field in the Variable edit view. They allow you to fill a variable's options list with things like `region`, `namespaces`, `metric names` and `dimension keys/values`.

In place of `region` you can specify `default` to use the default region configured in the data source for the query,
e.g. `metrics(AWS/DynamoDB, default)` or `dimension_values(default, ..., ..., ...)`.

Read more about the available dimensions in the [CloudWatch Metrics and Dimensions Reference](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CW_Support_For_AWS.html).

| Name                                                                    | Description                                                                                                                                                                        |
| ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `regions()`                                                             | Returns a list of all AWS regions                                                                                                                                                  |
| `namespaces()`                                                          | Returns a list of namespaces CloudWatch support.                                                                                                                                   |
| `metrics(namespace, [region])`                                          | Returns a list of metrics in the namespace. (specify region or use "default" for custom metrics)                                                                                   |
| `dimension_keys(namespace)`                                             | Returns a list of dimension keys in the namespace.                                                                                                                                 |
| `dimension_values(region, namespace, metric, dimension_key, [filters])` | Returns a list of dimension values matching the specified `region`, `namespace`, `metric`, `dimension_key` or you can use dimension `filters` to get more specific result as well. |
| `ebs_volume_ids(region, instance_id)`                                   | Returns a list of volume ids matching the specified `region`, `instance_id`.                                                                                                       |
| `ec2_instance_attribute(region, attribute_name, filters)`               | Returns a list of attributes matching the specified `region`, `attribute_name`, `filters`.                                                                                         |
| `resource_arns(region, resource_type, tags)`                            | Returns a list of ARNs matching the specified `region`, `resource_type` and `tags`.                                                                                                |
| `statistics()`                                                          | Returns a list of all the standard statistics                                                                                                                                      |

For details about the metrics CloudWatch provides, please refer to the [CloudWatch documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).

#### Examples templated queries

Example dimension queries which will return list of resources for individual AWS Services:

| Query                                                                                                                         | Service          |
| ----------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `dimension_values(us-east-1,AWS/ELB,RequestCount,LoadBalancerName)`                                                           | ELB              |
| `dimension_values(us-east-1,AWS/ElastiCache,CPUUtilization,CacheClusterId)`                                                   | ElastiCache      |
| `dimension_values(us-east-1,AWS/Redshift,CPUUtilization,ClusterIdentifier)`                                                   | RedShift         |
| `dimension_values(us-east-1,AWS/RDS,CPUUtilization,DBInstanceIdentifier)`                                                     | RDS              |
| `dimension_values(us-east-1,AWS/S3,BucketSizeBytes,BucketName)`                                                               | S3               |
| `dimension_values(us-east-1,CWAgent,disk_used_percent,device,{"InstanceId":"$instance_id"})`                                  | CloudWatch Agent |
| `resource_arns(eu-west-1,elasticloadbalancing:loadbalancer,{"elasticbeanstalk:environment-name":["myApp-dev","myApp-prod"]})` | ELB              |
| `resource_arns(eu-west-1,elasticloadbalancing:loadbalancer,{"Component":["$service"],"Environment":["$environment"]})`        | ELB              |
| `resource_arns(eu-west-1,ec2:instance,{"elasticbeanstalk:environment-name":["myApp-dev","myApp-prod"]})`                      | EC2              |

## ec2_instance_attribute examples

### JSON filters

The `ec2_instance_attribute` query takes `filters` in JSON format.
You can specify [pre-defined filters of ec2:DescribeInstances](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html).
Note that the actual filtering takes place on Amazon's servers, not in Grafana.

Filters syntax:

```javascript
{ "filter_name1": [ "filter_value1" ], "filter_name2": [ "filter_value2" ] }
```

Example `ec2_instance_attribute()` query

```javascript
ec2_instance_attribute(us - east - 1, InstanceId, { 'tag:Environment': ['production'] });
```

### Selecting attributes

Only 1 attribute per instance can be returned. Any flat attribute can be selected (i.e. if the attribute has a single value and isn't an object or array). Below is a list of available flat attributes:

- `AmiLaunchIndex`
- `Architecture`
- `ClientToken`
- `EbsOptimized`
- `EnaSupport`
- `Hypervisor`
- `IamInstanceProfile`
- `ImageId`
- `InstanceId`
- `InstanceLifecycle`
- `InstanceType`
- `KernelId`
- `KeyName`
- `LaunchTime`
- `Platform`
- `PrivateDnsName`
- `PrivateIpAddress`
- `PublicDnsName`
- `PublicIpAddress`
- `RamdiskId`
- `RootDeviceName`
- `RootDeviceType`
- `SourceDestCheck`
- `SpotInstanceRequestId`
- `SriovNetSupport`
- `SubnetId`
- `VirtualizationType`
- `VpcId`

Tags can be selected by prepending the tag name with `Tags.`

Example `ec2_instance_attribute()` query

```javascript
ec2_instance_attribute(us - east - 1, Tags.Name, { 'tag:Team': ['sysops'] });
```

## Using JSON format template variables

Some queries accept filters in JSON format and Grafana supports the conversion of template variables to JSON.

If `env = 'production', 'staging'`, following query will return ARNs of EC2 instances which `Environment` tag is `production` or `staging`.

```javascript
resource_arns(us-east-1, ec2:instance, {"Environment":${env:json}})
```

## Pricing

The Amazon CloudWatch data source for Grafana uses the `ListMetrics` and `GetMetricData` CloudWatch API calls to list and retrieve metrics.
Pricing for CloudWatch Logs is based on the amount of data ingested, archived, and analyzed via CloudWatch Logs Insights queries.
Please see the [CloudWatch pricing page](https://aws.amazon.com/cloudwatch/pricing/) for more details.

Every time you pick a dimension in the query editor Grafana will issue a ListMetrics request.
Whenever you make a change to the queries in the query editor, one new request to GetMetricData will be issued.

Please note that for Grafana version 6.5 or higher, all API requests to GetMetricStatistics have been replaced with calls to GetMetricData. This change enables better support for CloudWatch metric math and enables the automatic generation of search expressions when using wildcards or disabling the `Match Exact` option. While GetMetricStatistics qualified for the CloudWatch API free tier, this is not the case for GetMetricData calls. For more information, please refer to the [CloudWatch pricing page](https://aws.amazon.com/cloudwatch/pricing/).

## Service quotas

AWS defines quotas, or limits, for resources, actions, and items in your AWS account. Depending on the number of queries in your dashboard and the number of users accessing the dashboard, you may reach the usage limits for various CloudWatch and CloudWatch Logs resources. Note that quotas are defined per account and per region. If you're using multiple regions or have set up more than one CloudWatch data source to query against multiple accounts, you need to request a quota increase for each account and each region in which you hit the limit.

To request a quota increase, visit the [AWS Service Quotas console](https://console.aws.amazon.com/servicequotas/home?r#!/services/monitoring/quotas/L-5E141212).

Please see the AWS documentation for [Service Quotas](https://docs.aws.amazon.com/servicequotas/latest/userguide/intro.html) and [CloudWatch limits](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html) for more information.

## Configure the data source with grafana.ini

The Grafana [configuration]({{< relref "../../administration/configuration.md#aws" >}}) file includes an `AWS` section where you can customize the data source.

### allowed_auth_providers

Specify which authentication providers are allowed for the CloudWatch data source. The following providers are enabled by default in OSS Grafana: `default` (AWS SDK default), keys (Access and secret key), credentials (Credentials file), ec2_IAM_role (EC2 IAM role).

### assume_role_enabled

Allows you to disable `assume role (ARN)` in the CloudWatch data source. By default, assume role (ARN) is enabled for OSS Grafana.

### list_metrics_page_limit

When a custom namespace is specified in the query editor, the [List Metrics API](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_ListMetrics.html) is used to populate the _Metrics_ field and the _Dimension_ fields. The API is paginated and returns up to 500 results per page. The CloudWatch data source also limits the number of pages to 500. However, you can change this limit using the `list_metrics_page_limit` variable in the [grafana configuration file](https://grafana.com/docs/grafana/latest/administration/configuration/#aws).

## Configure the data source with provisioning

You can configure the CloudWatch data source by customizing configuration files in Grafana's provisioning system. To know more about provisioning and learn about available configuration options, refer to the [Provisioning Grafana]({{< relref "../../administration/provisioning/#datasources" >}}) topic.

Here are some provisioning examples for this data source.

### Using AWS SDK (default)

```yaml
apiVersion: 1
datasources:
  - name: CloudWatch
    type: cloudwatch
    jsonData:
      authType: default
      defaultRegion: eu-west-2
```

### Using credentials' profile name (non-default)

```yaml
apiVersion: 1

datasources:
  - name: CloudWatch
    type: cloudwatch
    jsonData:
      authType: credentials
      defaultRegion: eu-west-2
      customMetricsNamespaces: 'CWAgent,CustomNameSpace'
      profile: secondary
```

### Using `accessKey` and `secretKey`

```yaml
apiVersion: 1

datasources:
  - name: CloudWatch
    type: cloudwatch
    jsonData:
      authType: keys
      defaultRegion: eu-west-2
    secureJsonData:
      accessKey: '<your access key>'
      secretKey: '<your secret key>'
```

### Using AWS SDK Default and ARN of IAM Role to Assume

```yaml
apiVersion: 1
datasources:
  - name: CloudWatch
    type: cloudwatch
    jsonData:
      authType: default
      assumeRoleArn: arn:aws:iam::123456789012:root
      defaultRegion: eu-west-2
```
