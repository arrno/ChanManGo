import { Client } from "pg";
import * as fs from "fs/promises";
import * as path from "path";

export const sqlFilePath = path.join(__dirname, "schema.sql"); // Path to your .sql file

export async function ping(client: Client): Promise<void> {
    const tableResult = await client.query(
        "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';"
    );
    console.log(`Ping -> ${tableResult.rows.length} tables`);
    tableResult.rows.forEach((row) => {
        console.log(row.table_name);
    });
}

export async function clearDatabase(client: Client): Promise<void> {
    try {
        console.log("Clearing database...");

        // Get all table names
        const tablesResult = await client.query(
            "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';"
        );
        const tableNames = tablesResult.rows.map((row) => row.table_name);
        if (tableNames.length > 0) {
            await client.query(
                `DROP TABLE IF EXISTS ${tableNames.join(", ")} CASCADE;`
            );
            console.log("All tables dropped.");
        } else {
            console.log("No tables to drop.");
        }

        // Get all view names
        const viewsResult = await client.query(
            "SELECT table_name FROM information_schema.views WHERE table_schema = 'public';"
        );
        const viewNames = viewsResult.rows.map((row) => row.table_name);
        if (viewNames.length > 0) {
            await client.query(
                `DROP VIEW IF EXISTS ${viewNames.join(", ")} CASCADE;`
            );
            console.log("All views dropped.");
        } else {
            console.log("No views to drop.");
        }

        // Get all index names (excluding those automatically created by constraints)
        const indexesResult = await client.query(`
        SELECT indexname
        FROM pg_indexes
        WHERE schemaname = 'public'
        AND indexdef NOT LIKE '%UNIQUE CONSTRAINT%'
        AND indexdef NOT LIKE '%PRIMARY KEY%'
      `);
        const indexNames = indexesResult.rows.map((row) => row.indexname);
        if (indexNames.length > 0) {
            await Promise.all(
                indexNames.map(async (indexName) => {
                    await client.query(`DROP INDEX IF EXISTS ${indexName};`);
                })
            );
            console.log("All custom indexes dropped.");
        } else {
            console.log("No custom indexes to drop.");
        }

        console.log("Database cleared successfully.");
    } catch (error) {
        console.error("Error clearing database:", error);
        throw error;
    }
}

export async function bootDatabase(client: Client): Promise<void> {
    try {
        console.log("Booting database...");
        const sqlContent = await fs.readFile(sqlFilePath, "utf8");
        await client.query(sqlContent);
        console.log("Database booted successfully.");
    } catch (error) {
        console.error("Error booting database:", error);
        throw error;
    }
}
