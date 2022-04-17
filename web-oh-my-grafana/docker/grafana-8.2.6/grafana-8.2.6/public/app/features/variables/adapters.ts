import { ComponentType } from 'react';
import { Reducer } from 'redux';
import { Registry, UrlQueryValue, VariableType } from '@grafana/data';

import { VariableModel, VariableOption } from './types';
import { VariableEditorProps } from './editor/types';
import { VariablePickerProps } from './pickers/types';
import { createQueryVariableAdapter } from './query/adapter';
import { createCustomVariableAdapter } from './custom/adapter';
import { createTextBoxVariableAdapter } from './textbox/adapter';
import { createConstantVariableAdapter } from './constant/adapter';
import { createDataSourceVariableAdapter } from './datasource/adapter';
import { createIntervalVariableAdapter } from './interval/adapter';
import { createAdHocVariableAdapter } from './adhoc/adapter';
import { createSystemVariableAdapter } from './system/adapter';
import { VariablesState } from './state/types';

export interface VariableAdapter<Model extends VariableModel> {
  id: VariableType;
  description: string;
  name: string;
  initialState: Model;
  dependsOn: (variable: Model, variableToTest: Model) => boolean;
  setValue: (variable: Model, option: VariableOption, emitChanges?: boolean) => Promise<void>;
  setValueFromUrl: (variable: Model, urlValue: UrlQueryValue) => Promise<void>;
  updateOptions: (variable: Model, searchFilter?: string) => Promise<void>;
  getSaveModel: (variable: Model, saveCurrentAsDefault?: boolean) => Partial<Model>;
  getValueForUrl: (variable: Model) => string | string[];
  picker: ComponentType<VariablePickerProps<Model>>;
  editor: ComponentType<VariableEditorProps<Model>>;
  reducer: Reducer<VariablesState>;
  beforeAdding?: (model: any) => any;
}

export const getDefaultVariableAdapters = () => [
  createQueryVariableAdapter(),
  createCustomVariableAdapter(),
  createTextBoxVariableAdapter(),
  createConstantVariableAdapter(),
  createDataSourceVariableAdapter(),
  createIntervalVariableAdapter(),
  createAdHocVariableAdapter(),
  createSystemVariableAdapter(),
];

export const variableAdapters = new Registry<VariableAdapter<any>>();
