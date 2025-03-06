import { StreamingSDK, streamOptions } from "./sdk.js";

async function fun() {
    const client = new StreamingSDK(<streamOptions>{
        baseUrl: "ws://localhost:8080",
        authToken: "auth-token",
        channel: "MyChannel",
    });

    client
        .subscribe(["topicOne", "notifications"], (topic, message) => {
            console.log(`Received message for topic ${topic}. `, message);
        })
        .catch((err) => {
            console.log(err);
        });

    try {
        const response = await client.publish("topicOne", `{ Hello: "World" }`);
        console.log("Message publish response: ", response);
    } catch (err) {
        console.log("Message publish failed: ", err);
    } finally {
        new Promise((resolve) => {
            setTimeout(resolve, 1000);
        }).then(
            () => client.disconnect() // Ensure cleanup
        );
    }
}

fun();

// const test = async () => {
//     await new Promise((resolve) => setTimeout(resolve, 1000));
//     console.log("Late");
// };

// test();
