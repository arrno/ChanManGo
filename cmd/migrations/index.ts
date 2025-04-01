import { Client } from "pg";
import * as fs from "fs/promises";
import * as path from "path";
import { clearDatabase, bootDatabase, ping } from "./util";

// --- Configuration ---
export const dbConfig = {
    user: "myuser",
    host: "localhost",
    database: "postgres",
    password: "password",
    port: 5433, // Default PostgreSQL port
};

async function main() {
    const client = new Client(dbConfig);

    try {
        await client.connect();
        await ping(client);
    } catch (error) {
        console.error("An error occurred:", error);
    } finally {
        console.log("Closing client");
        await client.end();
    }

    // try {
    //     await client.connect();
    //     await clearDatabase(client);
    //     await bootDatabase(client);
    // } catch (error) {
    //     console.error("An error occurred:", error);
    // } finally {
    //     await client.end();
    // }
}

main().catch(console.error);
