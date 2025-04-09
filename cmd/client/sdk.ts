import {
    ChanbanSDK,
    loadFromEnv,
    ConnectionState,
    TOPIC_ALL,
} from "chanban-sdk";

export async function sdk() {
    // Initialize the SDK with your credentials
    const sdk = new ChanbanSDK({
        baseUrl: "http://localhost:8080",
        authToken: "FCEDiSm67vuLzUVpvVQg::ZVlDVoEa7tnneRg5u0qw",
        // topics: ["updates", "notifications", TOPIC_ALL],
        channel: "MyChan",
        debug: true,
    });

    const sdkTwo = new ChanbanSDK({
        baseUrl: "http://localhost:8080",
        authToken: "FCEDiSm67vuLzUVpvVQg::ZVlDVoEa7tnneRg5u0qw",
        // topics: ["alt", TOPIC_ALL],
        channel: "MyChan",
        debug: true,
    });

    // Listen for connection state changes
    const unsubscribe = sdk.onStateChange((state, connectionType) => {
        console.log(`${connectionType} connection state changed to: ${state}`);
    });

    try {
        // Subscribe to multiple topics
        console.log("Subscribing to topics...");
        await sdk.subscribe(
            (topic, message) => {
                console.log(`Received message on topic '${topic}':`, message);
            }
            // ["woo"]
        );

        for (let i = 0; i < 20; i++) {
            // Publish a message
            console.log("Publishing message...");
            await sdkTwo.publish("alt", {
                id: "msg-1",
                text: "Hello from Chanban SDK!",
                timestamp: new Date().toISOString(),
            });
        }

        // Keep the connection open for a while to receive messages
        console.log("Waiting for messages...");

        // Unsubscribe from state change events
        unsubscribe();

        // Unsubscribe from a topic
        sdk.unsubscribe("notifications");

        // Publish using HTTP fallback
        console.log("Publishing message via HTTP...");
        const response = await sdkTwo.publishHttp("alt", {
            id: "msg-2",
            text: "Hello via HTTP!",
            timestamp: new Date().toISOString(),
        });
        console.log("HTTP publish response:", response);

        await new Promise((resolve) => setTimeout(resolve, 3000));
    } catch (error) {
        console.error("Error:", error);
    } finally {
        // Clean up
        sdk.disconnect();
        sdkTwo.disconnect();
        console.log("Disconnected");
    }
}
