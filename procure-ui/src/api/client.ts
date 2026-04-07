const baseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })
  const text = await res.text()
  let data: unknown = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    /* not json */
  }
  if (!res.ok) {
    const errMsg =
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as { error: unknown }).error === 'string'
        ? (data as { error: string }).error
        : text || `${res.status} ${res.statusText}`
    throw new Error(errMsg)
  }
  return data as T
}

export function apiGet<T>(path: string) {
  return request<T>(path, { method: 'GET' })
}

export function apiPost<T>(path: string, body: unknown) {
  return request<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function getApiBaseUrl() {
  return baseUrl
}
