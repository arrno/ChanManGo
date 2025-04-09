// src/types.ts
var ConnectionState = /* @__PURE__ */ ((ConnectionState2) => {
  ConnectionState2["DISCONNECTED"] = "disconnected";
  ConnectionState2["CONNECTING"] = "connecting";
  ConnectionState2["CONNECTED"] = "connected";
  ConnectionState2["RECONNECTING"] = "reconnecting";
  ConnectionState2["ERROR"] = "error";
  return ConnectionState2;
})(ConnectionState || {});
var TOPIC_ALL = "_ALL_";

// src/sdk.ts
import * as ws from "ws";
var WebSocketImpl = typeof WebSocket !== "undefined" ? WebSocket : ws.WebSocket;
var ChanbanSDK = class {
  /**
   * Creates a new instance of the ChanbanSDK
   *
   * @param options Configuration options for the SDK
   */
  constructor(options) {
    // WebSocket connections
    this.pubSocket = null;
    this.subSocket = null;
    // Connection state tracking
    this.pubState = "disconnected" /* DISCONNECTED */;
    this.subState = "disconnected" /* DISCONNECTED */;
    // Reconnection tracking
    this.pubReconnectAttempts = 0;
    this.subReconnectAttempts = 0;
    // Topic subscriptions
    this.listeners = {};
    // Connection state change listeners
    this.stateChangeListeners = [];
    this.baseUrl = options.baseUrl;
    this.authToken = options.authToken;
    this.channel = options.channel;
    this.rootTopics = options.topics ?? [];
    this.headers = { ...options.headers };
    if (!this.headers["Authorization"]) {
      if (!this.authToken.startsWith("Bearer ")) {
        this.headers["Authorization"] = `Bearer ${this.authToken}`;
      } else {
        this.headers["Authorization"] = this.authToken;
      }
    }
    this.autoReconnect = options.autoReconnect ?? true;
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 5;
    this.reconnectDelay = options.reconnectDelay ?? 3e3;
    this.debug = options.debug ?? false;
    this.logDebug(
      "Initialized with options:",
      JSON.stringify({
        baseUrl: this.baseUrl,
        channel: this.channel,
        autoReconnect: this.autoReconnect,
        maxReconnectAttempts: this.maxReconnectAttempts,
        reconnectDelay: this.reconnectDelay,
        debug: this.debug
      })
    );
  }
  /**
   * Handling storing the new auth token and updating headers
   *
   * @param token The new auth token
   */
  updateAuthToken(token) {
    this.authToken = token;
    if (this.authToken.startsWith("Bearer ")) {
      this.headers["Authorization"] = this.authToken;
    } else {
      this.headers["Authorization"] = `Bearer ${this.authToken}`;
    }
  }
  /**
   * Rotates the account token via HTTP POST
   *
   * @returns Promise that resolves with the server response
   */
  async rotateToken() {
    const response = await fetch(`${this.baseUrl}/rotate-token`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...this.headers
      },
      body: `{}`
    });
    if (!response.ok) {
      throw new Error(
        `HTTP error: ${response.status} ${response.statusText}`
      );
    }
    const accountInfo = await response.json();
    if (accountInfo.token) {
      this.updateAuthToken(accountInfo.token);
    }
    return accountInfo;
  }
  /**
   * Sets up a WebSocket connection for publishing messages
   *
   * @returns Promise that resolves when the connection is established
   */
  async connectPublisher() {
    if (this.isSocketConnected(this.pubSocket)) {
      return;
    }
    this.updatePublisherState("connecting" /* CONNECTING */);
    const queryParams = Object.entries(this.headers).map(
      ([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`
    ).join("&");
    const url = `${this.getWebSocketBaseUrl()}/pub-ws?channel=${encodeURIComponent(
      this.channel
    )}&${queryParams}`;
    this.pubSocket = new WebSocketImpl(url);
    this.logDebug(`Attempting to connect publisher to ${url}`);
    return new Promise((resolve, reject) => {
      if (!this.pubSocket) {
        this.updatePublisherState("error" /* ERROR */);
        reject(new Error("Failed to create publisher WebSocket"));
        return;
      }
      this.pubSocket.onopen = () => {
        this.updatePublisherState("connected" /* CONNECTED */);
        this.pubReconnectAttempts = 0;
        this.logDebug("Publisher connected");
        resolve();
      };
      this.pubSocket.onclose = (event) => {
        const wasConnected = this.pubState === "connected" /* CONNECTED */;
        this.updatePublisherState("disconnected" /* DISCONNECTED */);
        this.logDebug(
          `Publisher disconnected: ${event?.code} ${event?.reason}`
        );
        if (wasConnected && this.autoReconnect) {
          this.attemptReconnectPublisher();
        }
        if (!wasConnected) {
          reject(
            new Error(
              `WebSocket error: ${event?.message || "Denied"}`
            )
          );
        }
      };
      this.pubSocket.onerror = (event) => {
        this.updatePublisherState("error" /* ERROR */);
        this.logDebug(
          `Publisher error: ${event?.message || "Unknown error"}`
        );
        reject(
          new Error(
            `WebSocket error: ${event?.message || "Unknown error"}`
          )
        );
      };
    });
  }
  /**
   * Sets up a WebSocket connection for subscribing to messages
   *
   * @returns Promise that resolves when the connection is established
   */
  async connectSubscriber() {
    if (this.isSocketConnected(this.subSocket)) {
      return;
    }
    this.updateSubscriberState("connecting" /* CONNECTING */);
    let queryParams = Object.entries(this.headers).map(
      ([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`
    ).join("&");
    if (this.rootTopics.length > 0) {
      queryParams += `&topics=${this.rootTopics.map((topic) => encodeURIComponent(topic)).join(",")}`;
    }
    const url = `${this.getWebSocketBaseUrl()}/sub?channel=${encodeURIComponent(
      this.channel
    )}&${queryParams}`;
    this.subSocket = new WebSocketImpl(url);
    this.logDebug(`Attempting to connect subscriber to ${url}`);
    return new Promise((resolve, reject) => {
      if (!this.subSocket) {
        this.updateSubscriberState("error" /* ERROR */);
        reject(new Error("Failed to create subscriber WebSocket"));
        return;
      }
      this.subSocket.onopen = () => {
        this.updateSubscriberState("connected" /* CONNECTED */);
        this.subReconnectAttempts = 0;
        this.logDebug("Subscriber connected");
        resolve();
      };
      this.subSocket.onclose = (event) => {
        const wasConnected = this.subState === "connected" /* CONNECTED */;
        this.updateSubscriberState("disconnected" /* DISCONNECTED */);
        this.logDebug(
          `Subscriber disconnected: ${event?.code} ${event?.reason}`
        );
        if (wasConnected && this.autoReconnect) {
          this.attemptReconnectSubscriber();
        }
        if (!wasConnected) {
          reject(
            new Error(
              `WebSocket error: ${event?.message || "Denied"}`
            )
          );
        }
      };
      this.subSocket.onerror = (event) => {
        this.updateSubscriberState("error" /* ERROR */);
        this.logDebug(
          `Subscriber error: ${event?.message || "Unknown error"}`
        );
        reject(
          new Error(
            `WebSocket error: ${event?.message || "Unknown error"}`
          )
        );
      };
      this.subSocket.onmessage = (event) => {
        try {
          const message = JSON.parse(
            event.data.toString()
          );
          const listeners = [
            ...this.listeners[message.topic] ?? [],
            ...this.listeners[TOPIC_ALL] ?? []
          ];
          if (listeners && listeners.length > 0) {
            let parsedPayload;
            try {
              parsedPayload = JSON.parse(message.payload);
            } catch (error) {
              this.logDebug(
                `Failed to parse message payload: ${error}`
              );
              return;
            }
            listeners.forEach((callback) => {
              try {
                callback(message.topic, parsedPayload);
              } catch (error) {
                this.logDebug(
                  `Error in message callback: ${error}`
                );
              }
            });
          }
        } catch (error) {
          this.logDebug(
            `Error processing WebSocket message: ${error}`
          );
        }
      };
    });
  }
  /**
   * Attempts to reconnect the publisher socket
   */
  attemptReconnectPublisher() {
    if (this.pubReconnectAttempts >= this.maxReconnectAttempts) {
      this.logDebug(
        `Max publisher reconnect attempts (${this.maxReconnectAttempts}) reached, giving up`
      );
      return;
    }
    this.updatePublisherState("reconnecting" /* RECONNECTING */);
    this.pubReconnectAttempts++;
    this.logDebug(
      `Attempting to reconnect publisher (attempt ${this.pubReconnectAttempts}/${this.maxReconnectAttempts})...`
    );
    setTimeout(() => {
      this.connectPublisher().catch((err) => {
        this.logDebug(`Failed to reconnect publisher: ${err.message}`);
        this.attemptReconnectPublisher();
      });
    }, this.reconnectDelay);
  }
  /**
   * Attempts to reconnect the subscriber socket
   */
  attemptReconnectSubscriber() {
    if (this.subReconnectAttempts >= this.maxReconnectAttempts) {
      this.logDebug(
        `Max subscriber reconnect attempts (${this.maxReconnectAttempts}) reached, giving up`
      );
      return;
    }
    this.updateSubscriberState("reconnecting" /* RECONNECTING */);
    this.subReconnectAttempts++;
    this.logDebug(
      `Attempting to reconnect subscriber (attempt ${this.subReconnectAttempts}/${this.maxReconnectAttempts})...`
    );
    setTimeout(() => {
      this.connectSubscriber().catch((err) => {
        this.logDebug(`Failed to reconnect subscriber: ${err.message}`);
        this.attemptReconnectSubscriber();
      });
    }, this.reconnectDelay);
  }
  /**
   * Updates the publisher connection state and notifies listeners
   *
   * @param state New connection state
   */
  updatePublisherState(state) {
    this.pubState = state;
    this.notifyStateChangeListeners(state, "publisher");
  }
  /**
   * Updates the subscriber connection state and notifies listeners
   *
   * @param state New connection state
   */
  updateSubscriberState(state) {
    this.subState = state;
    this.notifyStateChangeListeners(state, "subscriber");
  }
  /**
   * Notifies all state change listeners of a state change
   *
   * @param state New connection state
   * @param connectionType Type of connection that changed state
   */
  notifyStateChangeListeners(state, connectionType) {
    this.stateChangeListeners.forEach((listener) => {
      try {
        listener(state, connectionType);
      } catch (error) {
        this.logDebug(`Error in state change listener: ${error}`);
      }
    });
  }
  /**
   * Publishes a message to a topic
   *
   * @param topic Topic to publish to
   * @param payload Message payload to publish
   * @returns Promise that resolves when the message is published
   */
  async publish(topic, payload) {
    await this.connectPublisher();
    const message = {
      topic,
      payload: JSON.stringify(payload)
    };
    this.pubSocket?.send(JSON.stringify(message));
  }
  /**
   * Publishes a message via HTTP POST instead of WebSocket
   *
   * @param topic Topic to publish to
   * @param payload Message payload to publish
   * @returns Promise that resolves with the server response
   */
  async publishHttp(topic, payload) {
    const message = {
      topic,
      payload: JSON.stringify(payload)
    };
    const response = await fetch(
      `${this.baseUrl}/pub?channel=${encodeURIComponent(this.channel)}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...this.headers
        },
        body: JSON.stringify(message)
      }
    );
    if (!response.ok) {
      throw new Error(
        `HTTP error: ${response.status} ${response.statusText}`
      );
    }
    return await response.json();
  }
  /**
   * Subscribes to one or more topics
   *
   * @param callback Callback function to call when a message is received
   * @param topics Optional topic or topics to subscribe to filtered on client
   * @returns Promise that resolves when the subscription is established
   */
  async subscribe(callback, topics) {
    topics = topics ?? TOPIC_ALL;
    const topicArray = Array.isArray(topics) ? topics : [topics];
    await this.connectSubscriber();
    topicArray.forEach((topic) => {
      if (!this.listeners[topic]) {
        this.listeners[topic] = [];
      }
      this.listeners[topic].push(callback);
    });
  }
  /**
   * Unsubscribes from a topic
   *
   * @param topic Topic to unsubscribe from
   * @param callback Optional callback function to remove. If not provided, all callbacks for the topic will be removed.
   */
  unsubscribe(topic, callback) {
    if (!this.listeners[topic]) {
      return;
    }
    if (callback) {
      this.listeners[topic] = this.listeners[topic].filter(
        (cb) => cb !== callback
      );
      if (this.listeners[topic].length === 0) {
        delete this.listeners[topic];
      }
    } else {
      delete this.listeners[topic];
    }
  }
  /**
   * Disconnects from the server
   */
  disconnect() {
    if (this.pubSocket) {
      this.updatePublisherState("disconnected" /* DISCONNECTED */);
      this.pubSocket.close();
      this.pubSocket = null;
    }
    if (this.subSocket) {
      this.updateSubscriberState("disconnected" /* DISCONNECTED */);
      this.subSocket.close();
      this.subSocket = null;
    }
    this.logDebug("Disconnected from server");
  }
  /**
   * Adds a listener for connection state changes
   *
   * @param listener Function to call when the connection state changes
   * @returns Function to remove the listener
   */
  onStateChange(listener) {
    this.stateChangeListeners.push(listener);
    return () => {
      this.stateChangeListeners = this.stateChangeListeners.filter(
        (l) => l !== listener
      );
    };
  }
  /**
   * Gets the current connection state for the publisher
   */
  getPublisherState() {
    return this.pubState;
  }
  /**
   * Gets the current connection state for the subscriber
   */
  getSubscriberState() {
    return this.subState;
  }
  /**
   * Checks if a socket is connected
   *
   * @param socket WebSocket to check
   * @returns True if the socket is connected
   */
  isSocketConnected(socket) {
    if (!socket) return false;
    return socket.readyState === 1;
  }
  /**
   * Gets the WebSocket base URL (ws:// or wss://) from the HTTP base URL
   *
   * @returns WebSocket base URL
   */
  getWebSocketBaseUrl() {
    if (this.baseUrl.startsWith("https://")) {
      return this.baseUrl.replace("https://", "wss://");
    } else if (this.baseUrl.startsWith("http://")) {
      return this.baseUrl.replace("http://", "ws://");
    } else {
      return `ws://${this.baseUrl}`;
    }
  }
  /**
   * Logs a debug message if debug mode is enabled
   *
   * @param message Message to log
   * @param args Additional arguments to log
   */
  logDebug(message, ...args) {
    if (this.debug) {
      if (args.length > 0) {
        console.debug(`[ChanbanSDK] ${message}`, ...args);
      } else {
        console.debug(`[ChanbanSDK] ${message}`);
      }
    }
  }
};

// src/env.ts
function loadFromEnv() {
  const baseUrl = process.env.CHANBAN_BASE_URL;
  const authToken = process.env.CHANBAN_AUTH_TOKEN;
  const channel = process.env.CHANBAN_CHANNEL;
  if (!baseUrl) {
    throw new Error("CHANBAN_BASE_URL environment variable is required");
  }
  if (!authToken) {
    throw new Error("CHANBAN_AUTH_TOKEN environment variable is required");
  }
  if (!channel) {
    throw new Error("CHANBAN_CHANNEL environment variable is required");
  }
  const debug = process.env.CHANBAN_DEBUG === "true";
  const autoReconnect = process.env.CHANBAN_AUTO_RECONNECT !== "false";
  const maxReconnectAttempts = parseInt(process.env.CHANBAN_MAX_RECONNECT_ATTEMPTS || "5", 10);
  const reconnectDelay = parseInt(process.env.CHANBAN_RECONNECT_DELAY || "3000", 10);
  return {
    baseUrl,
    authToken,
    channel,
    debug,
    autoReconnect,
    maxReconnectAttempts,
    reconnectDelay
  };
}
export {
  ChanbanSDK,
  ConnectionState,
  loadFromEnv
};
