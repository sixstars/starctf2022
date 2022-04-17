import React, { useState, useCallback } from 'react';
import { css, cx } from '@emotion/css';
import { DataSourceSettings, SelectableValue } from '@grafana/data';
import { BasicAuthSettings } from './BasicAuthSettings';
import { HttpProxySettings } from './HttpProxySettings';
import { TLSAuthSettings } from './TLSAuthSettings';
import { CustomHeadersSettings } from './CustomHeadersSettings';
import { Select } from '../Forms/Legacy/Select/Select';
import { Input } from '../Forms/Legacy/Input/Input';
import { Switch } from '../Forms/Legacy/Switch/Switch';
import { Icon } from '../Icon/Icon';
import { FormField } from '../FormField/FormField';
import { InlineFormLabel } from '../FormLabel/FormLabel';
import { TagsInput } from '../TagsInput/TagsInput';
import { SigV4AuthSettings } from './SigV4AuthSettings';
import { useTheme } from '../../themes';
import { HttpSettingsProps } from './types';

const ACCESS_OPTIONS: Array<SelectableValue<string>> = [
  {
    label: 'Server (default)',
    value: 'proxy',
  },
  {
    label: 'Browser',
    value: 'direct',
  },
];

const DEFAULT_ACCESS_OPTION = {
  label: 'Server (default)',
  value: 'proxy',
};

const HttpAccessHelp = () => (
  <div className="grafana-info-box m-t-2">
    <p>
      Access mode controls how requests to the data source will be handled.
      <strong>
        &nbsp;<i>Server</i>
      </strong>{' '}
      should be the preferred way if nothing else is stated.
    </p>
    <div className="alert-title">Server access mode (Default):</div>
    <p>
      All requests will be made from the browser to Grafana backend/server which in turn will forward the requests to
      the data source and by that circumvent possible Cross-Origin Resource Sharing (CORS) requirements. The URL needs
      to be accessible from the grafana backend/server if you select this access mode.
    </p>
    <div className="alert-title">Browser access mode:</div>
    <p>
      All requests will be made from the browser directly to the data source and may be subject to Cross-Origin Resource
      Sharing (CORS) requirements. The URL needs to be accessible from the browser if you select this access mode.
    </p>
  </div>
);

