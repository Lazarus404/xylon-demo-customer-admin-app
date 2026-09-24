import { useCallback, useState } from "strike-fw/hooks";
import { Btn, Field, Stack, Text } from "strike-fw/ui";
import { AppBar } from "strike-fw-ui/surfaces/app-bar.js";
import { Tabs } from "strike-fw-ui/navigation/tabs.js";
import { Alert } from "strike-fw-ui/feedback/alert.js";
import { Container } from "strike-fw-ui/layout/container.js";
import { Card } from "strike-fw-ui/surfaces/card.js";
import * as api from "./api.js";

const TAB_ITEMS = [
  { id: "live", label: "Live" },
  { id: "history", label: "History" },
  { id: "fleet", label: "Fleet" }
];

const DEFAULT_MODERATE = '{\n  "action": "kick",\n  "participant_id": "p2"\n}';
const DEFAULT_RECORD = '{\n  "mode": "audio"\n}';

function formatResult(data) {
  if (data === null || data === undefined) return "(empty)";
  if (typeof data === "string") return data;
  try {
    return JSON.stringify(data, null, 2);
  } catch {
    return String(data);
  }
}

function fmtTs(sec) {
  if (sec == null) return "-";
  try {
    return new Date(Number(sec) * 1000).toISOString();
  } catch {
    return String(sec);
  }
}

