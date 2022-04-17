import { DataFrame } from '../types/dataFrame';

/**
 * Returns true if both frames have the same name, fields, labels and configs.
 *
 * @example
 * To compare multiple frames use:
 * ```
 * compareArrayValues(a, b, framesHaveSameStructure);
 * ```
 * NOTE: this does a shallow check on the FieldConfig properties, when using the query
 * editor, this should be sufficient, however if applications are mutating properties
 * deep in the FieldConfig this will not recognize a change
 *
 * @beta
 */
export function compareDataFrameStructures(a: DataFrame, b: DataFrame, skipConfig?: boolean): boolean {
  if (a === b) {
    return true;
  }

  if (a?.fields?.length !== b?.fields?.length) {
    return false;
  }

  if (a.name !== b.name) {
    return false;
  }

  for (let i = 0; i < a.fields.length; i++) {
    const fA = a.fields[i];
    const fB = b.fields[i];

    if (fA.type !== fB.type || fA.name !== fB.name) {
      return false;
    }

    // Do not check the config fields
    if (skipConfig) {
      continue;
    }

    // Check if labels are different
    if (fA.labels && fB.labels && !shallowCompare(fA.labels, fB.labels)) {
      return false;
    }

    const cfgA = fA.config as any;
    const cfgB = fB.config as any;

    let aKeys = Object.keys(cfgA);
    let bKeys = Object.keys(cfgB);

    if (aKeys.length !== bKeys.length) {
      return false;
    }

    for (const key of aKeys) {
      if (!(key in cfgB)) {
        return false;
      }

      if (key === 'custom') {
        if (!shallowCompare(cfgA[key], cfgB[key])) {
          return false;
        }
      } else if (cfgA[key] !== cfgB[key]) {
        return false;
      }
    }
  }

  return true;
}

/**
 * Check if all values in two arrays match the compare funciton
 *
 * @beta
 */
export function compareArrayValues<T>(a: T[], b: T[], cmp: (a: T, b: T) => boolean) {
  if (a === b) {
    return true;
  }
  if (a?.length !== b?.length) {
    return false;
  }
  for (let i = 0; i < a.length; i++) {
    if (!cmp(a[i], b[i])) {
      return false;
    }
  }
  return true;
}

type Cmp = (valA: any, valB: any) => boolean;

const defaultCmp: Cmp = (a, b) => a === b;

/**
 * Checks if two objects are equal shallowly
 *
 * @beta
 */
export function shallowCompare<T extends {}>(a: T, b: T, cmp: Cmp = defaultCmp) {
  if (a === b) {
    return true;
  }

  const aKeys = Object.keys(a);
  const bKeys = Object.keys(b);

  if (aKeys.length !== bKeys.length) {
    return false;
  }

  for (let key of aKeys) {
    //@ts-ignore
    if (!cmp(a[key], b[key])) {
      return false;
    }
  }

  return true;
}
