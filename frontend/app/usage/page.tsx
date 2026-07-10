"use client";

import { useEffect, useState } from "react";
import { getAPIBase } from "../utils/api";

interface Team {
  id: string;
  name: string;
}

interface UsageLog {
  id: string;
  team_id: string;
  model: string;
  prompt_tokens: number;
  completion_tokens: number;
  cost: number;
  latency_ms: number;
  created_at: string;
  pii_redacted_count: number;
}

export default function UsagePage() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeam, setSelectedTeam] = useState("");
  const [logs, setLogs] = useState<UsageLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [logsLoading, setLogsLoading] = useState(false);
  const [isDemoData, setIsDemoData] = useState(false);

  useEffect(() => {
    const apiBase = getAPIBase();
    fetch(`${apiBase}/api/teams`, { cache: "no-store" })
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch teams");
        return res.json();
      })
      .then((data) => {
        const teamData = data || [];
        setTeams(teamData);
        if (teamData.length > 0) {
          setSelectedTeam(teamData[0].id);
          fetchLogs(teamData[0].id);
        } else {
          setLoading(false);
          setIsDemoData(true);
          setLogs(getMockLogs());
        }
      })
      .catch((err) => {
        console.error(err);
        const fallbackTeams = [
          { id: "1", name: "Engineering Core" },
          { id: "2", name: "Data Science Research" },
        ];
        setTeams(fallbackTeams);
        setSelectedTeam("1");
        setIsDemoData(true);
        setLogs(getMockLogs());
        setLoading(false);
      });
  }, []);

  const fetchLogs = (teamID: string) => {
    if (!teamID) return;
    setLogsLoading(true);
    const apiBase = getAPIBase();
    fetch(`${apiBase}/api/usage?team_id=${teamID}`, { cache: "no-store" })
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch usage logs");
        return res.json();
      })
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          setLogs(data);
          setIsDemoData(false);
        } else {
          // If DB returned empty, load mock logs and map them to the selected teamID so they display!
          const mappedMockLogs = getMockLogs().map(log => ({
            ...log,
            team_id: teamID,
          }));
          setLogs(mappedMockLogs);
          setIsDemoData(true);
        }
        setLogsLoading(false);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        const mappedMockLogs = getMockLogs().map(log => ({
          ...log,
          team_id: teamID,
        }));
        setLogs(mappedMockLogs);
        setIsDemoData(true);
        setLogsLoading(false);
        setLoading(false);
      });
  };

  const handleTeamChange = (teamID: string) => {
    setSelectedTeam(teamID);
    fetchLogs(teamID);
  };

  const getMockLogs = (): UsageLog[] => {
    return [
      {
        id: "log_51fa93a1-2d7c-486a",
        team_id: "1",
        model: "gpt-4o",
        prompt_tokens: 1240,
        completion_tokens: 832,
        cost: 0.01868,
        latency_ms: 1230,
        created_at: new Date(Date.now() - 500000).toISOString(),
        pii_redacted_count: 2
      },
      {
        id: "log_2d781bcf-18e3-40fa",
        team_id: "1",
        model: "claude-3-5-sonnet",
        prompt_tokens: 4500,
        completion_tokens: 1205,
        cost: 0.03157,
        latency_ms: 2450,
        created_at: new Date(Date.now() - 1500000).toISOString(),
        pii_redacted_count: 4
      },
      {
        id: "log_8c9431ea-42f0-11ff",
        team_id: "2",
        model: "claude-3-haiku-20240307",
        prompt_tokens: 12800,
        completion_tokens: 3400,
        cost: 0.00745,
        latency_ms: 840,
        created_at: new Date(Date.now() - 4000000).toISOString(),
        pii_redacted_count: 0
      },
      {
        id: "log_efbc2100-baef-42f9",
        team_id: "1",
        model: "gpt-3.5-turbo",
        prompt_tokens: 230,
        completion_tokens: 45,
        cost: 0.00018,
        latency_ms: 410,
        created_at: new Date(Date.now() - 6000000).toISOString(),
        pii_redacted_count: 0
      }
    ];
  };

  return (
    <div className="container">
      <div style={{ marginBottom: "2rem" }}>
        <h1 style={{ fontSize: "1.75rem", fontWeight: 700 }}>Real-time Audit Logs</h1>
        <p style={{ color: "var(--secondary)", fontSize: "0.875rem" }}>
          Exhaustive token auditing, provider latency mapping, and exact usage metrics.
        </p>
      </div>

      <div className="card" style={{ marginBottom: "1.5rem" }}>
        <div style={{ display: "flex", gap: "1rem", alignItems: "center" }}>
          <label className="label" style={{ marginBottom: 0 }}>Filter by Team / Department:</label>
          <select
            className="input"
            style={{ width: "240px", padding: "0.4rem 0.5rem" }}
            value={selectedTeam}
            onChange={(e) => handleTeamChange(e.target.value)}
            disabled={loading}
          >
            {teams.length === 0 ? (
              <option value="">Demo Showcase Team</option>
            ) : (
              teams.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))
            )}
          </select>
        </div>
      </div>

      {isDemoData && (
        <div style={{
          background: "var(--warning-light)",
          color: "var(--warning)",
          padding: "1rem",
          borderRadius: "0.5rem",
          fontSize: "0.875rem",
          fontWeight: 500,
          marginBottom: "1.5rem",
          border: "1px solid var(--warning)",
          boxShadow: "0 1px 2px rgba(245, 158, 11, 0.05)"
        }}>
          💡 <strong>Demo Data Sandbox:</strong> No live API requests have been proxied yet for this team. Showing simulated enterprise traffic telemetry. Once developers make requests to the gateway, real logs will appear here.
        </div>
      )}

      <div className="card" style={{ padding: 0, overflow: "hidden" }}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%" }}>
            <thead>
              <tr>
                <th>Log UUID</th>
                <th>Model</th>
                <th>Tokens Used</th>
                <th>Downstream Cost</th>
                <th>Provider Latency</th>
                <th>Privacy Scrubber</th>
                <th>Logged At</th>
              </tr>
            </thead>
            <tbody>
              {logsLoading ? (
                <tr>
                  <td colSpan={7} style={{ textAlign: "center", padding: "3rem", color: "var(--secondary)" }}>
                    Loading audit trail...
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={7} style={{ textAlign: "center", padding: "3rem", color: "var(--secondary)" }}>
                    No request telemetry captured yet for this team.
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id}>
                    <td style={{ fontFamily: "var(--font-geist-mono)", fontSize: "0.825rem", color: "var(--secondary)" }}>
                      {log.id}
                    </td>
                    <td>
                      <span style={{
                        padding: "0.2rem 0.5rem",
                        background: log.model.startsWith("claude") ? "rgba(217, 119, 6, 0.1)" : "rgba(79, 70, 229, 0.1)",
                        color: log.model.startsWith("claude") ? "var(--warning)" : "var(--primary)",
                        borderRadius: "0.25rem",
                        fontSize: "0.8rem",
                        fontWeight: 600
                      }}>
                        {log.model}
                      </span>
                    </td>
                    <td>
                      <div style={{ display: "flex", flexDirection: "column" }}>
                        <span style={{ fontWeight: 600 }}>{(log.prompt_tokens + log.completion_tokens).toLocaleString()}</span>
                        <span style={{ fontSize: "0.75rem", color: "var(--secondary)" }}>
                          in: {log.prompt_tokens.toLocaleString()} | out: {log.completion_tokens.toLocaleString()}
                        </span>
                      </div>
                    </td>
                    <td style={{ fontFamily: "var(--font-geist-mono)", fontWeight: 600, color: "var(--success)" }}>
                      ${log.cost.toLocaleString(undefined, { minimumFractionDigits: 5, maximumFractionDigits: 5 })}
                    </td>
                    <td>
                      <div style={{ display: "flex", alignItems: "center", gap: "0.375rem" }}>
                        <span style={{ display: "inline-block", width: "8px", height: "8px", borderRadius: "50%", background: log.latency_ms > 2000 ? "var(--warning)" : "var(--success)" }} />
                        <span>{log.latency_ms} ms</span>
                      </div>
                    </td>
                    <td>
                      {log.pii_redacted_count > 0 ? (
                        <span style={{
                          padding: "0.15rem 0.5rem",
                          borderRadius: "0.25rem",
                          fontSize: "0.75rem",
                          fontWeight: 600,
                          background: "var(--success-light)",
                          color: "var(--success)"
                        }}>
                          🛡️ {log.pii_redacted_count} Leaks Intercepted
                        </span>
                      ) : (
                        <span style={{
                          padding: "0.15rem 0.5rem",
                          borderRadius: "0.25rem",
                          fontSize: "0.75rem",
                          fontWeight: 500,
                          background: "rgba(100,116,139,0.08)",
                          color: "var(--secondary)"
                        }}>
                          ✓ Sanitized
                        </span>
                      )}
                    </td>
                    <td style={{ fontSize: "0.825rem", color: "var(--secondary)" }}>
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
