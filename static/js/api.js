// @ts-check

/**
 * @template T
 * @typedef {{ ok: boolean, status: number, data: T }} JSONResult
 */

/**
 * @typedef {{ timeoutMs?: number }} TimeoutOpts
 */

/**
 * 簡易fetch(JSON専用) with timeout
 * @template T
 * @param {string} url
 * @param {RequestInit} [init]
 * @param {TimeoutOpts} [opts]
 * @returns {Promise<JSONResult<T>>}
 */
export async function fetchJSON(url, init = {}, opts = {}) {
  const { timeoutMs = 10000 } = opts || {};
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), timeoutMs);
  try {
    const res = await fetch(url, { ...init, signal: ctrl.signal });
    /** @type {T} */
    const data = await res.json().catch(() => (/** @type {T} */(/** @type {unknown} */({}))));
    return { ok: res.ok, status: res.status, data };
  } finally { clearTimeout(t); }
}

// Monitoring APIs
/** @returns {Promise<JSONResult<unknown>>} */
export async function getActive() {
  return fetchJSON('/api/monitoring/active', { cache: 'no-store' });
}

/**
 * @param {string} videoId
 * @param {number=} maxUsers
 * @returns {Promise<JSONResult<unknown>>}
 */
export async function startMonitoring(videoId, maxUsers) {
  const body = { video_input: String(videoId) };
  if (typeof maxUsers === 'number') body.max_users = maxUsers;
  return fetchJSON('/api/monitoring/start', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
  });
}

/**
 * @param {string} [videoId]
 * @returns {Promise<JSONResult<unknown>>}
 */
export async function resumeMonitoring(videoId) {
  const body = videoId ? { video_input: String(videoId) } : {};
  return fetchJSON('/api/monitoring/resume', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
  });
}

/** @returns {Promise<JSONResult<unknown>>} */
export async function stopMonitoring() {
  return fetchJSON('/api/monitoring/stop', { method: 'DELETE' });
}

/** @param {string} videoId */
export async function getUsers(videoId) {
  return fetchJSON(`/api/monitoring/${encodeURIComponent(videoId)}/users`, { cache: 'no-store' });
}

/** @param {string} videoId */
export async function getStatus(videoId) {
  return fetchJSON(`/api/monitoring/${encodeURIComponent(videoId)}/status`, { cache: 'no-store' });
}

// Logs APIs
/** @param {Record<string,string|number|undefined>} params */
export async function getLogs(params) {
  const sp = new URLSearchParams(params || {});
  return fetchJSON('/api/logs' + (sp.toString() ? ('?' + sp.toString()) : ''));
}

/** @param {Record<string,string|number|undefined>} params */
export async function getLogsStats(params) {
  const p = { ...(params || {}), stats: 1 };
  const sp = new URLSearchParams(p);
  return fetchJSON('/api/logs?' + sp.toString());
}

/** @returns {Promise<JSONResult<unknown>>} */
export async function clearLogs() {
  return fetchJSON('/api/logs', { method: 'DELETE' });
}
