import { ChanbanSDK, loadFromEnv } from './index';

// Try to load configuration from environment variables
async function main() {
  try {
    // Load configuration from environment variables
    const config = loadFromEnv();
    console.log('Configuration loaded from environment:', config);

    // Initialize the SDK with the loaded configuration
    const sdk = new ChanbanSDK(config);

    // Subscribe to a topic
    await sdk.subscribe('updates', (topic, message) => {
      console.log(`Received message on topic '${topic}':`, message);
    });

    // Publish a message
    await sdk.publish('updates', {
      id: `msg-${Date.now()}`,
      text: 'This message was sent using environment-based configuration',
      timestamp: new Date().toISOString()
    });

    // Keep the connection open for a while
    console.log('Waiting for messages...');
    await new Promise(resolve => setTimeout(resolve, 10000));

    // Clean up
    sdk.disconnect();
    console.log('Disconnected');

  } catch (error) {
    console.error('Error:', error);
  }
}

// Run the example
main().catch(console.error);