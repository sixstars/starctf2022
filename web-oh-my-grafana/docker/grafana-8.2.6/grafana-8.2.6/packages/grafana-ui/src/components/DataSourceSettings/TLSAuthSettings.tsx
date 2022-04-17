import React from 'react';
import { KeyValue } from '@grafana/data';
import { css, cx } from '@emotion/css';
import { Tooltip } from '../Tooltip/Tooltip';
import { Icon } from '../Icon/Icon';
import { CertificationKey } from './CertificationKey';
import { HttpSettingsBaseProps } from './types';
import { FormField } from '../FormField/FormField';

export const TLSAuthSettings: React.FC<HttpSettingsBaseProps> = ({ dataSourceConfig, onChange }) => {
  const hasTLSCACert = dataSourceConfig.secureJsonFields && dataSourceConfig.secureJsonFields.tlsCACert;
  const hasTLSClientCert = dataSourceConfig.secureJsonFields && dataSourceConfig.secureJsonFields.tlsClientCert;
  const hasTLSClientKey = dataSourceConfig.secureJsonFields && dataSourceConfig.secureJsonFields.tlsClientKey;
  const hasServerName = dataSourceConfig.jsonData && dataSourceConfig.jsonData.serverName;

  const onResetClickFactory = (field: string) => (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    const newSecureJsonFields: KeyValue<boolean> = { ...dataSourceConfig.secureJsonFields };
    newSecureJsonFields[field] = false;
    onChange({
      ...dataSourceConfig,
      secureJsonFields: newSecureJsonFields,
    });
  };

  const onCertificateChangeFactory = (field: string) => (event: React.SyntheticEvent<HTMLTextAreaElement>) => {
    const newSecureJsonData = { ...dataSourceConfig.secureJsonData };
    newSecureJsonData[field] = event.currentTarget.value;

    onChange({
      ...dataSourceConfig,
      secureJsonData: newSecureJsonData,
    });
  };

  const onServerNameLabelChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const newJsonData = {
      ...dataSourceConfig.jsonData,
      serverName: event.currentTarget.value,
    };

    onChange({
      ...dataSourceConfig,
      jsonData: newJsonData,
    });
  };

  return (
    <div className="gf-form-group">
      <div
        className={cx(
          'gf-form',
          css`
            align-items: baseline;
          `
        )}
      >
        <h6>TLS/SSL Auth Details</h6>
        <Tooltip
          placement="right-end"
          content="TLS/SSL Certs are encrypted and stored in the Grafana database."
          theme="info"
        >
          <div className="gf-form-help-icon gf-form-help-icon--right-normal">
            <Icon name="info-circle" size="xs" style={{ marginLeft: '10px' }} />
          </div>
        </Tooltip>
      </div>
      <div>
        {dataSourceConfig.jsonData.tlsAuthWithCACert && (
          <CertificationKey
            hasCert={!!hasTLSCACert}
            onChange={onCertificateChangeFactory('tlsCACert')}
            placeholder="Begins with -----BEGIN CERTIFICATE-----"
            label="CA Cert"
            onClick={onResetClickFactory('tlsCACert')}
          />
        )}

        {dataSourceConfig.jsonData.tlsAuth && (
          <>
            <div className="gf-form">
              <FormField
                label="ServerName"
                labelWidth={7}
                inputWidth={30}
                placeholder="domain.example.com"
                value={hasServerName && dataSourceConfig.jsonData.serverName}
                onChange={onServerNameLabelChange}
              />
            </div>
            <CertificationKey
              hasCert={!!hasTLSClientCert}
              label="Client Cert"
              onChange={onCertificateChangeFactory('tlsClientCert')}
              placeholder="Begins with -----BEGIN CERTIFICATE-----"
              onClick={onResetClickFactory('tlsClientCert')}
            />

            <CertificationKey
              hasCert={!!hasTLSClientKey}
              label="Client Key"
              placeholder="Begins with -----BEGIN RSA PRIVATE KEY-----"
              onChange={onCertificateChangeFactory('tlsClientKey')}
              onClick={onResetClickFactory('tlsClientKey')}
            />
          </>
        )}
      </div>
    </div>
  );
};
