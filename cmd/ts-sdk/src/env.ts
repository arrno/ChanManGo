import { ChanbanOptions } from "./types";

/**
 * Load SDK configuration from environment variables
 *
 * @returns SDK options loaded from environment variables
 * @throws Error if required environment variables are missing
 */
export function loadFromEnv(): ChanbanOptions {
    // Required configuration
    const authToken = process.env.CHANBAN_AUTH_TOKEN;
    const channel = process.env.CHANBAN_CHANNEL;

    if (!authToken) {
        throw new Error("CHANBAN_AUTH_TOKEN environment variable is required");
    }

    if (!channel) {
        throw new Error("CHANBAN_CHANNEL environment variable is required");
    }

    // Optional configuration
    const debug = process.env.CHANBAN_DEBUG === "true";
    const autoReconnect = process.env.CHANBAN_AUTO_RECONNECT !== "false"; // Default to true
    const maxReconnectAttempts = parseInt(
        process.env.CHANBAN_MAX_RECONNECT_ATTEMPTS || "5",
        10
    );
    const reconnectDelay = parseInt(
        process.env.CHANBAN_RECONNECT_DELAY || "3000",
        10
    );

    return {
        authToken,
        channel,
        debug,
        autoReconnect,
        maxReconnectAttempts,
        reconnectDelay,
    };
}
