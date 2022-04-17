import React, { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { Alert, LoadingPlaceholder, useStyles2, withErrorBoundary } from '@grafana/ui';
import { useDispatch } from 'react-redux';
import { Redirect } from 'react-router-dom';
import { Receiver } from 'app/plugins/datasource/alertmanager/types';
import { useCleanup } from '../../../core/hooks/useCleanup';
import { AlertingPageWrapper } from './components/AlertingPageWrapper';
import { AlertManagerPicker } from './components/AlertManagerPicker';
import { AmRootRoute } from './components/amroutes/AmRootRoute';
import { AmSpecificRouting } from './components/amroutes/AmSpecificRouting';
import { useAlertManagerSourceName } from './hooks/useAlertManagerSourceName';
import { useUnifiedAlertingSelector } from './hooks/useUnifiedAlertingSelector';
import { fetchAlertManagerConfigAction, updateAlertManagerConfigAction } from './state/actions';
import { AmRouteReceiver, FormAmRoute } from './types/amroutes';
import { amRouteToFormAmRoute, formAmRouteToAmRoute, stringsToSelectableValues } from './utils/amroutes';
import { initialAsyncRequestState } from './utils/redux';
import { isVanillaPrometheusAlertManagerDataSource } from './utils/datasource';

const AmRoutes: FC = () => {
  const dispatch = useDispatch();
  const styles = useStyles2(getStyles);
  const [isRootRouteEditMode, setIsRootRouteEditMode] = useState(false);
  const [alertManagerSourceName, setAlertManagerSourceName] = useAlertManagerSourceName();

  const readOnly = alertManagerSourceName ? isVanillaPrometheusAlertManagerDataSource(alertManagerSourceName) : true;

  const amConfigs = useUnifiedAlertingSelector((state) => state.amConfigs);

  const fetchConfig = useCallback(() => {
    if (alertManagerSourceName) {
      dispatch(fetchAlertManagerConfigAction(alertManagerSourceName));
    }
  }, [alertManagerSourceName, dispatch]);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  const { result, loading: resultLoading, error: resultError } =
    (alertManagerSourceName && amConfigs[alertManagerSourceName]) || initialAsyncRequestState;

  const config = result?.alertmanager_config;
  const [rootRoute, id2ExistingRoute] = useMemo(() => amRouteToFormAmRoute(config?.route), [config?.route]);

  const receivers = stringsToSelectableValues(
    (config?.receivers ?? []).map((receiver: Receiver) => receiver.name)
  ) as AmRouteReceiver[];

  const enterRootRouteEditMode = () => {
    setIsRootRouteEditMode(true);
  };

  const exitRootRouteEditMode = () => {
    setIsRootRouteEditMode(false);
  };

  useCleanup((state) => state.unifiedAlerting.saveAMConfig);
  const handleSave = (data: Partial<FormAmRoute>) => {
    const newData = formAmRouteToAmRoute(
      alertManagerSourceName,
      {
        ...rootRoute,
        ...data,
      },
      id2ExistingRoute
    );

    if (isRootRouteEditMode) {
      exitRootRouteEditMode();
    }

    dispatch(
      updateAlertManagerConfigAction({
        newConfig: {
          ...result,
          alertmanager_config: {
            ...result.alertmanager_config,
            route: newData,
          },
        },
        oldConfig: result,
        alertManagerSourceName: alertManagerSourceName!,
        successMessage: 'Saved',
        refetch: true,
      })
    );
  };

  if (!alertManagerSourceName) {
    return <Redirect to="/alerting/routes" />;
  }

  return (
    <AlertingPageWrapper pageId="am-routes">
      <AlertManagerPicker current={alertManagerSourceName} onChange={setAlertManagerSourceName} />
      {resultError && !resultLoading && (
        <Alert severity="error" title="Error loading Alertmanager config">
          {resultError.message || 'Unknown error.'}
        </Alert>
      )}
      {resultLoading && <LoadingPlaceholder text="Loading Alertmanager config..." />}
      {result && !resultLoading && !resultError && (
        <>
          <AmRootRoute
            alertManagerSourceName={alertManagerSourceName}
            isEditMode={isRootRouteEditMode}
            onSave={handleSave}
            onEnterEditMode={enterRootRouteEditMode}
            onExitEditMode={exitRootRouteEditMode}
            receivers={receivers}
            routes={rootRoute}
          />
          <div className={styles.break} />
          <AmSpecificRouting
            onChange={handleSave}
            readOnly={readOnly}
            onRootRouteEdit={enterRootRouteEditMode}
            receivers={receivers}
            routes={rootRoute}
          />
        </>
      )}
    </AlertingPageWrapper>
  );
};

export default withErrorBoundary(AmRoutes, { style: 'page' });

const getStyles = (theme: GrafanaTheme2) => ({
  break: css`
    width: 100%;
    height: 0;
    margin-bottom: ${theme.spacing(2)};
    border-bottom: solid 1px ${theme.colors.border.medium};
  `,
});
