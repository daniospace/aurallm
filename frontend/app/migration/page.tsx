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
  is_shadow: boolean;
  primary_log_id: string;
}

interface MigrationRun {
  primary: UsageLog;
  shadow: UsageLog;
}

export default function MigrationAuditorPage() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeam, setSelectedTeam] = useState("");
  const [runs, setRuns] = useState<MigrationRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [runsLoading, setRunsLoading] = useState(false);
  const [isDemoData, setIsDemoData] = useState(false);

  useEffect(() => {
    const apiBase = getAPIBase();
    fetch(`${apiBase}/api/teams`, { cache: "no-store" })
      .then((res) => {
        if (!res.ok) {
          console.warn("Could not connect to API server. Falling back to sandbox.");
          const fallbackTeams = [
            { id: "1", name: "Engineering Core" },
            { id: "2", name: "Data Science Research" },
          ];
          setTeams(fallbackTeams);
          setSelectedTeam("1");
          setRuns(getMockRuns());
          setIsDemoData(true);
          setLoading(false);
          return null;
        }
        return res.json();
      })
      .then((data) => {
        if (!data) return;
        const teamData = data || [];
        setTeams(teamData);
        if (teamData.length > 0) {
          setSelectedTeam(teamData[0].id);
          fetchMigrationRuns(teamData[0].id);
        } else {
          setLoading(false);
          setRuns(getMockRuns());
          setIsDemoData(true);
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
        setRuns(getMockRuns());
        setIsDemoData(true);
        setLoading(false);
      });
  }, []);

  const fetchMigrationRuns = (teamID: string) => {
    if (!teamID) return;
    setRunsLoading(true);

    const apiBase = getAPIBase();
    fetch(`${apiBase}/api/usage?team_id=${teamID}`, { cache: "no-store" })
      .then((res) => {
        if (!res.ok) {
          console.warn("Could not fetch usage logs. Loading simulated runs.");
          setRuns(getMockRuns().map(run => ({
            primary: { ...run.primary, team_id: teamID },
            shadow: { ...run.shadow, team_id: teamID },
          })));
          setIsDemoData(true);
          setRunsLoading(false);
          setLoading(false);
          return null;
        }
        return res.json();
      })
      .then((data) => {
        if (!data) return;
        if (Array.isArray(data) && data.length > 0) {
          const primaryLogs = data.filter((l: UsageLog) => !l.is_shadow);
          const shadowLogs = data.filter((l: UsageLog) => l.is_shadow);

          const matchedRuns: MigrationRun[] = [];
          primaryLogs.forEach((p) => {
            const linked = shadowLogs.find((s) => s.primary_log_id === p.id);
            if (linked) {
              matchedRuns.push({
                primary: p,
                shadow: linked,
              });
            }
          });

          if (matchedRuns.length > 0) {
            setRuns(matchedRuns);
            setIsDemoData(false);
          } else {
            // No shadow runs in DB, fall back to mock runs mapped to active team ID
            setRuns(getMockRuns().map(run => ({
              primary: { ...run.primary, team_id: teamID },
              shadow: { ...run.shadow, team_id: teamID },
            })));
            setIsDemoData(true);
          }
        } else {
          setRuns(getMockRuns().map(run => ({
            primary: { ...run.primary, team_id: teamID },
            shadow: { ...run.shadow, team_id: teamID },
          })));
          setIsDemoData(true);
        }
        setRunsLoading(false);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        setRuns(getMockRuns().map(run => ({
          primary: { ...run.primary, team_id: teamID },
          shadow: { ...run.shadow, team_id: teamID },
        })));
        setIsDemoData(true);
        setRunsLoading(false);
        setLoading(false);
      });
  };

  const handleTeamChange = (teamID: string) => {
    setSelectedTeam(teamID);
    fetchMigrationRuns(teamID);
  };

  const getMockRuns = (): MigrationRun[] => {
    const baseDate = new Date();
    return [
      {
        primary: {
          id: "primary-1",
          team_id: "1",
          model: "gpt-4o",
          prompt_tokens: 3500,
          completion_tokens: 950,
          cost: 0.03175,
          latency_ms: 1240,
          created_at: new Date(baseDate.getTime() - 120000).toISOString(),
          is_shadow: false,
          primary_log_id: "",
        },
        shadow: {
          id: "shadow-1",
          team_id: "1",
          model: "claude-3-5-sonnet",
          prompt_tokens: 3500,
          completion_tokens: 980,
          cost: 0.02520,
          latency_ms: 1530,
          created_at: new Date(baseDate.getTime() - 118000).toISOString(),
          is_shadow: true,
          primary_log_id: "primary-1",
        },
      },
      {
        primary: {
          id: "primary-2",
          team_id: "1",
          model: "gpt-4o",
          prompt_tokens: 45000,
          completion_tokens: 15000,
          cost: 0.45000,
          latency_ms: 5450,
          created_at: new Date(baseDate.getTime() - 900000).toISOString(),
          is_shadow: false,
          primary_log_id: "",
        },
        shadow: {
          id: "shadow-2",
          team_id: "1",
          model: "claude-3-haiku-20240307",
          prompt_tokens: 45000,
          completion_tokens: 14800,
          cost: 0.02975,
          latency_ms: 1320,
          created_at: new Date(baseDate.getTime() - 898000).toISOString(),
          is_shadow: true,
          primary_log_id: "primary-2",
        },
      },
      {
        primary: {
          id: "primary-3",
          team_id: "1",
          model: "gpt-4-turbo",
          prompt_tokens: 8200,
          completion_tokens: 4200,
          cost: 0.20800,
          latency_ms: 3840,
          created_at: new Date(baseDate.getTime() - 3600000).toISOString(),
          is_shadow: false,
          primary_log_id: "",
        },
        shadow: {
          id: "shadow-3",
          team_id: "1",
          model: "gpt-3.5-turbo",
          prompt_tokens: 8200,
          completion_tokens: 4100,
          cost: 0.01025,
          latency_ms: 980,
          created_at: new Date(baseDate.getTime() - 3590000).toISOString(),
          is_shadow: true,
          primary_log_id: "primary-3",
        },
      },
    ];
  };

  // Calculate aggregates for current runs
  const totalPrimaryCost = runs.reduce((acc, r) => acc + r.primary.cost, 0);
  const totalShadowCost = runs.reduce((acc, r) => acc + r.shadow.cost, 0);
  const costSavings = totalPrimaryCost - totalShadowCost;
  const savingsPercent = totalPrimaryCost > 0 ? (costSavings / totalPrimaryCost) * 100 : 0;

  const avgPrimaryLatency = runs.length > 0 ? runs.reduce((acc, r) => acc + r.primary.latency_ms, 0) / runs.length : 0;
  const avgShadowLatency = runs.length > 0 ? runs.reduce((acc, r) => acc + r.shadow.latency_ms, 0) / runs.length : 0;
  const avgLatencyDiff = avgShadowLatency - avgPrimaryLatency;
  const avgLatencyDiffPercent = avgPrimaryLatency > 0 ? (avgLatencyDiff / avgPrimaryLatency) * 100 : 0;

  return (
    <div className="container">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "2rem" }}>
        <div>
          <h1 style={{ fontSize: "1.75rem", fontWeight: 700 }}>Bifrost Migration Auditor</h1>
          <p style={{ color: "var(--secondary)", fontSize: "0.875rem" }}>
            Real-time shadow-routing analysis to compare cost, speed, and tokens side-by-side with production models.
          </p>
        </div>
        <div style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
          <span style={{ fontSize: "0.825rem", fontWeight: "bold" }}>Trigger via header:</span>
          <code style={{
            background: "var(--primary-light)",
            color: "var(--primary)",
            padding: "0.3rem 0.5rem",
            borderRadius: "0.375rem",
            fontSize: "0.8rem",
            fontWeight: 600,
            fontFamily: "var(--font-geist-mono)"
          }}>
            X-Shadow-Model: claude-3-5-sonnet
          </code>
        </div>
      </div>

      {/* Aggregate comparison panels */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))", gap: "1.5rem", marginBottom: "2rem" }}>
        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Potential Cost Saved</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0", color: costSavings > 0 ? "var(--success)" : "inherit" }}>
            ${costSavings.toLocaleString(undefined, { minimumFractionDigits: 5, maximumFractionDigits: 5 })}
          </h2>
          <span style={{
            fontSize: "0.75rem",
            background: costSavings > 0 ? "var(--success-light)" : "var(--border)",
            color: costSavings > 0 ? "var(--success)" : "var(--secondary)",
            padding: "0.1rem 0.4rem",
            borderRadius: "1rem",
            fontWeight: 600
          }}>
            {savingsPercent.toFixed(1)}% Cost Reduction
          </span>
        </div>

        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Active Auditor Runs</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0" }}>
            {runs.length} Parallel Audits
          </h2>
          <div style={{ fontSize: "0.75rem", color: "var(--secondary)" }}>
            Capture and log comparisons on-the-fly
          </div>
        </div>

        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Average Latency Impact</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0", color: avgLatencyDiff > 0 ? "var(--warning)" : "var(--success)" }}>
            {avgLatencyDiff > 0 ? "+" : ""}{avgLatencyDiff.toFixed(0)} ms
          </h2>
          <span style={{
            fontSize: "0.75rem",
            background: avgLatencyDiff > 0 ? "var(--warning-light)" : "var(--success-light)",
            color: avgLatencyDiff > 0 ? "var(--warning)" : "var(--success)",
            padding: "0.1rem 0.4rem",
            borderRadius: "1rem",
            fontWeight: 600
          }}>
            {avgLatencyDiff > 0 ? `+${avgLatencyDiffPercent.toFixed(1)}% Slower` : `${Math.abs(avgLatencyDiffPercent).toFixed(1)}% Faster`}
          </span>
        </div>
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
          💡 <strong>Auditor Sandbox Preview:</strong> No live parallel shadow routing requests have been captured yet for this team. Showing simulated audit comparison metrics. Add the <code>X-Shadow-Model</code> header to your API clients to stream real audits here!
        </div>
      )}

      {/* Main Parallel Logs Table */}
      <div className="card" style={{ padding: 0, overflow: "hidden" }}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%" }}>
            <thead>
              <tr style={{ background: "var(--background)" }}>
                <th colSpan={2} style={{ borderRight: "1px solid var(--border)", textAlign: "center" }}>Primary Production Run</th>
                <th colSpan={2} style={{ borderRight: "1px solid var(--border)", textAlign: "center" }}>Background Shadow Run</th>
                <th colSpan={3} style={{ textAlign: "center" }}>Quantitative Comparison Delta</th>
              </tr>
              <tr>
                <th>Model</th>
                <th style={{ borderRight: "1px solid var(--border)" }}>Cost / Latency</th>
                <th>Model</th>
                <th style={{ borderRight: "1px solid var(--border)" }}>Cost / Latency</th>
                <th>Cost Savings</th>
                <th>Speed delta</th>
                <th>Audited At</th>
              </tr>
            </thead>
            <tbody>
              {runsLoading ? (
                <tr>
                  <td colSpan={7} style={{ textAlign: "center", padding: "3rem", color: "var(--secondary)" }}>
                    Analyzing background parallels...
                  </td>
                </tr>
              ) : runs.length === 0 ? (
                <tr>
                  <td colSpan={7} style={{ textAlign: "center", padding: "3rem", color: "var(--secondary)" }}>
                    No shadow routing telemetry audited yet.
                  </td>
                </tr>
              ) : (
                runs.map((run, index) => {
                  const costDiff = run.primary.cost - run.shadow.cost;
                  const runSavingsPercent = run.primary.cost > 0 ? (costDiff / run.primary.cost) * 100 : 0;

                  const latencyDiff = run.shadow.latency_ms - run.primary.latency_ms;
                  const runLatencyPercent = run.primary.latency_ms > 0 ? (latencyDiff / run.primary.latency_ms) * 100 : 0;

                  return (
                    <tr key={index}>
                      {/* Primary */}
                      <td style={{ fontWeight: 600 }}>
                        <span style={{ padding: "0.2rem 0.5rem", background: "rgba(79, 70, 229, 0.1)", color: "var(--primary)", borderRadius: "0.25rem", fontSize: "0.8rem" }}>
                          {run.primary.model}
                        </span>
                      </td>
                      <td style={{ borderRight: "1px solid var(--border)" }}>
                        <div style={{ display: "flex", flexDirection: "column" }}>
                          <span style={{ fontWeight: 600 }}>${run.primary.cost.toLocaleString(undefined, { minimumFractionDigits: 5, maximumFractionDigits: 5 })}</span>
                          <span style={{ fontSize: "0.75rem", color: "var(--secondary)" }}>{run.primary.latency_ms} ms</span>
                        </div>
                      </td>

                      {/* Shadow */}
                      <td style={{ fontWeight: 600 }}>
                        <span style={{ padding: "0.2rem 0.5rem", background: "rgba(217, 119, 6, 0.1)", color: "var(--warning)", borderRadius: "0.25rem", fontSize: "0.8rem" }}>
                          {run.shadow.model}
                        </span>
                      </td>
                      <td style={{ borderRight: "1px solid var(--border)" }}>
                        <div style={{ display: "flex", flexDirection: "column" }}>
                          <span style={{ fontWeight: 600 }}>${run.shadow.cost.toLocaleString(undefined, { minimumFractionDigits: 5, maximumFractionDigits: 5 })}</span>
                          <span style={{ fontSize: "0.75rem", color: "var(--secondary)" }}>{run.shadow.latency_ms} ms</span>
                        </div>
                      </td>

                      {/* Cost Delta */}
                      <td>
                        <span style={{
                          padding: "0.2rem 0.5rem",
                          borderRadius: "0.25rem",
                          fontSize: "0.8rem",
                          fontWeight: 700,
                          background: costDiff > 0 ? "var(--success-light)" : "var(--danger-light)",
                          color: costDiff > 0 ? "var(--success)" : "var(--danger)"
                        }}>
                          {costDiff > 0 ? "Saved" : "Extra"} ${Math.abs(costDiff).toLocaleString(undefined, { minimumFractionDigits: 5, maximumFractionDigits: 5 })} ({Math.abs(runSavingsPercent).toFixed(1)}%)
                        </span>
                      </td>

                      {/* Latency Delta */}
                      <td>
                        <span style={{
                          padding: "0.2rem 0.5rem",
                          borderRadius: "0.25rem",
                          fontSize: "0.8rem",
                          fontWeight: 700,
                          background: latencyDiff <= 0 ? "var(--success-light)" : "var(--warning-light)",
                          color: latencyDiff <= 0 ? "var(--success)" : "var(--warning)"
                        }}>
                          {latencyDiff <= 0 ? "Faster by " + Math.abs(latencyDiff) + "ms" : "Slower by " + latencyDiff + "ms"} ({Math.abs(runLatencyPercent).toFixed(0)}%)
                        </span>
                      </td>

                      {/* Date */}
                      <td style={{ fontSize: "0.825rem", color: "var(--secondary)" }}>
                        {new Date(run.primary.created_at).toLocaleTimeString()}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
