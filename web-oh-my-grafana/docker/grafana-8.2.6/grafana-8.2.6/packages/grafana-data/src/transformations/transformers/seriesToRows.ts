import { omit } from 'lodash';
import { map } from 'rxjs/operators';

import { DataTransformerID } from './ids';
import { DataTransformerInfo } from '../../types/transformations';
import {
  Field,
  FieldType,
  TIME_SERIES_METRIC_FIELD_NAME,
  TIME_SERIES_TIME_FIELD_NAME,
  TIME_SERIES_VALUE_FIELD_NAME,
} from '../../types/dataFrame';
import { isTimeSeries } from '../../dataframe/utils';
import { MutableDataFrame, sortDataFrame } from '../../dataframe';
import { ArrayVector } from '../../vector';
import { getFrameDisplayName } from '../../field/fieldState';

export interface SeriesToRowsTransformerOptions {}

export const seriesToRowsTransformer: DataTransformerInfo<SeriesToRowsTransformerOptions> = {
  id: DataTransformerID.seriesToRows,
  name: 'Series to rows',
  description: 'Combines multiple series into a single serie and appends a column with metric name per value.',
  defaultOptions: {},
  operator: (options) => (source) =>
    source.pipe(
      map((data) => {
        if (!Array.isArray(data) || data.length <= 1) {
          return data;
        }

        if (!isTimeSeries(data)) {
          return data;
        }

        const timeFieldByIndex: Record<number, number> = {};
        const targetFields = new Set<string>();
        const dataFrame = new MutableDataFrame();
        const metricField: Field = {
          name: TIME_SERIES_METRIC_FIELD_NAME,
          values: new ArrayVector(),
          config: {},
          type: FieldType.string,
        };

        for (let frameIndex = 0; frameIndex < data.length; frameIndex++) {
          const frame = data[frameIndex];

          for (let fieldIndex = 0; fieldIndex < frame.fields.length; fieldIndex++) {
            const field = frame.fields[fieldIndex];

            if (field.type === FieldType.time) {
              timeFieldByIndex[frameIndex] = fieldIndex;

              if (!targetFields.has(TIME_SERIES_TIME_FIELD_NAME)) {
                dataFrame.addField(copyFieldStructure(field, TIME_SERIES_TIME_FIELD_NAME));
                dataFrame.addField(metricField);
                targetFields.add(TIME_SERIES_TIME_FIELD_NAME);
              }
              continue;
            }

            if (!targetFields.has(TIME_SERIES_VALUE_FIELD_NAME)) {
              dataFrame.addField(copyFieldStructure(field, TIME_SERIES_VALUE_FIELD_NAME));
              targetFields.add(TIME_SERIES_VALUE_FIELD_NAME);
            }
          }
        }

        for (let frameIndex = 0; frameIndex < data.length; frameIndex++) {
          const frame = data[frameIndex];

          for (let valueIndex = 0; valueIndex < frame.length; valueIndex++) {
            const timeFieldIndex = timeFieldByIndex[frameIndex];
            const valueFieldIndex = timeFieldIndex === 0 ? 1 : 0;

            dataFrame.add({
              [TIME_SERIES_TIME_FIELD_NAME]: frame.fields[timeFieldIndex].values.get(valueIndex),
              [TIME_SERIES_METRIC_FIELD_NAME]: getFrameDisplayName(frame),
              [TIME_SERIES_VALUE_FIELD_NAME]: frame.fields[valueFieldIndex].values.get(valueIndex),
            });
          }
        }

        return [sortDataFrame(dataFrame, 0, true)];
      })
    ),
};

const copyFieldStructure = (field: Field, name: string): Field => {
  return {
    ...omit(field, ['values', 'state', 'labels', 'config', 'name']),
    name: name,
    values: new ArrayVector(),
    config: {
      ...omit(field.config, ['displayName', 'displayNameFromDS']),
    },
  };
};
