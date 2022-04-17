import React, { ChangeEvent, FocusEvent, KeyboardEvent, ReactElement, useCallback, useEffect, useState } from 'react';

import { TextBoxVariableModel } from '../types';
import { toVariablePayload } from '../state/types';
import { changeVariableProp } from '../state/sharedReducer';
import { VariablePickerProps } from '../pickers/types';
import { Input } from '@grafana/ui';
import { variableAdapters } from '../adapters';
import { useDispatch } from 'react-redux';

export interface Props extends VariablePickerProps<TextBoxVariableModel> {}

export function TextBoxVariablePicker({ variable, onVariableChange }: Props): ReactElement {
  const dispatch = useDispatch();
  const [updatedValue, setUpdatedValue] = useState(variable.current.value);
  useEffect(() => {
    setUpdatedValue(variable.current.value);
  }, [variable]);

  const updateVariable = useCallback(() => {
    if (variable.current.value === updatedValue) {
      return;
    }

    dispatch(
      changeVariableProp(
        toVariablePayload({ id: variable.id, type: variable.type }, { propName: 'query', propValue: updatedValue })
      )
    );

    if (onVariableChange) {
      onVariableChange({
        ...variable,
        current: { ...variable.current, value: updatedValue },
      });
      return;
    }

    variableAdapters.get(variable.type).updateOptions(variable);
  }, [variable, updatedValue, dispatch, onVariableChange]);

  const onChange = useCallback((event: ChangeEvent<HTMLInputElement>) => setUpdatedValue(event.target.value), [
    setUpdatedValue,
  ]);

  const onBlur = (e: FocusEvent<HTMLInputElement>) => updateVariable();
  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.keyCode === 13) {
      event.preventDefault();
      updateVariable();
    }
  };

  return (
    <Input
      type="text"
      value={updatedValue}
      onChange={onChange}
      onBlur={onBlur}
      onKeyDown={onKeyDown}
      placeholder="Enter variable value"
      id={variable.id}
    />
  );
}
