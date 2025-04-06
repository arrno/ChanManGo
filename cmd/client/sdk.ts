import { ChanbanSDK, loadFromEnv, ConnectionState } from "chanban-sdk";

export async function sdk() {
    // Initialize the SDK with your credentials
    const pubSdk = new ChanbanSDK({
        baseUrl: "http://localhost:8080",
        authToken: "7vOgmp2yE3keVfF2bQXP::nR8vZ8WCGylzTH2LS9M1",
        channel: "MyChan",
        debug: true,
    });

    const subSdk = new ChanbanSDK({
        baseUrl: "http://localhost:8080",
        authToken: "g675bjNgR5VygTPSfiSl::4nEA9gYtgUwry9yHqDPN",
        channel: "MyChan",
        debug: true,
    });

    // Listen for connection state changes
    const unsubscribePub = pubSdk.onStateChange((state, connectionType) => {
        console.log(
            `${connectionType} PUB connection state changed to: ${state}`
        );
    });

    const unsubscribeSub = subSdk.onStateChange((state, connectionType) => {
        console.log(
            `${connectionType} SUB connection state changed to: ${state}`
        );
    });

    try {
        // Subscribe to multiple topics
        console.log("Subscribing to topics...");
        await subSdk.subscribe(
            ["updates", "notifications"],
            (topic, message) => {
                console.log(`Received message on topic '${topic}':`, message);
            }
        );

        // Publish a message
        console.log("Publishing message...");
        await pubSdk.publish("updates", {
            id: "msg-1",
            text: "Hello from Chanban SDK!",
            timestamp: new Date().toISOString(),
        });

        // Keep the connection open for a while to receive messages
        console.log("Waiting for messages...");

        // Unsubscribe from state change events
        unsubscribePub();
        unsubscribeSub();

        // Unsubscribe from a topic
        subSdk.unsubscribe("notifications");

        // Publish using HTTP fallback
        console.log("Publishing message via HTTP...");
        const response = await pubSdk.publishHttp("updates", {
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
        subSdk.disconnect();
        pubSdk.disconnect();
        console.log("Disconnected");
    }
}
