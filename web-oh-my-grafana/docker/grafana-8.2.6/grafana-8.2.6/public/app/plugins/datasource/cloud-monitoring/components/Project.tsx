import React, { useEffect, useState } from 'react';
import { SelectableValue } from '@grafana/data';
import { Select } from '@grafana/ui';
import CloudMonitoringDatasource from '../datasource';
import { SELECT_WIDTH } from '../constants';
import { QueryEditorRow } from '.';

export interface Props {
  datasource: CloudMonitoringDatasource;
  onChange: (projectName: string) => void;
  templateVariableOptions: Array<SelectableValue<string>>;
  projectName: string;
}

export function Project({ projectName, datasource, onChange, templateVariableOptions }: Props) {
  const [projects, setProjects] = useState<Array<SelectableValue<string>>>([]);
  useEffect(() => {
    datasource.getProjects().then((projects) =>
      setProjects([
        {
          label: 'Template Variables',
          options: templateVariableOptions,
        },
        ...projects,
      ])
    );
  }, [datasource, templateVariableOptions]);

  return (
    <QueryEditorRow label="Project">
      <Select
        menuShouldPortal
        width={SELECT_WIDTH}
        allowCustomValue
        formatCreateLabel={(v) => `Use project: ${v}`}
        onChange={({ value }) => onChange(value!)}
        options={projects}
        value={{ value: projectName, label: projectName }}
        placeholder="Select Project"
      />
    </QueryEditorRow>
  );
}