export function App() {
  const [tab, setTab] = useState("live");
  const [status, setStatus] = useState("Idle");
  const [statusTone, setStatusTone] = useState("muted");
  const [sessionId, setSessionId] = useState("room-1");
  const [liveSessions, setLiveSessions] = useState([]);
  const [moderateJson, setModerateJson] = useState(DEFAULT_MODERATE);
  const [recordJson, setRecordJson] = useState(DEFAULT_RECORD);
  const [log, setLog] = useState("");
  const [busy, setBusy] = useState(false);
  const [historyRows, setHistoryRows] = useState([]);
  const [roster, setRoster] = useState([]);
  const [fleetNodes, setFleetNodes] = useState([]);
  const [fleetSummary, setFleetSummary] = useState(null);

  const onStatus = useCallback((text, tone = "info") => {
    setStatus(text);
    setStatusTone(tone);
  }, []);

  function appendLog(kind, text) {
    const line = `[${new Date().toISOString()}] ${kind}: ${text}`;
    setLog((prev) => (prev ? prev + "\n\n" + line : line));
  }

  async function run(label, fn) {
    setBusy(true);
    try {
      const data = await fn();
      appendLog(label, formatResult(data));
      onStatus(`${label} ok`, "ok");
      return data;
    } catch (err) {
      const msg = String(err && err.message ? err.message : err);
      appendLog(`${label} error`, msg);
      onStatus(msg, "danger");
      return null;
    } finally {
      setBusy(false);
    }
  }

  async function refreshLive() {
    const data = await run("sessions", () => api.listLiveSessions());
    const list = (data && data.sessions) || [];
    setLiveSessions(Array.isArray(list) ? list : []);
  }

  async function refreshPresence() {
    const data = await run("presence", () => api.presence(sessionId.trim()));
    setRoster(api.parsePresenceRoster(data));
  }

  async function refreshHistory() {
    const data = await run("history", () => api.historySessions());
    const list = (data && data.sessions) || [];
    setHistoryRows(Array.isArray(list) ? list : []);
  }

  async function refreshFleet(force) {
    const nodesData = await run("fleet nodes", () => api.fleetNodes(!!force));
    const list = (nodesData && nodesData.nodes) || [];
    setFleetNodes(Array.isArray(list) ? list : []);
    const sumData = await run("fleet summary", () => api.fleetSummary(false));
    setFleetSummary((sumData && sumData.summary) || null);
  }

  async function drainNode(node) {
    await run("fleet drain", () => api.fleetAccepting(node, false));
    await refreshFleet(true);
  }

  async function undrainNode(node) {
    await run("fleet undrain", () => api.fleetAccepting(node, true));
    await refreshFleet(true);
  }

  function pickSession(id) {
    setSessionId(id);
    setTab("live");
    setRoster([]);
    onStatus(`Selected ${id}`, "info");
  }

  function fillModerateKick(participantId) {
    setModerateJson(
      JSON.stringify({ action: "kick", participant_id: participantId }, null, 2)
    );
  }

  return (
    <div class="admin-shell">
      <AppBar
        elevated
        position="sticky"
        start={<span class="admin-brand-mark" aria-hidden="true" />}
      >
        <div>
          <p class="admin-brand">Xylon Admin</p>
          <p class="admin-tagline">Demo console | Live | History | Fleet</p>
        </div>
      </AppBar>

      <Container class="admin-main" size="xl">
        <Text as="p" class="admin-hint">
          Browser talks only to this admin process. Live routes proxy through the
          platform interface; History and Fleet read local SQLite (Fleet poller
          refreshes node snapshots).
        </Text>

        <Tabs
          value={tab}
          onChange={setTab}
          items={TAB_ITEMS.map((it) => ({
            id: it.id,
            label: it.label,
            panel: null
          }))}
        />

        <div class="admin-tab-body">
          <div hidden={tab !== "live"}>
            <Card outlined title="Live room" subtitle="Presence, moderate, record">
              <Stack gap={12}>
                <div class="admin-panel-grid admin-panel-grid--2">
                  <Field
                    label="Session id"
                    value={sessionId}
                    onInput={(e) => setSessionId(e.target.value)}
                  />
                </div>
                <Stack row gap={8}>
                  <Btn type="button" disabled={busy} onClick={refreshLive}>
                    Refresh live list
                  </Btn>
                </Stack>
                {liveSessions.length > 0 ? (
                  <div class="admin-live-list">
                    {liveSessions.map((s) => {
                      const id = s.session_id || s.id || String(s);
                      return (
                        <button type="button" key={id} onClick={() => pickSession(id)}>
                          {id}
                        </button>
                      );
                    })}
                  </div>
                ) : null}
                <Stack row gap={8}>
                  <Btn
                    type="button"
                    disabled={busy || !sessionId.trim()}
                    onClick={refreshPresence}
                  >
                    Refresh presence
                  </Btn>
                  <Btn
                    type="button"
                    disabled={busy || !sessionId.trim()}
                    onClick={() =>
                      run("record status", () => api.recordGet(sessionId.trim()))
                    }
                  >
                    Record status
                  </Btn>
                </Stack>
                <div class="admin-roster">
                  <table class="admin-history-table">
                    <thead>
                      <tr>
                        <th>Participant</th>
                        <th>Role</th>
                        <th>Meta</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      {roster.length === 0 ? (
                        <tr>
                          <td colspan="4">No roster yet. Refresh presence for this session.</td>
                        </tr>
                      ) : (
                        roster.map((row) => (
                          <tr key={row.id}>
                            <td>{row.id}</td>
                            <td>{row.role || "-"}</td>
                            <td>{row.meta || "-"}</td>
                            <td>
                              <button type="button" onClick={() => fillModerateKick(row.id)}>
                                Kick
                              </button>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
                <Field
                  label="Moderate JSON"
                  as="textarea"
                  rows={3}
                  value={moderateJson}
                  onInput={(e) => setModerateJson(e.target.value)}
                />
                <Btn
                  type="button"
                  disabled={busy || !sessionId.trim()}
                  onClick={() =>
                    run("moderate", () =>
                      api.moderate(sessionId.trim(), JSON.parse(moderateJson))
                    )
                  }
                >
                  Moderate
                </Btn>
                <Field
                  label="Record start JSON"
                  as="textarea"
                  rows={3}
                  value={recordJson}
                  onInput={(e) => setRecordJson(e.target.value)}
                />
                <Stack row gap={8}>
                  <Btn
                    type="button"
                    disabled={busy || !sessionId.trim()}
                    onClick={() =>
                      run("record start", () =>
                        api.recordStart(sessionId.trim(), JSON.parse(recordJson))
                      )
                    }
                  >
                    Record start
                  </Btn>
                  <Btn
                    type="button"
                    disabled={busy || !sessionId.trim()}
                    onClick={() =>
                      run("record stop", () => api.recordStop(sessionId.trim(), {}))
                    }
                  >
                    Record stop
                  </Btn>
                </Stack>
                <pre class="admin-log">{log || "Results appear here."}</pre>
              </Stack>
            </Card>
          </div>

          <div hidden={tab !== "history"}>
            <Card outlined title="Past sessions" subtitle="SQLite on this admin process">
              <Stack gap={12}>
                <Btn type="button" disabled={busy} onClick={refreshHistory}>
                  Refresh history
                </Btn>
                <div class="admin-history">
                  <table class="admin-history-table">
                    <thead>
                      <tr>
                        <th>Session</th>
                        <th>Created</th>
                        <th>Closed</th>
                        <th>Region</th>
                      </tr>
                    </thead>
                    <tbody>
                      {historyRows.length === 0 ? (
                        <tr>
                          <td colspan="4">No rows yet. Refresh after webhooks arrive.</td>
                        </tr>
                      ) : (
                        historyRows.map((row) => (
                          <tr key={row.session_id}>
                            <td>
                              <button
                                type="button"
                                onClick={() => pickSession(row.session_id)}
                              >
                                {row.session_id}
                              </button>
                            </td>
                            <td>{fmtTs(row.created_at)}</td>
                            <td>{fmtTs(row.closed_at)}</td>
                            <td>{row.last_region || "-"}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </Stack>
            </Card>
          </div>

          <div hidden={tab !== "fleet"}>
            <Card outlined title="Fleet" subtitle="Node snapshots from SQLite (polled via interface)">
              <Stack gap={12}>
                <Stack row gap={8}>
                  <Btn type="button" disabled={busy} onClick={() => refreshFleet(false)}>
                    Refresh cache
                  </Btn>
                  <Btn type="button" disabled={busy} onClick={() => refreshFleet(true)}>
                    Refresh now
                  </Btn>
                </Stack>
                {fleetSummary ? (
                  <Text as="p" class="admin-hint">
                    nodes={fleetSummary.node_count ?? 0} healthy=
                    {fleetSummary.healthy_count ?? 0} unhealthy=
                    {fleetSummary.unhealthy_count ?? 0} sessions=
                    {fleetSummary.session_count ?? 0} participants=
                    {fleetSummary.participant_count ?? 0} regions=
                    {(fleetSummary.regions || []).join(",") || "-"} hub=
                    {fleetSummary.hub_node || "-"}
                  </Text>
                ) : null}
                <div class="admin-history">
                  <table class="admin-history-table">
                    <thead>
                      <tr>
                        <th>Node</th>
                        <th>Region</th>
                        <th>Sessions</th>
                        <th>Participants</th>
                        <th>Hub</th>
                        <th>Cohost hubs</th>
                        <th>Reachable</th>
                        <th>Accepting</th>
                        <th>Polled</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      {fleetNodes.length === 0 ? (
                        <tr>
                          <td colspan="10">
                            No snapshots yet. Refresh now or wait for the poller.
                          </td>
                        </tr>
                      ) : (
                        fleetNodes.map((row) => (
                          <tr key={row.node}>
                            <td>{row.node}</td>
                            <td>{row.region || "-"}</td>
                            <td>{row.session_count ?? 0}</td>
                            <td>{row.participant_count ?? 0}</td>
                            <td>{row.hub ? "yes" : "no"}</td>
                            <td>{row.cohost_hub_count ?? 0}</td>
                            <td>{row.reachable === false ? "no" : "yes"}</td>
                            <td>{row.accepting === false ? "no" : "yes"}</td>
                            <td>{fmtTs(row.polled_at)}</td>
                            <td>
                              {row.accepting === false ? (
                                <Btn
                                  size="sm"
                                  disabled={busy}
                                  onClick={() => undrainNode(row.node)}
                                >
                                  Undrain
                                </Btn>
                              ) : (
                                <Btn
                                  size="sm"
                                  tone="danger"
                                  disabled={busy}
                                  onClick={() => drainNode(row.node)}
                                >
                                  Drain
                                </Btn>
                              )}
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </Stack>
            </Card>
          </div>
        </div>

        <Alert class="admin-status" tone={statusTone} title="Status">
          {status}
        </Alert>
      </Container>
    </div>
  );
}
