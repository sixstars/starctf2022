import { SynchronousDataTransformerInfo } from '../../types';
import { map } from 'rxjs/operators';

import { DataTransformerID } from './ids';
import { DataFrame, Field, FieldConfig, FieldType } from '../../types/dataFrame';
import { ArrayVector } from '../../vector/ArrayVector';
import { AlignedData, join } from './joinDataFrames';
import { getDisplayProcessor } from '../../field';
import { createTheme, GrafanaTheme2 } from '../../themes';

/**
 * @internal
 */
/* eslint-disable */
// prettier-ignore
export const histogramBucketSizes = [
  1e-9,  2e-9,  2.5e-9,  4e-9,  5e-9,
  1e-8,  2e-8,  2.5e-8,  4e-8,  5e-8,
  1e-7,  2e-7,  2.5e-7,  4e-7,  5e-7,
  1e-6,  2e-6,  2.5e-6,  4e-6,  5e-6,
  1e-5,  2e-5,  2.5e-5,  4e-5,  5e-5,
  1e-4,  2e-4,  2.5e-4,  4e-4,  5e-4,
  1e-3,  2e-3,  2.5e-3,  4e-3,  5e-3,
  1e-2,  2e-2,  2.5e-2,  4e-2,  5e-2,
  1e-1,  2e-1,  2.5e-1,  4e-1,  5e-1,
  1,     2,              4,     5,
  1e+1,  2e+1,  2.5e+1,  4e+1,  5e+1,
  1e+2,  2e+2,  2.5e+2,  4e+2,  5e+2,
  1e+3,  2e+3,  2.5e+3,  4e+3,  5e+3,
  1e+4,  2e+4,  2.5e+4,  4e+4,  5e+4,
  1e+5,  2e+5,  2.5e+5,  4e+5,  5e+5,
  1e+6,  2e+6,  2.5e+6,  4e+6,  5e+6,
  1e+7,  2e+7,  2.5e+7,  4e+7,  5e+7,
  1e+8,  2e+8,  2.5e+8,  4e+8,  5e+8,
  1e+9,  2e+9,  2.5e+9,  4e+9,  5e+9,
];
/* eslint-enable */

const histFilter = [null];
const histSort = (a: number, b: number) => a - b;

/**
 * @alpha
 */
export interface HistogramTransformerOptions {
  bucketSize?: number; // 0 is auto
  bucketOffset?: number;
  // xMin?: number;
  // xMax?: number;
  combine?: boolean; // if multiple series are input, join them into one
}

/**
 * This is a helper class to use the same text in both a panel and transformer UI
 *
 * @internal
 */
export const histogramFieldInfo = {
  bucketSize: {
    name: 'Bucket size',
    description: undefined,
  },
  bucketOffset: {
    name: 'Bucket offset',
    description: 'for non-zero-based buckets',
  },
  combine: {
    name: 'Combine series',
    description: 'combine all series into a single histogram',
  },
};

/**
 * @alpha
 */
export const histogramTransformer: SynchronousDataTransformerInfo<HistogramTransformerOptions> = {
  id: DataTransformerID.histogram,
  name: 'Histogram',
  description: 'Calculate a histogram from input data',
  defaultOptions: {
    fields: {},
  },

  operator: (options) => (source) => source.pipe(map((data) => histogramTransformer.transformer(options)(data))),

  transformer: (options: HistogramTransformerOptions) => (data: DataFrame[]) => {
    if (!Array.isArray(data) || data.length === 0) {
      return data;
    }
    const hist = buildHistogram(data, options);
    if (hist == null) {
      return [];
    }
    return [histogramFieldsToFrame(hist)];
  },
};

/**
 * @internal
 */
export const histogramFrameBucketMinFieldName = 'BucketMin';

/**
 * @internal
 */
export const histogramFrameBucketMaxFieldName = 'BucketMax';

/**
 * @alpha
 */
export interface HistogramFields {
  bucketMin: Field;
  bucketMax: Field;
  counts: Field[]; // frequency
}

/**
 * Given a frame, find the explicit histogram fields
 *
 * @alpha
 */
export function getHistogramFields(frame: DataFrame): HistogramFields | undefined {
  let bucketMin: Field | undefined = undefined;
  let bucketMax: Field | undefined = undefined;
  const counts: Field[] = [];
  for (const field of frame.fields) {
    if (field.name === histogramFrameBucketMinFieldName) {
      bucketMin = field;
    } else if (field.name === histogramFrameBucketMaxFieldName) {
      bucketMax = field;
    } else if (field.type === FieldType.number) {
      counts.push(field);
    }
  }
  if (bucketMin && bucketMax && counts.length) {
    return {
      bucketMin,
      bucketMax,
      counts,
    };
  }
  return undefined;
}

const APPROX_BUCKETS = 20;

/**
 * @alpha
 */
