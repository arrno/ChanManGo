import { McastClient } from "mcast-client";

const newClient = (): McastClient => {
    return new McastClient({
        authToken: "",
        channel: "MyChan",
        debug: true, // Optional: enables debug logging
    });
};

async function run(): Promise<void> {
    // Initialize the client
    const client = newClient();

    try {
        // Listen for connection state changes
        client.onStateChange((state, connectionType) => {
            console.log(`${connectionType} connection state: ${state}`);
        });

        // Subscribe to topics
        await client.subscribe(
            (topic, message) => {
                console.log(`Received message on topic '${topic}':`, message);
            },
            ["updates", "notifications"]
        );

        // Publish a message
        await client.publish("updates", {
            id: "msg-123",
            text: "Hello world!",
            timestamp: new Date().toISOString(),
        });

        const response = await client.publishHttp("notifications", {
            id: "msg-2",
            text: "Hello via HTTP!",
            timestamp: new Date().toISOString(),
        });
        console.log("HTTP publish response:", response);

        await new Promise((resolve) => {
            setTimeout(() => resolve("Ok"), 3000);
        });
    } catch (err) {
        console.log("Oh no!", err);
    } finally {
        // When done, clean up
        client.disconnect();
    }
}

async function multiProducer(): Promise<void> {
    // Initialize the client
    const client = newClient();

    try {
        // Listen for connection state changes
        client.onStateChange((state, connectionType) => {
            console.log(`${connectionType} connection state: ${state}`);
        });

        // Subscribe to topics
        await client.subscribe(
            (topic, message) => {
                console.log(`Received message on topic '${topic}':`, message);
            },
            ["updates", "notifications"]
        );

        // blows up due to shared state
        // await Promise.all(
        //     Array.from({ length: 10 }, (_, i) =>
        //         client.publish("updates", {
        //             id: `${i}`,
        //             text: "Hello world!",
        //             timestamp: new Date().toISOString(),
        //         })
        //     )
        // );

        // Expected to work
        for (let i = 0; i < 10; i++) {
            await client.publish("updates", {
                id: `${i}`,
                text: "Hello world!",
                timestamp: new Date().toISOString(),
            });
        }

        await new Promise((resolve) => {
            setTimeout(() => resolve("Ok"), 3000);
        });
    } catch (err) {
        console.log("Oh no!", err);
    } finally {
        // When done, clean up
        client.disconnect();
    }
}

async function multiProducerAsync(): Promise<void> {
    let client = newClient();
    let pubClients: Array<McastClient> = [];

    try {
        // Listen for connection state changes
        client.onStateChange((state, connectionType) => {
            console.log(`${connectionType} connection state: ${state}`);
        });

        // Subscribe to topics
        await client.subscribe(
            (topic, message) => {
                console.log(`Received message on topic '${topic}':`, message);
            },
            ["updates", "notifications"]
        );

        try {
            await Promise.all(
                Array.from({ length: 10 }, (_, i) => {
                    let client = newClient();
                    pubClients.push(client);
                    client.publish("updates", {
                        id: `${i}`,
                        text: "Hello world!",
                        timestamp: new Date().toISOString(),
                    });
                })
            );
            console.log("Publish done");
        } catch (err) {
            console.log("Oh no! Publish error:", err);
        }

        await new Promise((resolve) => {
            setTimeout(() => resolve("Ok"), 3000);
        });
    } catch (err) {
        console.log("Oh no!", err);
    } finally {
        // When done, clean up
        client.disconnect();
        pubClients.forEach((c) => c.disconnect());
    }
}

async function multiConsumerAsync(): Promise<void> {
    let subClients: Array<McastClient> = Array.from({ length: 10 }, (_, i) =>
        newClient()
    );

    let client = newClient();
    try {
        await Promise.all(
            subClients.map((client) =>
                client.subscribe(
                    (topic, message) => {
                        console.log(
                            `Received message on topic '${topic}':`,
                            message
                        );
                    },
                    ["updates", "notifications"]
                )
            )
        );
        // Listen for connection state changes
        client.onStateChange((state, connectionType) => {
            console.log(`${connectionType} connection state: ${state}`);
        });

        await client.publish("updates", {
            id: "0",
            text: "Hello world!",
            timestamp: new Date().toISOString(),
        });

        await new Promise((resolve) => {
            setTimeout(() => resolve("Ok"), 3000);
        });
    } catch (err) {
        console.log("Oh no!", err);
    } finally {
        // When done, clean up
        client.disconnect();
        subClients.forEach((client) => client.disconnect());
    }
}

async function multiConsumer(): Promise<void> {
    let client = newClient();
    try {
        // blows up due to shared state
        // await Promise.all(
        //     Array.from({ length: 10 }, (_, i) => {
        //         client.subscribe(
        //             (topic, message) => {
        //                 console.log(
        //                     `Received message on topic '${topic}':`,
        //                     message
        //                 );
        //             },
        //             ["updates", "notifications"]
        //         );
        //     })
        // );

        // works
        for (let i = 0; i < 10; i++) {
            await client.subscribe(
                (topic, message) => {
                    console.log(
                        `Received message on topic '${topic}':`,
                        message
                    );
                },
                ["updates", "notifications"]
            );
        }

        // Listen for connection state changes
        client.onStateChange((state, connectionType) => {
            console.log(`${connectionType} connection state: ${state}`);
        });

        await client.publish("updates", {
            id: "0",
            text: "Hello world!",
            timestamp: new Date().toISOString(),
        });

        await new Promise((resolve) => {
            setTimeout(() => resolve("Ok"), 3000);
        });
    } catch (err) {
        console.log("Oh no!", err);
    } finally {
        // When done, clean up
        client.disconnect();
    }
}

multiConsumer();
