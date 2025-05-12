import {
    McastClient,
    loadFromEnv,
    ConnectionState,
    TOPIC_ALL,
} from "mcast-client";

export async function sdk() {
    // Initialize the SDK with your credentials
    const client = new McastClient({
        authToken: "",
        // topics: ["updates", "notifications", TOPIC_ALL],
        channel: "MyChan",
        debug: true,
    });

    // Listen for connection state changes
    const unsubscribe = client.onStateChange((state, connectionType) => {
        console.log(`${connectionType} connection state changed to: ${state}`);
    });

    try {
        // Subscribe to multiple topics
        console.log("Subscribing to topics...");
        await client.subscribe(
            (topic, message) => {
                console.log(`Received message on topic '${topic}':`, message);
            }
            // ["woo"]
        );

        for (let i = 0; i < 20; i++) {
            console.log("Publishing message...");
            await client.publish("alt", {
                id: "msg-1",
                text: "Hello from Chanban SDK!",
                timestamp: new Date().toISOString(),
            });

            // Publish using HTTP fallback
            console.log("Publishing message via HTTP...");
            const response = await client.publishHttp("alt", {
                id: "msg-2",
                text: "Hello via HTTP!",
                timestamp: new Date().toISOString(),
            });
            console.log("HTTP publish response:", response);

            console.log(
                `Sleeping for 60 seconds: ${new Date().toLocaleString()}`
            );
            await new Promise((resolve) => setTimeout(resolve, 60000)); // one minute
        }
    } catch (error) {
        console.error("Error:", error);
    } finally {
        // Clean up
        client.disconnect();
        console.log("Disconnected");
    }
}