export const DataSourceHttpSettings: React.FC<HttpSettingsProps> = (props) => {
  const {
    defaultUrl,
    dataSourceConfig,
    onChange,
    showAccessOptions,
    sigV4AuthToggleEnabled,
    showForwardOAuthIdentityOption,
    azureAuthSettings,
  } = props;
  let urlTooltip;
  const [isAccessHelpVisible, setIsAccessHelpVisible] = useState(false);
  const theme = useTheme();

  const onSettingsChange = useCallback(
    (change: Partial<DataSourceSettings<any, any>>) => {
      onChange({
        ...dataSourceConfig,
        ...change,
      });
    },
    [dataSourceConfig, onChange]
  );

  switch (dataSourceConfig.access) {
    case 'direct':
      urlTooltip = (
        <>
          Your access method is <em>Browser</em>, this means the URL needs to be accessible from the browser.
        </>
      );
      break;
    case 'proxy':
      urlTooltip = (
        <>
          Your access method is <em>Server</em>, this means the URL needs to be accessible from the grafana
          backend/server.
        </>
      );
      break;
    default:
      urlTooltip = 'Specify a complete HTTP URL (for example http://your_server:8080)';
  }

  const accessSelect = (
    <Select
      menuShouldPortal
      width={20}
      options={ACCESS_OPTIONS}
      value={ACCESS_OPTIONS.filter((o) => o.value === dataSourceConfig.access)[0] || DEFAULT_ACCESS_OPTION}
      onChange={(selectedValue) => onSettingsChange({ access: selectedValue.value })}
    />
  );

  const isValidUrl = /^(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?$/.test(
    dataSourceConfig.url
  );

  const notValidStyle = css`
    box-shadow: inset 0 0px 5px ${theme.palette.red};
  `;

  const inputStyle = cx({ [`width-20`]: true, [notValidStyle]: !isValidUrl });

  const urlInput = (
    <Input
      className={inputStyle}
      placeholder={defaultUrl}
      value={dataSourceConfig.url}
      onChange={(event) => onSettingsChange({ url: event.currentTarget.value })}
    />
  );

  return (
    <div className="gf-form-group">
      <>
        <h3 className="page-heading">HTTP</h3>
        <div className="gf-form-group">
          <div className="gf-form">
            <FormField label="URL" labelWidth={13} tooltip={urlTooltip} inputEl={urlInput} />
          </div>

          {showAccessOptions && (
            <>
              <div className="gf-form-inline">
                <div className="gf-form">
                  <FormField label="Access" labelWidth={13} inputWidth={20} inputEl={accessSelect} />
                </div>
                <div className="gf-form">
                  <label
                    className="gf-form-label query-keyword pointer"
                    onClick={() => setIsAccessHelpVisible((isVisible) => !isVisible)}
                  >
                    Help&nbsp;
                    <Icon name={isAccessHelpVisible ? 'angle-down' : 'angle-right'} style={{ marginBottom: 0 }} />
                  </label>
                </div>
              </div>
              {isAccessHelpVisible && <HttpAccessHelp />}
            </>
          )}
          {dataSourceConfig.access === 'proxy' && (
            <div className="gf-form-group">
              <div className="gf-form">
                <InlineFormLabel
                  width={13}
                  tooltip="Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source."
                >
                  Allowed cookies
                </InlineFormLabel>
                <TagsInput
                  tags={dataSourceConfig.jsonData.keepCookies}
                  width={40}
                  onChange={(cookies) =>
                    onSettingsChange({ jsonData: { ...dataSourceConfig.jsonData, keepCookies: cookies } })
                  }
                />
              </div>
              <div className="gf-form">
                <FormField
                  label="Timeout"
                  type="number"
                  labelWidth={13}
                  inputWidth={20}
                  tooltip="HTTP request timeout in seconds"
                  value={dataSourceConfig.jsonData.timeout}
                  onChange={(event) => {
                    onSettingsChange({
                      jsonData: { ...dataSourceConfig.jsonData, timeout: parseInt(event.currentTarget.value, 10) },
                    });
                  }}
                />
              </div>
            </div>
          )}
        </div>
      </>

      <>
        <h3 className="page-heading">Auth</h3>
        <div className="gf-form-group">
          <div className="gf-form-inline">
            <Switch
              label="Basic auth"
              labelClass="width-13"
              checked={dataSourceConfig.basicAuth}
              onChange={(event) => {
                onSettingsChange({ basicAuth: event!.currentTarget.checked });
              }}
            />
            <Switch
              label="With Credentials"
              labelClass="width-13"
              checked={dataSourceConfig.withCredentials}
              onChange={(event) => {
                onSettingsChange({ withCredentials: event!.currentTarget.checked });
              }}
              tooltip="Whether credentials such as cookies or auth headers should be sent with cross-site requests."
            />
          </div>

          {azureAuthSettings?.azureAuthEnabled && (
            <div className="gf-form-inline">
              <Switch
                label="Azure Authentication"
                labelClass="width-13"
                checked={dataSourceConfig.jsonData.azureAuth || false}
                onChange={(event) => {
                  onSettingsChange({
                    jsonData: { ...dataSourceConfig.jsonData, azureAuth: event!.currentTarget.checked },
                  });
                }}
                tooltip="Use Azure authentication for Azure endpoint."
              />
            </div>
          )}

          {sigV4AuthToggleEnabled && (
            <div className="gf-form-inline">
              <Switch
                label="SigV4 auth"
                labelClass="width-13"
                checked={dataSourceConfig.jsonData.sigV4Auth || false}
                onChange={(event) => {
                  onSettingsChange({
                    jsonData: { ...dataSourceConfig.jsonData, sigV4Auth: event!.currentTarget.checked },
                  });
                }}
              />
            </div>
          )}

          {dataSourceConfig.access === 'proxy' && (
            <HttpProxySettings
              dataSourceConfig={dataSourceConfig}
              onChange={(jsonData) => onSettingsChange({ jsonData })}
              showForwardOAuthIdentityOption={showForwardOAuthIdentityOption}
            />
          )}
        </div>
        {dataSourceConfig.basicAuth && (
          <>
            <h6>Basic Auth Details</h6>
            <div className="gf-form-group">
              <BasicAuthSettings {...props} />
            </div>
          </>
        )}

        {azureAuthSettings?.azureAuthEnabled &&
          azureAuthSettings?.azureSettingsUI &&
          dataSourceConfig.jsonData.azureAuth && (
            <azureAuthSettings.azureSettingsUI dataSourceConfig={dataSourceConfig} onChange={onChange} />
          )}

        {dataSourceConfig.jsonData.sigV4Auth && sigV4AuthToggleEnabled && <SigV4AuthSettings {...props} />}

        {(dataSourceConfig.jsonData.tlsAuth || dataSourceConfig.jsonData.tlsAuthWithCACert) && (
          <TLSAuthSettings dataSourceConfig={dataSourceConfig} onChange={onChange} />
        )}

        {dataSourceConfig.access === 'proxy' && (
          <CustomHeadersSettings dataSourceConfig={dataSourceConfig} onChange={onChange} />
        )}
      </>
    </div>
  );
};
