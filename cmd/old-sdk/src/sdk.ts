export type streamOptions = {
    baseUrl: string;
    authToken: string;
    channel: string;
    headers: { [key: string]: string };
};

export type message = {
    topic: string;
    payload: { [key: string]: any };
};

export type messageSer = {
    topic: string;
    payload: string;
};

export class StreamingSDK {
    baseUrl: string;
    authToken: string;
    channel: string;
    headers: { [key: string]: string };
    pubSocket: null | WebSocket;
    subSocket: null | WebSocket;
    listeners: {
        [key: string]: Array<
            (topic: string, message: Record<string, any>) => void
        >;
    };

    constructor({ baseUrl, authToken, channel, headers = {} }: streamOptions) {
        this.baseUrl = baseUrl;
        this.authToken = authToken;
        this.channel = channel;
        this.headers = { ...headers, Authorization: `Bearer ${authToken}` };
        this.pubSocket = null;
        this.subSocket = null;
        this.listeners = {};
    }

    async ensurePublisherConnected() {
        if (this.pubSocket && this.pubSocket.readyState === WebSocket.OPEN) {
            return;
        }
        const queryParams = new URLSearchParams(this.headers).toString();
        this.pubSocket = new WebSocket(
            `${this.baseUrl}/pub-ws?channel=${this.channel}&${queryParams}`
        );
        await new Promise((resolve, reject) => {
            if (!this.pubSocket) {
                return;
            }
            this.pubSocket.onopen = resolve;
            this.pubSocket.onclose = (event) => {
                // console.warn("Pub WebSocket closed:", event.code, event.reason);
            };
            this.pubSocket.onerror = (event) => {
                reject(event);
            };
        });
    }

    async ensureSubscriberConnected() {
        if (this.subSocket && this.subSocket.readyState == WebSocket.OPEN) {
            return;
        }

        const queryParams = new URLSearchParams(this.headers).toString();
        this.subSocket = new WebSocket(
            `${this.baseUrl}/sub?channel=${this.channel}&${queryParams}`
        );
        await new Promise((resolve, reject) => {
            if (!this.subSocket) {
                reject();
                return;
            }
            this.subSocket.onopen = (event) => {
                console.log("confirmed");
                resolve(event);
            };
            this.subSocket.onerror = (event) => {
                reject(event);
            };
            this.subSocket.onclose = (event) => {
                // console.warn("Sub WebSocket closed:", event.code, event.reason);
            };
            this.subSocket.onmessage = (event) => {
                const message: messageSer = JSON.parse(event.data);
                if (this.listeners[message.topic]) {
                    this.listeners[message.topic].forEach((callback) =>
                        callback(message.topic, JSON.parse(message.payload))
                    );
                }
            };
        });
    }

    async publish(topic: string, payload: { [key: string]: any }) {
        await this.ensurePublisherConnected();
        let message: messageSer = { topic, payload: JSON.stringify(payload) };
        this.pubSocket?.send(JSON.stringify(message));
        return { success: true };
    }

    async subscribe(
        topics: string[],
        callback: (topic: string, message: Object) => void
    ) {
        await this.ensureSubscriberConnected();
        topics.forEach((topic) => {
            if (!this.listeners[topic]) {
                this.listeners[topic] = [];
            }
            this.listeners[topic].push(callback);
        });
    }

    unsubscribe(topic: string) {
        if (this.listeners[topic]) {
            delete this.listeners[topic];
        }
    }

    disconnect() {
        if (this.pubSocket) {
            this.pubSocket.close();
            this.pubSocket = null;
        }
        if (this.subSocket) {
            this.subSocket.close();
            this.pubSocket = null;
        }
    }
}
