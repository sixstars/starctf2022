import React from 'react';
import { DataSourceSettings } from '@grafana/data';

export interface AzureAuthSettings {
  azureAuthEnabled: boolean;
  azureSettingsUI?: React.ComponentType<HttpSettingsBaseProps>;
}

export interface HttpSettingsBaseProps {
  /** The configuration object of the data source */
  dataSourceConfig: DataSourceSettings<any, any>;
  /** Callback for handling changes to the configuration object */
  onChange: (config: DataSourceSettings) => void;
  /** Show the Forward OAuth identity option */
  showForwardOAuthIdentityOption?: boolean;
}

export interface HttpSettingsProps extends HttpSettingsBaseProps {
  /** The default url for the data source */
  defaultUrl: string;
  /** Show the http access help box */
  showAccessOptions?: boolean;
  /** Show the SigV4 auth toggle option */
  sigV4AuthToggleEnabled?: boolean;
  /** Azure authentication settings **/
  azureAuthSettings?: AzureAuthSettings;
}
