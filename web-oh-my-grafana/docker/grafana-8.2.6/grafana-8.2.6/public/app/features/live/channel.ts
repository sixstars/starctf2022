import {
  LiveChannelConfig,
  LiveChannelStatusEvent,
  LiveChannelEvent,
  LiveChannelEventType,
  LiveChannelConnectionState,
  LiveChannelPresenceStatus,
  LiveChannelAddress,
  DataFrameJSON,
} from '@grafana/data';
import Centrifuge, {
  JoinLeaveContext,
  PublicationContext,
  SubscribeErrorContext,
  SubscribeSuccessContext,
  SubscriptionEvents,
  UnsubscribeContext,
} from 'centrifuge/dist/centrifuge';

import { Subject, of, Observable } from 'rxjs';

/**
 * Internal class that maps Centrifuge support to GrafanaLive
 */
export class CentrifugeLiveChannel<T = any> {
  readonly currentStatus: LiveChannelStatusEvent;

  readonly opened = Date.now();
  readonly id: string;
  readonly addr: LiveChannelAddress;

  readonly stream = new Subject<LiveChannelEvent<T>>();

  // Hold on to the last header with schema
  lastMessageWithSchema?: DataFrameJSON;

  /** Static definition of the channel definition.  This may describe the channel usage */
  config?: LiveChannelConfig;
  subscription?: Centrifuge.Subscription;
  shutdownCallback?: () => void;

  constructor(id: string, addr: LiveChannelAddress) {
    this.id = id;
    this.addr = addr;
    this.currentStatus = {
      type: LiveChannelEventType.Status,
      id,
      timestamp: this.opened,
      state: LiveChannelConnectionState.Pending,
    };
  }

  // This should only be called when centrifuge is connected
  initalize(config: LiveChannelConfig): SubscriptionEvents {
    if (this.config) {
      throw new Error('Channel already initalized: ' + this.id);
    }
    this.config = config;

    const events: SubscriptionEvents = {
      // Called when a message is recieved from the socket
      publish: (ctx: PublicationContext) => {
        try {
          if (ctx.data) {
            if (ctx.data.schema) {
              this.lastMessageWithSchema = ctx.data as DataFrameJSON;
            }

            this.stream.next({
              type: LiveChannelEventType.Message,
              message: ctx.data,
            });
          }

          // Clear any error messages
          if (this.currentStatus.error) {
            this.currentStatus.timestamp = Date.now();
            delete this.currentStatus.error;
            this.sendStatus();
          }
        } catch (err) {
          console.log('publish error', this.addr, err);
          this.currentStatus.error = err;
          this.currentStatus.timestamp = Date.now();
          this.sendStatus();
        }
      },
      error: (ctx: SubscribeErrorContext) => {
        this.currentStatus.timestamp = Date.now();
        this.currentStatus.error = ctx.error;
        this.sendStatus();
      },
      subscribe: (ctx: SubscribeSuccessContext) => {
        this.currentStatus.timestamp = Date.now();
        this.currentStatus.state = LiveChannelConnectionState.Connected;
        delete this.currentStatus.error;

        if (ctx.data?.schema) {
          this.lastMessageWithSchema = ctx.data as DataFrameJSON;
        }

        this.sendStatus(ctx.data);
      },
      unsubscribe: (ctx: UnsubscribeContext) => {
        this.currentStatus.timestamp = Date.now();
        this.currentStatus.state = LiveChannelConnectionState.Disconnected;
        this.sendStatus();
      },
    };

    if (config.hasPresence) {
      events.join = (ctx: JoinLeaveContext) => {
        this.stream.next({ type: LiveChannelEventType.Join, user: ctx.info.user });
      };
      events.leave = (ctx: JoinLeaveContext) => {
        this.stream.next({ type: LiveChannelEventType.Leave, user: ctx.info.user });
      };
    }
    return events;
  }

  private sendStatus(message?: any) {
    const copy = { ...this.currentStatus };
    if (message) {
      copy.message = message;
    }
    this.stream.next(copy);
  }

  disconnectIfNoListeners = () => {
    const count = this.stream.observers.length;
    if (count === 0) {
      this.disconnect();
    }
  };

  /**
   * Get the stream of events and
   */
  getStream() {
    return new Observable((subscriber) => {
      subscriber.next({ ...this.currentStatus });
      const sub = this.stream.subscribe(subscriber);
      return () => {
        sub.unsubscribe();
        const count = this.stream.observers.length;

        // Wait 1/4 second to fully disconnect
        if (count === 0) {
          setTimeout(this.disconnectIfNoListeners, 250);
        }
      };
    }) as Observable<LiveChannelEvent<T>>;
  }

  /**
   * This is configured by the server when the config supports presence
   */
  async getPresence(): Promise<LiveChannelPresenceStatus> {
    if (!this.subscription) {
      return Promise.reject('not subscribed');
    }

    return this.subscription!.presence().then((v) => {
      return {
        users: Object.keys(v.presence),
      };
    });
  }

  /**
   * This will close and terminate all streams for this channel
   */
  disconnect() {
    this.currentStatus.state = LiveChannelConnectionState.Shutdown;
    this.currentStatus.timestamp = Date.now();

    if (this.subscription) {
      this.subscription.unsubscribe();
      this.subscription.removeAllListeners(); // they keep all listeners attached after unsubscribe
      this.subscription = undefined;
    }

    this.stream.complete();

    this.stream.next({ ...this.currentStatus });
    this.stream.complete();

    if (this.shutdownCallback) {
      this.shutdownCallback();
    }
  }

  shutdownWithError(err: string) {
    this.currentStatus.error = err;
    this.sendStatus();
    this.disconnect();
  }
}

export function getErrorChannel<TMessage>(msg: string, id: string, addr: LiveChannelAddress) {
  return {
    id,
    opened: Date.now(),
    addr,

    // return an error
    getStream: () =>
      of({
        type: LiveChannelEventType.Status,
        id,
        timestamp: Date.now(),
        state: LiveChannelConnectionState.Invalid,
        error: msg,
      }),

    // already disconnected
    disconnect: () => {},
  };
}
