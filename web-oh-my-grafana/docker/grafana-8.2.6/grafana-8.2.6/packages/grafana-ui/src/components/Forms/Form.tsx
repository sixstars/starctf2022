import React, { HTMLProps, useEffect } from 'react';
import { useForm, Mode, DeepPartial, UnpackNestedValue, SubmitHandler } from 'react-hook-form';
import { FormAPI } from '../../types';
import { css } from '@emotion/css';

interface FormProps<T> extends Omit<HTMLProps<HTMLFormElement>, 'onSubmit'> {
  validateOn?: Mode;
  validateOnMount?: boolean;
  validateFieldsOnMount?: string | string[];
  defaultValues?: UnpackNestedValue<DeepPartial<T>>;
  onSubmit: SubmitHandler<T>;
  children: (api: FormAPI<T>) => React.ReactNode;
  /** Sets max-width for container. Use it instead of setting individual widths on inputs.*/
  maxWidth?: number | 'none';
}

export function Form<T>({
  defaultValues,
  onSubmit,
  validateOnMount = false,
  validateFieldsOnMount,
  children,
  validateOn = 'onSubmit',
  maxWidth = 600,
  ...htmlProps
}: FormProps<T>) {
  const { handleSubmit, trigger, formState, ...rest } = useForm<T>({
    mode: validateOn,
    defaultValues,
  });

  useEffect(() => {
    if (validateOnMount) {
      //@ts-expect-error
      trigger(validateFieldsOnMount);
    }
  }, [trigger, validateFieldsOnMount, validateOnMount]);

  return (
    <form
      className={css`
        max-width: ${maxWidth !== 'none' ? maxWidth + 'px' : maxWidth};
        width: 100%;
      `}
      onSubmit={handleSubmit(onSubmit)}
      {...htmlProps}
    >
      {children({ errors: formState.errors, formState, ...rest })}
    </form>
  );
}
