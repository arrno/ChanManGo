/**
 * Configuration options for the ChanbanSDK
 */
interface ChanbanOptions {
    /**
     * The base URL of the server (e.g. "http://localhost:8080" or "https://api.example.com")
     */
    baseUrl: string;
    /**
     * Authentication token for the client
     */
    authToken: string;
    /**
     * The channel to connect to
     */
    channel: string;
    /**
     * Additional headers to include in requests
     */
    headers?: Record<string, string>;
    /**
     * Whether to automatically reconnect on connection loss
     * @default true
     */
    autoReconnect?: boolean;
    /**
     * Maximum reconnection attempts before giving up
     * @default 5
     */
    maxReconnectAttempts?: number;
    /**
     * Delay between reconnection attempts (in milliseconds)
     * @default 3000
     */
    reconnectDelay?: number;
    /**
     * Whether to log debug information to the console
     * @default false
     */
    debug?: boolean;
}
/**
 * Message object for publishing and receiving
 */
interface Message {
    /**
     * Topic of the message
     */
    topic: string;
    /**
     * Payload of the message
     */
    payload: Record<string, any>;
}
/**
 * Serialized message format used internally
 */
interface SerializedMessage {
    /**
     * Topic of the message
     */
    topic: string;
    /**
     * Serialized payload of the message
     */
    payload: string;
}
/**
 * Connection states for WebSocket connections
 */
declare enum ConnectionState {
    DISCONNECTED = "disconnected",
    CONNECTING = "connecting",
    CONNECTED = "connected",
    RECONNECTING = "reconnecting",
    ERROR = "error"
}
/**
 * Response from the server to the client
 */
interface StandardResponse {
    success: boolean;
    message?: string;
    error?: string;
    data?: any;
}
/**
 * Callback type for message subscription
 */
type MessageCallback = (topic: string, message: Record<string, any>) => void;

/**
 * Chanban WebSocket SDK
 *
 * Manages WebSocket connections to a Chanban server for real-time messaging
 */
declare class ChanbanSDK {
    private baseUrl;
    private authToken;
    private channel;
    private headers;
    private autoReconnect;
    private maxReconnectAttempts;
    private reconnectDelay;
    private debug;
    private pubSocket;
    private subSocket;
    private pubState;
    private subState;
    private pubReconnectAttempts;
    private subReconnectAttempts;
    private listeners;
    private stateChangeListeners;
    /**
     * Creates a new instance of the ChanbanSDK
     *
     * @param options Configuration options for the SDK
     */
    constructor(options: ChanbanOptions);
    /**
     * Sets up a WebSocket connection for publishing messages
     *
     * @returns Promise that resolves when the connection is established
     */
    private connectPublisher;
    /**
     * Sets up a WebSocket connection for subscribing to messages
     *
     * @returns Promise that resolves when the connection is established
     */
    private connectSubscriber;
    /**
     * Attempts to reconnect the publisher socket
     */
    private attemptReconnectPublisher;
    /**
     * Attempts to reconnect the subscriber socket
     */
    private attemptReconnectSubscriber;
    /**
     * Updates the publisher connection state and notifies listeners
     *
     * @param state New connection state
     */
    private updatePublisherState;
    /**
     * Updates the subscriber connection state and notifies listeners
     *
     * @param state New connection state
     */
    private updateSubscriberState;
    /**
     * Notifies all state change listeners of a state change
     *
     * @param state New connection state
     * @param connectionType Type of connection that changed state
     */
    private notifyStateChangeListeners;
    /**
     * Publishes a message to a topic
     *
     * @param topic Topic to publish to
     * @param payload Message payload to publish
     * @returns Promise that resolves when the message is published
     */
    publish(topic: string, payload: Record<string, any>): Promise<void>;
    /**
     * Publishes a message via HTTP POST instead of WebSocket
     *
     * @param topic Topic to publish to
     * @param payload Message payload to publish
     * @returns Promise that resolves with the server response
     */
    publishHttp(topic: string, payload: Record<string, any>): Promise<StandardResponse>;
    /**
     * Subscribes to one or more topics
     *
     * @param topics Topic or topics to subscribe to
     * @param callback Callback function to call when a message is received
     * @returns Promise that resolves when the subscription is established
     */
    subscribe(topics: string | string[], callback: MessageCallback): Promise<void>;
    /**
     * Unsubscribes from a topic
     *
     * @param topic Topic to unsubscribe from
     * @param callback Optional callback function to remove. If not provided, all callbacks for the topic will be removed.
     */
    unsubscribe(topic: string, callback?: MessageCallback): void;
    /**
     * Disconnects from the server
     */
    disconnect(): void;
    /**
     * Adds a listener for connection state changes
     *
     * @param listener Function to call when the connection state changes
     * @returns Function to remove the listener
     */
    onStateChange(listener: (state: ConnectionState, connectionType: "publisher" | "subscriber") => void): () => void;
    /**
     * Gets the current connection state for the publisher
     */
    getPublisherState(): ConnectionState;
    /**
     * Gets the current connection state for the subscriber
     */
    getSubscriberState(): ConnectionState;
    /**
     * Checks if a socket is connected
     *
     * @param socket WebSocket to check
     * @returns True if the socket is connected
     */
    private isSocketConnected;
    /**
     * Gets the WebSocket base URL (ws:// or wss://) from the HTTP base URL
     *
     * @returns WebSocket base URL
     */
    private getWebSocketBaseUrl;
    /**
     * Logs a debug message if debug mode is enabled
     *
     * @param message Message to log
     * @param args Additional arguments to log
     */
    private logDebug;
}

/**
 * Load SDK configuration from environment variables
 *
 * @returns SDK options loaded from environment variables
 * @throws Error if required environment variables are missing
 */
declare function loadFromEnv(): ChanbanOptions;

export { type ChanbanOptions, ChanbanSDK, ConnectionState, type Message, type MessageCallback, type SerializedMessage, type StandardResponse, loadFromEnv };
