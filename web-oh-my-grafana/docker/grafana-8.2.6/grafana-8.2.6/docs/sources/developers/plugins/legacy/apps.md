+++
title = "Legacy app plugins"
keywords = ["grafana", "plugins", "documentation"]
aliases = ["/docs/grafana/latest/plugins/developing/apps/"]
+++

# Legacy app plugins

App plugins are Grafana plugins that can bundle data source and panel plugins within one package. They also enable the plugin author to create custom pages within Grafana. The custom pages enable the plugin author to include things like documentation, sign-up forms, or to control other services with HTTP requests.

Data source and panel plugins will show up like normal plugins. The app pages will be available in the main menu.

{{< figure class="float-right"  src="/static/img/docs/v3/app-in-main-menu.png" caption="App in Main Menu" >}}

## Enabling app plugins

After installing an app, it has to be enabled before it shows up as a data source or panel. You can do that on the app page in the configuration tab.

## Developing an App Plugin

An App is a bundle of panels, dashboards and/or data source(s). There is nothing different about developing panels and data sources for an app.

Apps have to be enabled in Grafana and should import any included dashboards when the user enables it. A ConfigCtrl class should be created and the dashboards imported in the postUpdate hook. See example below:

```javascript
export class ConfigCtrl {
  /** @ngInject */
  constructor($scope, $injector, $q) {
    this.$q = $q;
    this.enabled = false;
    this.appEditCtrl.setPostUpdateHook(this.postUpdate.bind(this));
  }

  postUpdate() {
    if (!this.appModel.enabled) {
      return;
    }

    // TODO, whatever you want
    console.log('Post Update:', this);
  }
}
ConfigCtrl.templateUrl = 'components/config/config.html';
```

If possible, a link to a dashboard or custom page should be shown after enabling the app to guide the user to the appropriate place.

{{< figure class="float-right"  src="/static/img/docs/app_plugin_after_enable.png" caption="After enabling" >}}

### Develop your own App

> Our goal is not to have a very extensive documentation but rather have actual
> code that people can look at. An example implementation of an app can be found
> in this [example app repo](https://github.com/grafana/simple-app-plugin)