export function buildHistogram(frames: DataFrame[], options?: HistogramTransformerOptions): HistogramFields | null {
  let bucketSize = options?.bucketSize;
  let bucketOffset = options?.bucketOffset ?? 0;

  // if bucket size is auto, try to calc from all numeric fields
  if (!bucketSize) {
    let allValues: number[] = [];

    // TODO: include field configs!
    for (const frame of frames) {
      for (const field of frame.fields) {
        if (field.type === FieldType.number) {
          allValues = allValues.concat(field.values.toArray());
        }
      }
    }

    allValues.sort((a, b) => a - b);

    let smallestDelta = Infinity;

    // TODO: case of 1 value needs work
    if (allValues.length === 1) {
      smallestDelta = 1;
    } else {
      for (let i = 1; i < allValues.length; i++) {
        let delta = allValues[i] - allValues[i - 1];

        if (delta !== 0) {
          smallestDelta = Math.min(smallestDelta, delta);
        }
      }
    }

    let min = allValues[0];
    let max = allValues[allValues.length - 1];

    let range = max - min;

    const targetSize = range / APPROX_BUCKETS;

    // choose bucket
    for (let i = 0; i < histogramBucketSizes.length; i++) {
      let _bucketSize = histogramBucketSizes[i];

      if (targetSize < _bucketSize && _bucketSize >= smallestDelta) {
        bucketSize = _bucketSize;
        break;
      }
    }
  }

  const getBucket = (v: number) => incrRoundDn(v - bucketOffset, bucketSize!) + bucketOffset;

  let histograms: AlignedData[] = [];
  let counts: Field[] = [];
  let config: FieldConfig | undefined = undefined;

  for (const frame of frames) {
    for (const field of frame.fields) {
      if (field.type === FieldType.number) {
        let fieldHist = histogram(field.values.toArray(), getBucket, histFilter, histSort) as AlignedData;
        histograms.push(fieldHist);
        counts.push({
          ...field,
          config: {
            ...field.config,
            unit: undefined,
          },
        });
        if (!config && field.config.unit) {
          config = field.config;
        }
      }
    }
  }

  // Quit early for empty a
  if (!counts.length) {
    return null;
  }

  // align histograms
  let joinedHists = join(histograms);

  // zero-fill all undefined values (missing buckets -> 0 counts)
  for (let histIdx = 1; histIdx < joinedHists.length; histIdx++) {
    let hist = joinedHists[histIdx];

    for (let bucketIdx = 0; bucketIdx < hist.length; bucketIdx++) {
      if (hist[bucketIdx] == null) {
        hist[bucketIdx] = 0;
      }
    }
  }

  const bucketMin: Field = {
    name: histogramFrameBucketMinFieldName,
    values: new ArrayVector(joinedHists[0]),
    type: FieldType.number,
    state: undefined,
    config: config ?? {},
  };
  const bucketMax = {
    ...bucketMin,
    name: histogramFrameBucketMaxFieldName,
    values: new ArrayVector(joinedHists[0].map((v) => v + bucketSize!)),
  };

  if (options?.combine) {
    const vals = new Array(joinedHists[0].length).fill(0);
    for (let i = 1; i < joinedHists.length; i++) {
      for (let j = 0; j < vals.length; j++) {
        vals[j] += joinedHists[i][j];
      }
    }
    counts = [
      {
        ...counts[0],
        name: 'Count',
        values: new ArrayVector(vals),
        type: FieldType.number,
        state: undefined,
      },
    ];
  } else {
    counts.forEach((field, i) => {
      field.values = new ArrayVector(joinedHists[i + 1]);
    });
  }

  return {
    bucketMin,
    bucketMax,
    counts,
  };
}

// function incrRound(num: number, incr: number) {
//   return Math.round(num / incr) * incr;
// }

// function incrRoundUp(num: number, incr: number) {
//   return Math.ceil(num / incr) * incr;
// }

function incrRoundDn(num: number, incr: number) {
  return Math.floor(num / incr) * incr;
}

function histogram(
  vals: number[],
  getBucket: (v: number) => number,
  filterOut?: any[] | null,
  sort?: ((a: any, b: any) => number) | null
) {
  let hist = new Map();

  for (let i = 0; i < vals.length; i++) {
    let v = vals[i];

    if (v != null) {
      v = getBucket(v);
    }

    let entry = hist.get(v);

    if (entry) {
      entry.count++;
    } else {
      hist.set(v, { value: v, count: 1 });
    }
  }

  filterOut && filterOut.forEach((v) => hist.delete(v));

  let bins = [...hist.values()];

  sort && bins.sort((a, b) => sort(a.value, b.value));

  let values = Array(bins.length);
  let counts = Array(bins.length);

  for (let i = 0; i < bins.length; i++) {
    values[i] = bins[i].value;
    counts[i] = bins[i].count;
  }

  return [values, counts];
}

/**
 * @internal
 */
export function histogramFieldsToFrame(info: HistogramFields, theme?: GrafanaTheme2): DataFrame {
  if (!info.bucketMin.display) {
    const display = getDisplayProcessor({
      field: info.bucketMin,
      theme: theme ?? createTheme(),
    });
    info.bucketMin.display = display;
    info.bucketMax.display = display;
  }
  return {
    fields: [info.bucketMin, info.bucketMax, ...info.counts],
    length: info.bucketMin.values.length,
  };
}
