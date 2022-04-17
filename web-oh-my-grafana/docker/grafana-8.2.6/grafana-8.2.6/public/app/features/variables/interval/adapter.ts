import { cloneDeep } from 'lodash';
import { IntervalVariableModel } from '../types';
import { dispatch } from '../../../store/store';
import { setOptionAsCurrent, setOptionFromUrl } from '../state/actions';
import { VariableAdapter } from '../adapters';
import { initialIntervalVariableModelState, intervalVariableReducer } from './reducer';
import { toVariableIdentifier } from '../state/types';
import { IntervalVariableEditor } from './IntervalVariableEditor';
import { updateAutoValue, updateIntervalVariableOptions } from './actions';
import { optionPickerFactory } from '../pickers';

export const createIntervalVariableAdapter = (): VariableAdapter<IntervalVariableModel> => {
  return {
    id: 'interval',
    description: 'Define a timespan interval (ex 1m, 1h, 1d)',
    name: 'Interval',
    initialState: initialIntervalVariableModelState,
    reducer: intervalVariableReducer,
    picker: optionPickerFactory<IntervalVariableModel>(),
    editor: IntervalVariableEditor,
    dependsOn: () => {
      return false;
    },
    setValue: async (variable, option, emitChanges = false) => {
      await dispatch(updateAutoValue(toVariableIdentifier(variable)));
      await dispatch(setOptionAsCurrent(toVariableIdentifier(variable), option, emitChanges));
    },
    setValueFromUrl: async (variable, urlValue) => {
      await dispatch(updateAutoValue(toVariableIdentifier(variable)));
      await dispatch(setOptionFromUrl(toVariableIdentifier(variable), urlValue));
    },
    updateOptions: async (variable) => {
      await dispatch(updateIntervalVariableOptions(toVariableIdentifier(variable)));
    },
    getSaveModel: (variable) => {
      const { index, id, state, global, ...rest } = cloneDeep(variable);
      return rest;
    },
    getValueForUrl: (variable) => {
      return variable.current.value;
    },
  };
};
