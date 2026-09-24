/** Thin same-origin fetch helpers for admin :8092 proxies + history. */

async function parseJSON(res) {
  const text = await res.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }
  if (!res.ok) {
    const err = new Error(
      typeof data === "object" && data && data.error
        ? String(data.error)
        : `HTTP ${res.status}`
    );
    err.status = res.status;
    err.body = data;
    throw err;
  }
  return data;
}

/**
 * Normalize Control presence JSON into roster rows.
 * Accepts `{participants: {id: meta}}` or `{participants: [{participant_id|id, ...}]}`.
 */
export function parsePresenceRoster(data) {
  if (!data || typeof data !== "object") return [];
  const raw = data.participants;
  if (!raw) return [];
  if (Array.isArray(raw)) {
    return raw.map((row, i) => {
      if (row == null || typeof row !== "object") {
        return { id: String(row), role: "", meta: "" };
      }
      const id = row.participant_id || row.id || `row-${i}`;
      const role = row.role != null ? String(row.role) : "";
      const { participant_id: _a, id: _b, role: _c, ...rest } = row;
      const meta = Object.keys(rest).length ? JSON.stringify(rest) : "";
      return { id: String(id), role, meta };
    });
  }
  if (typeof raw === "object") {
    return Object.keys(raw)
      .sort()
      .map((id) => {
        const metaObj = raw[id];
        let role = "";
        let meta = "";
        if (metaObj != null && typeof metaObj === "object") {
          if (metaObj.role != null) role = String(metaObj.role);
          const { role: _r, ...rest } = metaObj;
          if (Object.keys(rest).length) meta = JSON.stringify(rest);
        } else if (metaObj != null) {
          meta = String(metaObj);
        }
        return { id, role, meta };
      });
  }
  return [];
}

export async function listLiveSessions() {
  return parseJSON(await fetch("/v1/sessions"));
}

export async function presence(sessionId) {
  return parseJSON(await fetch(`/v1/sessions/${encodeURIComponent(sessionId)}/presence`));
}

export async function moderate(sessionId, body) {
  return parseJSON(
    await fetch(`/v1/sessions/${encodeURIComponent(sessionId)}/moderate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    })
  );
}

export async function recordStart(sessionId, body) {
  return parseJSON(
    await fetch(`/v1/sessions/${encodeURIComponent(sessionId)}/record/start`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body ?? {})
    })
  );
}

export async function recordStop(sessionId, body) {
  return parseJSON(
    await fetch(`/v1/sessions/${encodeURIComponent(sessionId)}/record/stop`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body ?? {})
    })
  );
}

export async function recordGet(sessionId) {
  return parseJSON(await fetch(`/v1/sessions/${encodeURIComponent(sessionId)}/record`));
}

export async function historySessions(limit = 50) {
  return parseJSON(await fetch(`/v1/history/sessions?limit=${limit}`));
}

export async function fleetNodes(refresh = false) {
  const q = refresh ? "?refresh=1" : "";
  return parseJSON(await fetch(`/v1/fleet/nodes${q}`));
}

export async function fleetSummary(refresh = false) {
  const q = refresh ? "?refresh=1" : "";
  return parseJSON(await fetch(`/v1/fleet/summary${q}`));
}

export async function fleetAccepting(node, accepting) {
  return parseJSON(
    await fetch("/v1/fleet/accepting", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ node, accepting })
    })
  );
}
