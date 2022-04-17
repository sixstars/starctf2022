import {
  DataLink,
  DataQuery,
  Field,
  InternalDataLink,
  InterpolateFunction,
  LinkModel,
  ScopedVars,
  TimeRange,
} from '../types';
import { locationUtil } from './location';
import { serializeStateToUrlParam } from './url';

export const DataLinkBuiltInVars = {
  keepTime: '__url_time_range',
  timeRangeFrom: '__from',
  timeRangeTo: '__to',
  includeVars: '__all_variables',
  seriesName: '__series.name',
  fieldName: '__field.name',
  valueTime: '__value.time',
  valueNumeric: '__value.numeric',
  valueText: '__value.text',
  valueRaw: '__value.raw',
  // name of the calculation represented by the value
  valueCalc: '__value.calc',
};

// We inject these because we cannot import them directly as they reside inside grafana main package.
export type LinkToExploreOptions = {
  link: DataLink;
  scopedVars: ScopedVars;
  range: TimeRange;
  field: Field;
  internalLink: InternalDataLink;
  onClickFn?: (options: { datasourceUid: string; query: any; range?: TimeRange }) => void;
  replaceVariables: InterpolateFunction;
};

export function mapInternalLinkToExplore(options: LinkToExploreOptions): LinkModel<Field> {
  const { onClickFn, replaceVariables, link, scopedVars, range, field, internalLink } = options;

  const interpolatedQuery = interpolateQuery(link, scopedVars, replaceVariables);
  const title = link.title ? link.title : internalLink.datasourceName;

  return {
    title: replaceVariables(title, scopedVars),
    // In this case this is meant to be internal link (opens split view by default) the href will also points
    // to explore but this way you can open it in new tab.
    href: generateInternalHref(internalLink.datasourceName, interpolatedQuery, range),
    onClick: onClickFn
      ? () => {
          onClickFn({
            datasourceUid: internalLink.datasourceUid,
            query: interpolatedQuery,
            range,
          });
        }
      : undefined,
    target: '_self',
    origin: field,
  };
}

/**
 * Generates href for internal derived field link.
 */
function generateInternalHref<T extends DataQuery = any>(datasourceName: string, query: T, range: TimeRange): string {
  return locationUtil.assureBaseUrl(
    `/explore?left=${encodeURIComponent(
      serializeStateToUrlParam({
        range: range.raw,
        datasource: datasourceName,
        queries: [query],
      })
    )}`
  );
}

function interpolateQuery<T extends DataQuery = any>(
  link: DataLink,
  scopedVars: ScopedVars,
  replaceVariables: InterpolateFunction
): T {
  let stringifiedQuery = '';
  try {
    stringifiedQuery = JSON.stringify(link.internal?.query || '');
  } catch (err) {
    // should not happen and not much to do about this, possibly something non stringifiable in the query
    console.error(err);
  }

  // Replace any variables inside the query. This may not be the safest as it can also replace keys etc so may not
  // actually work with every datasource query right now.
  stringifiedQuery = replaceVariables(stringifiedQuery, scopedVars);

  let replacedQuery = {} as T;
  try {
    replacedQuery = JSON.parse(stringifiedQuery);
  } catch (err) {
    // again should not happen and not much to do about this, probably some issue with how we replaced the variables.
    console.error(stringifiedQuery, err);
  }

  return replacedQuery;
}
