import { getEchoSrv, EchoEventType } from '../services/EchoSrv';
import {
  InteractionEchoEvent,
  MetaAnalyticsEvent,
  MetaAnalyticsEventPayload,
  PageviewEchoEvent,
} from '../types/analytics';
import { locationService } from '../services';
import { config } from '../config';

/**
 * Helper function to report meta analytics to the {@link EchoSrv}.
 *
 * @public
 */
export const reportMetaAnalytics = (payload: MetaAnalyticsEventPayload) => {
  getEchoSrv().addEvent<MetaAnalyticsEvent>({
    type: EchoEventType.MetaAnalytics,
    payload,
  });
};

/**
 * Helper function to report pageview events to the {@link EchoSrv}.
 *
 * @public
 */
export const reportPageview = () => {
  const location = locationService.getLocation();
  const page = `${config.appSubUrl ?? ''}${location.pathname}${location.search}${location.hash}`;
  getEchoSrv().addEvent<PageviewEchoEvent>({
    type: EchoEventType.Pageview,
    payload: {
      page,
    },
  });
};

/**
 * Helper function to report interaction events to the {@link EchoSrv}.
 *
 * @public
 */
export const reportInteraction = (interactionName: string, properties?: Record<string, any>) => {
  getEchoSrv().addEvent<InteractionEchoEvent>({
    type: EchoEventType.Interaction,
    payload: {
      interactionName,
      properties,
    },
  });
};
