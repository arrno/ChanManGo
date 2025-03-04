/**
 * StreamingSDK provides real-time message publishing and subscribing.
 */
class StreamingSDK {
  /**
   * @param {Object} config - Configuration options.
   * @param {string} config.baseUrl - Base URL for the WebSocket API.
   * @param {string} config.authToken - Authentication token.
   * @param {string} config.channel - Channel name for messaging.
   * @param {Object} [config.headers={}] - Additional headers.
   */
  constructor({ baseUrl, authToken, channel, headers = {} }) {
    this.baseUrl = baseUrl;
    this.authToken = authToken;
    this.channel = channel;
    this.headers = { ...headers, Authorization: `Bearer ${authToken}` };
    this.pubSocket = null;
    this.subSocket = null;
    this.listeners = {};
  }

  /**
   * Ensures the publisher WebSocket is connected.
   * @private
   * @returns {Promise<void>}
   */
  async ensurePublisherConnected() {
    if (!this.pubSocket || this.pubSocket.readyState !== WebSocket.OPEN) {
      this.pubSocket = new WebSocket(`${this.baseUrl}/ws/publish?channel=${this.channel}`, []);
      await new Promise((resolve, reject) => {
        this.pubSocket.onopen = () => {
          Object.entries(this.headers).forEach(([key, value]) => {
            this.pubSocket.send(JSON.stringify({ type: 'header', key, value }));
          });
          resolve();
        };
        this.pubSocket.onerror = reject;
      });
    }
  }

  /**
   * Ensures the subscriber WebSocket is connected.
   * @private
   * @returns {Promise<void>}
   */
  async ensureSubscriberConnected() {
    if (!this.subSocket || this.subSocket.readyState !== WebSocket.OPEN) {
      this.subSocket = new WebSocket(`${this.baseUrl}/ws/subscribe?channel=${this.channel}`, []);
      await new Promise((resolve, reject) => {
        this.subSocket.onopen = () => {
          Object.entries(this.headers).forEach(([key, value]) => {
            this.subSocket.send(JSON.stringify({ type: 'header', key, value }));
          });
          resolve();
        };
        this.subSocket.onerror = reject;
        this.subSocket.onmessage = (event) => {
          const message = JSON.parse(event.data);
          if (this.listeners[message.topic]) {
            this.listeners[message.topic].forEach(callback => callback(message.topic, message.payload));
          }
        };
      });
    }
  }

  /**
   * Publishes a message to a specific topic.
   * @param {string} topic - The topic to publish to.
   * @param {any} payload - The message payload.
   * @returns {Promise<Object>} - Success response.
   */
  async publish(topic, payload) {
    await this.ensurePublisherConnected();
    this.pubSocket.send(JSON.stringify({ type: 'publish', topic, payload }));
    return { success: true };
  }

  /**
   * Subscribes to one or more topics.
   * @param {string[]} topics - List of topics to subscribe to.
   * @param {function(string, any):void} callback - Callback function when a message is received.
   * @returns {Promise<void>}
   */
  async subscribe(topics, callback) {
    await this.ensureSubscriberConnected();
    topics.forEach(topic => {
      if (!this.listeners[topic]) {
        this.listeners[topic] = [];
      }
      this.listeners[topic].push(callback);
    });
  }

  /**
   * Unsubscribes from a specific topic.
   * @param {string} topic - The topic to unsubscribe from.
   */
  unsubscribe(topic) {
    if (this.listeners[topic]) {
      delete this.listeners[topic];
    }
  }

  /**
   * Disconnects both publisher and subscriber WebSockets.
   */
  disconnect() {
    if (this.pubSocket) {
      this.pubSocket.close();
      this.pubSocket = null;
    }
    if (this.subSocket) {
      this.subSocket.close();
      this.subSocket = null;
    }
  }
}

export default StreamingSDK;
