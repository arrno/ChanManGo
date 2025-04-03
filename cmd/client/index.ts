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
// Example usage:
async function example() {
    try {
        const getResult = await httpGet<Record<string, any>>(
            `${url}/clients?page=1&count=10`,
            "password"
        );
        console.log("GET result:", getResult);

        // const postData = { title: "foo", body: "bar", userId: 1 };
        // const postResult = await httpPost(
        //     "https://jsonplaceholder.typicode.com/posts",
        //     postData
        // );
        // console.log("POST result:", postResult);
    } catch (error) {
        console.error("OOPS! ->:", error);
    }
}

example();
