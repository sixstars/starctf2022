import React, { useEffect, useState, useMemo } from 'react';
import { Alert, Button, ConfirmModal, TextArea, HorizontalGroup, Field, Form } from '@grafana/ui';
import { useAlertManagerSourceName } from './hooks/useAlertManagerSourceName';
import { AlertingPageWrapper } from './components/AlertingPageWrapper';
import { AlertManagerPicker } from './components/AlertManagerPicker';
import { GRAFANA_RULES_SOURCE_NAME, isVanillaPrometheusAlertManagerDataSource } from './utils/datasource';
import { useDispatch } from 'react-redux';
import {
  deleteAlertManagerConfigAction,
  fetchAlertManagerConfigAction,
  updateAlertManagerConfigAction,
} from './state/actions';
import { useUnifiedAlertingSelector } from './hooks/useUnifiedAlertingSelector';
import { initialAsyncRequestState } from './utils/redux';

interface FormValues {
  configJSON: string;
}

export default function Admin(): JSX.Element {
  const dispatch = useDispatch();
  const [alertManagerSourceName, setAlertManagerSourceName] = useAlertManagerSourceName();
  const [showConfirmDeleteAMConfig, setShowConfirmDeleteAMConfig] = useState(false);
  const { loading: isDeleting } = useUnifiedAlertingSelector((state) => state.deleteAMConfig);
  const { loading: isSaving } = useUnifiedAlertingSelector((state) => state.saveAMConfig);
  const readOnly = alertManagerSourceName ? isVanillaPrometheusAlertManagerDataSource(alertManagerSourceName) : false;

  const configRequests = useUnifiedAlertingSelector((state) => state.amConfigs);

  const { result: config, loading: isLoadingConfig, error: loadingError } =
    (alertManagerSourceName && configRequests[alertManagerSourceName]) || initialAsyncRequestState;

  useEffect(() => {
    if (alertManagerSourceName) {
      dispatch(fetchAlertManagerConfigAction(alertManagerSourceName));
    }
  }, [alertManagerSourceName, dispatch]);

  const resetConfig = () => {
    if (alertManagerSourceName) {
      dispatch(deleteAlertManagerConfigAction(alertManagerSourceName));
    }
    setShowConfirmDeleteAMConfig(false);
  };

  const defaultValues = useMemo(
    (): FormValues => ({
      configJSON: config ? JSON.stringify(config, null, 2) : '',
    }),
    [config]
  );

  const loading = isDeleting || isLoadingConfig || isSaving;

  const onSubmit = (values: FormValues) => {
    if (alertManagerSourceName) {
      dispatch(
        updateAlertManagerConfigAction({
          newConfig: JSON.parse(values.configJSON),
          oldConfig: config,
          alertManagerSourceName,
          successMessage: 'Alertmanager configuration updated.',
          refetch: true,
        })
      );
    }
  };

  return (
    <AlertingPageWrapper pageId="alerting-admin">
      <AlertManagerPicker current={alertManagerSourceName} onChange={setAlertManagerSourceName} />
      {loadingError && !loading && (
        <Alert severity="error" title="Error loading Alertmanager configuration">
          {loadingError.message || 'Unknown error.'}
        </Alert>
      )}
      {isDeleting && alertManagerSourceName !== GRAFANA_RULES_SOURCE_NAME && (
        <Alert severity="info" title="Resetting Alertmanager configuration">
          It might take a while...
        </Alert>
      )}
      {alertManagerSourceName && config && (
        <Form defaultValues={defaultValues} onSubmit={onSubmit} key={defaultValues.configJSON}>
          {({ register, errors }) => (
            <>
              {!readOnly && (
                <Field
                  disabled={loading}
                  label="Configuration"
                  invalid={!!errors.configJSON}
                  error={errors.configJSON?.message}
                >
                  <TextArea
                    {...register('configJSON', {
                      required: { value: true, message: 'Required.' },
                      validate: (v) => {
                        try {
                          JSON.parse(v);
                          return true;
                        } catch (e) {
                          return e.message;
                        }
                      },
                    })}
                    id="configuration"
                    rows={25}
                  />
                </Field>
              )}
              {readOnly && (
                <Field label="Configuration">
                  <pre data-testid="readonly-config">{defaultValues.configJSON}</pre>
                </Field>
              )}
              {!readOnly && (
                <HorizontalGroup>
                  <Button type="submit" variant="primary" disabled={loading}>
                    Save
                  </Button>
                  <Button
                    type="button"
                    disabled={loading}
                    variant="destructive"
                    onClick={() => setShowConfirmDeleteAMConfig(true)}
                  >
                    Reset configuration
                  </Button>
                </HorizontalGroup>
              )}
              {!!showConfirmDeleteAMConfig && (
                <ConfirmModal
                  isOpen={true}
                  title="Reset Alertmanager configuration"
                  body={`Are you sure you want to reset configuration ${
                    alertManagerSourceName === GRAFANA_RULES_SOURCE_NAME
                      ? 'for the Grafana Alertmanager'
                      : `for "${alertManagerSourceName}"`
                  }? Contact points and notification policies will be reset to their defaults.`}
                  confirmText="Yes, reset configuration"
                  onConfirm={resetConfig}
                  onDismiss={() => setShowConfirmDeleteAMConfig(false)}
                />
              )}
            </>
          )}
        </Form>
      )}
    </AlertingPageWrapper>
  );
}
