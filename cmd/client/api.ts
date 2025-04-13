async function httpDelete<T>(url: string, token: string): Promise<T> {
    const response = await fetch(url, {
        method: "DELETE",
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    try {
        return (await response.json()) as T;
    } catch (error) {
        throw new Error("Failed to parse JSON");
    }
}

async function httpGet<T>(url: string, token: string): Promise<T> {
    const response = await fetch(url, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    try {
        return (await response.json()) as T;
    } catch (error) {
        throw new Error("Failed to parse JSON");
    }
}

async function httpPost<T>(
    url: string,
    token: string,
    data: Record<string, any>
): Promise<T> {
    const response = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(data),
    });
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    try {
        return (await response.json()) as T;
    } catch (error) {
        throw new Error("Failed to parse JSON");
    }
}

const url = "http://localhost:8080";
const token = "Xp4Ve4azTepGQ1qE3LRo::INwOx6XCoqTSzCbK6T4Z";

// 6B2tz2fpuOSqiZSXiqxV
// Example usage:
export async function api() {
    try {
        // const result = await httpPost<Record<string, any>>(
        //     `${url}/client`,
        //     token,
        //     {
        //         clientName: "CoolKid",
        //     }
        // );
        // console.log("POST result:", result);

        // const getResult = await httpGet<Record<string, any>>(
        //     `${url}/clients?page=1&count=10&archived=true`,
        //     token
        // );
        // console.log("GET result:", getResult);

        // const getResult = await httpGet<Record<string, any>>(
        //     `${url}/client?clientId=5j3bNZ7EjN0hjbu7nn3P`,
        //     token
        // );
        // console.log("GET result:", getResult);

        // const getResult = await httpDelete<Record<string, any>>(
        //     `${url}/client?clientId=5j3bNZ7EjN0hjbu7nn3P`,
        //     token
        // );
        // console.log("DELETE result:", getResult);

        // const result = await httpPost<Record<string, any>>(
        //     `${url}/rotate-token`,
        //     token,
        //     {}
        // );
        // console.log("POST result:", result);

        const result = await httpPost<Record<string, any>>(
            `${url}/beta/signup`,
            "",
            {}
        );
        console.log("POST result:", result);
    } catch (error) {
        console.error("OOPS! ->:", error);
    }
}
