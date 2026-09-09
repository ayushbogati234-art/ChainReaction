const API = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
 const response = await fetch(API + path, init);
 const body = await response.json();
 if (!response.ok) throw new Error(body.error?.message || "The request failed.");
 return body as T;
}
