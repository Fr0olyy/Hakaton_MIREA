import { NextRequest } from "next/server";

const BACKEND_URL = (process.env.BACKEND_INTERNAL_URL || "http://localhost:8080").replace(/\/$/, "");
const DEMO_EMAIL = process.env.DEMO_AUTH_EMAIL || "demo@dataforge.local";
const DEMO_PASSWORD = process.env.DEMO_AUTH_PASSWORD || "dataforge-demo";

let cachedToken = "";
let cachedTokenAt = 0;

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

function normalizeBackendPath(path: string[]) {
  const rawPath = path.join("/").replace(/^\/+/, "");

  // Allows both frontend variants:
  // /api/backend/projects/:id
  // /api/backend/api/projects/:id
  //
  // Backend expects:
  // /api/projects/:id
  if (rawPath === "api" || rawPath.startsWith("api/")) {
    return rawPath;
  }

  return `api/${rawPath}`;
}

async function proxy(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;

  const backendPath = normalizeBackendPath(path);
  const upstreamURL = new URL(`${BACKEND_URL}/${backendPath}`);
  upstreamURL.search = request.nextUrl.search;

  console.log("PROXY TO:", upstreamURL.toString());

  const headers = new Headers(request.headers);
  headers.delete("host");
  headers.delete("connection");
  headers.delete("content-length");

  const isAuthRoute = backendPath.startsWith("api/auth");

  if (!headers.has("authorization") && !isAuthRoute) {
    const token = await getDemoToken();
    if (token) {
      headers.set("authorization", `Bearer ${token}`);
    }
  }

  const hasBody = request.method !== "GET" && request.method !== "HEAD";

  const upstream = await fetch(upstreamURL, {
    method: request.method,
    headers,
    body: hasBody ? await request.arrayBuffer() : undefined,
    cache: "no-store",
  });

  const responseHeaders = new Headers(upstream.headers);
  responseHeaders.delete("content-encoding");
  responseHeaders.delete("content-length");
  responseHeaders.delete("transfer-encoding");

  return new Response(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: responseHeaders,
  });
}

export async function GET(request: NextRequest, context: RouteContext) {
  return proxy(request, context);
}

export async function POST(request: NextRequest, context: RouteContext) {
  return proxy(request, context);
}

export async function PUT(request: NextRequest, context: RouteContext) {
  return proxy(request, context);
}

export async function PATCH(request: NextRequest, context: RouteContext) {
  return proxy(request, context);
}

export async function DELETE(request: NextRequest, context: RouteContext) {
  return proxy(request, context);
}

async function getDemoToken() {
  if (cachedToken && Date.now() - cachedTokenAt < 20 * 60 * 60 * 1000) {
    return cachedToken;
  }

  const login = await authRequest("/api/auth/login", {
    email: DEMO_EMAIL,
    password: DEMO_PASSWORD,
  });

  if (login?.token) {
    cachedToken = login.token;
    cachedTokenAt = Date.now();
    return cachedToken;
  }

  const registered = await authRequest("/api/auth/register", {
    email: DEMO_EMAIL,
    password: DEMO_PASSWORD,
    name: "DataForge Demo",
  });

  if (registered?.token) {
    cachedToken = registered.token;
    cachedTokenAt = Date.now();
    return cachedToken;
  }

  const retry = await authRequest("/api/auth/login", {
    email: DEMO_EMAIL,
    password: DEMO_PASSWORD,
  });

  cachedToken = retry?.token || "";
  cachedTokenAt = cachedToken ? Date.now() : 0;

  return cachedToken;
}

async function authRequest(path: string, payload: Record<string, string>) {
  try {
    const response = await fetch(`${BACKEND_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as { token?: string };
  } catch {
    return null;
  }
}