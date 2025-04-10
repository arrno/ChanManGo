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
        authToken: "OBZx5SEpFIAZfJfsHyDG::EGmUVXliEipRW50FOWvp",
        // topics: ["updates", "notifications", TOPIC_ALL],
        channel: "MyChan",
        debug: true,
    });

    let subs: Array<ChanbanSDK> = [];
    for (let i = 0; i < 10; i++) {
        subs.push(
            new ChanbanSDK({
                baseUrl: "http://localhost:8080",
                authToken: "OBZx5SEpFIAZfJfsHyDG::EGmUVXliEipRW50FOWvp",
                // topics: ["updates", "notifications", TOPIC_ALL],
                channel: "MyChan",
                debug: true,
            })
        );
        subs[subs.length - 1]?.subscribe(
            (topic, message) => {
                console.log(
                    `Received message on ${i} topic '${topic}':`,
                    message
                );
            }
            // ["woo"]
        );
    }

    const sdkTwo = new ChanbanSDK({
        baseUrl: "http://localhost:8080",
        authToken: "OBZx5SEpFIAZfJfsHyDG::EGmUVXliEipRW50FOWvp",
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

        console.log("Publishing message...");
        await sdkTwo.publish("alt", {
            id: "msg-1",
            text: "Hello from Chanban SDK!",
            timestamp: new Date().toISOString(),
        });

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
        subs.forEach((sub) => sub.disconnect());
        console.log("Disconnected");
    }
}
