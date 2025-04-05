import { ChanbanSDK, loadFromEnv, ConnectionState } from "chanban-sdk";

export async function sdk() {
    // Initialize the SDK with your credentials
    const sdk = new ChanbanSDK({
        baseUrl: "http://localhost:8080",
        authToken: "admin::password",
        channel: "main",
        debug: true,
    });

    // Listen for connection state changes
    const unsubscribe = sdk.onStateChange((state, connectionType) => {
        console.log(`${connectionType} connection state changed to: ${state}`);
    });

    try {
        // Subscribe to multiple topics
        console.log("Subscribing to topics...");
        await sdk.subscribe(["updates", "notifications"], (topic, message) => {
            console.log(`Received message on topic '${topic}':`, message);
        });

        // Publish a message
        console.log("Publishing message...");
        await sdk.publish("updates", {
            id: "msg-1",
            text: "Hello from Chanban SDK!",
            timestamp: new Date().toISOString(),
        });

        // Keep the connection open for a while to receive messages
        console.log("Waiting for messages...");
        await new Promise((resolve) => setTimeout(resolve, 3000));

        // Unsubscribe from state change events
        unsubscribe();

        // Unsubscribe from a topic
        sdk.unsubscribe("notifications");

        // Publish using HTTP fallback
        console.log("Publishing message via HTTP...");
        const response = await sdk.publishHttp("updates", {
            id: "msg-2",
            text: "Hello via HTTP!",
            timestamp: new Date().toISOString(),
        });
        console.log("HTTP publish response:", response);
    } catch (error) {
        console.error("Error:", error);
    } finally {
        // Clean up
        sdk.disconnect();
        console.log("Disconnected");
    }
}
