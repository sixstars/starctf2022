import { act, render, screen } from '@testing-library/react';
import React, { Component } from 'react';
import AppRootPage from './AppRootPage';
import { getPluginSettings } from './PluginSettingsCache';
import { importAppPlugin } from './plugin_loader';
import { getMockPlugin } from './__mocks__/pluginMocks';
import { AppPlugin, PluginType, AppRootProps, NavModelItem } from '@grafana/data';
import { Route, Router } from 'react-router-dom';
import { locationService, setEchoSrv } from '@grafana/runtime';
import { GrafanaRoute } from 'app/core/navigation/GrafanaRoute';
import { Echo } from 'app/core/services/echo/Echo';

jest.mock('./PluginSettingsCache', () => ({
  getPluginSettings: jest.fn(),
}));
jest.mock('./plugin_loader', () => ({
  importAppPlugin: jest.fn(),
}));

const importAppPluginMock = importAppPlugin as jest.Mock<
  ReturnType<typeof importAppPlugin>,
  Parameters<typeof importAppPlugin>
>;

const getPluginSettingsMock = getPluginSettings as jest.Mock<
  ReturnType<typeof getPluginSettings>,
  Parameters<typeof getPluginSettings>
>;

class RootComponent extends Component<AppRootProps> {
  static timesMounted = 0;
  componentDidMount() {
    RootComponent.timesMounted += 1;
    const node: NavModelItem = {
      text: 'My Great plugin',
      children: [
        {
          text: 'A page',
          url: '/apage',
          id: 'a',
        },
        {
          text: 'Another page',
          url: '/anotherpage',
          id: 'b',
        },
      ],
    };
    this.props.onNavChanged({
      main: node,
      node,
    });
  }

  render() {
    return <p>my great plugin</p>;
  }
}

function renderUnderRouter() {
  const route = { component: AppRootPage };
  locationService.push('/a/my-awesome-plugin');

  render(
    <Router history={locationService.getHistory()}>
      <Route path="/a/:pluginId" exact render={(props) => <GrafanaRoute {...props} route={route as any} />} />
    </Router>
  );
}

describe('AppRootPage', () => {
  beforeEach(() => {
    jest.resetAllMocks();
    setEchoSrv(new Echo());
  });

  it('should not mount plugin twice if nav is changed', async () => {
    // reproduces https://github.com/grafana/grafana/pull/28105

    getPluginSettingsMock.mockResolvedValue(
      getMockPlugin({
        type: PluginType.app,
        enabled: true,
      })
    );

    const plugin = new AppPlugin();
    plugin.root = RootComponent;

    importAppPluginMock.mockResolvedValue(plugin);

    renderUnderRouter();

    // check that plugin and nav links were rendered, and plugin is mounted only once
    expect(await screen.findByText('my great plugin')).toBeVisible();
    expect(await screen.findByLabelText('Tab A page')).toBeVisible();
    expect(await screen.findByLabelText('Tab Another page')).toBeVisible();
    expect(RootComponent.timesMounted).toEqual(1);
  });

  it('should not render component if not at plugin path', async () => {
    getPluginSettingsMock.mockResolvedValue(
      getMockPlugin({
        type: PluginType.app,
        enabled: true,
      })
    );

    class RootComponent extends Component<AppRootProps> {
      static timesRendered = 0;
      render() {
        RootComponent.timesRendered += 1;
        return <p>my great component</p>;
      }
    }

    const plugin = new AppPlugin();
    plugin.root = RootComponent;

    importAppPluginMock.mockResolvedValue(plugin);

    renderUnderRouter();

    expect(await screen.findByText('my great component')).toBeVisible();

    // renders the first time
    expect(RootComponent.timesRendered).toEqual(1);

    await act(async () => {
      locationService.push('/foo');
    });

    expect(RootComponent.timesRendered).toEqual(1);

    await act(async () => {
      locationService.push('/a/my-awesome-plugin');
    });

    expect(RootComponent.timesRendered).toEqual(2);
  });
});
