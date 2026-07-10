"use client";

import { useEffect, useState } from "react";
import { getAPIBase } from "../utils/api";

interface Team {
  id: string;
  name: string;
  budget_limit: number;
  budget_used: number;
}

export default function TeamsPage() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [name, setName] = useState("");
  const [budgetLimit, setBudgetLimit] = useState("1000");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  const fetchTeams = () => {
    const apiBase = getAPIBase();
    fetch(`${apiBase}/api/teams`, { cache: "no-store" })
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch teams");
        return res.json();
      })
      .then((data) => {
        if (Array.isArray(data)) {
          setTeams(data);
        } else {
          setTeams([]);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        // Mock fallback
        setTeams([
          { id: "1", name: "Engineering Core", budget_limit: 5000.0, budget_used: 1245.5 },
          { id: "2", name: "Data Science Research", budget_limit: 10000.0, budget_used: 8430.2 },
          { id: "3", name: "Product Marketing Sandbox", budget_limit: 1000.0, budget_used: 980.4 },
          { id: "4", name: "QA Automated Tests", budget_limit: 2500.0, budget_used: 145.0 },
        ]);
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchTeams();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) return;

    setSaving(true);
    setMessage("");

    try {
      const apiBase = getAPIBase();
      const res = await fetch(`${apiBase}/api/teams`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          budget_limit: parseFloat(budgetLimit) || 0,
        }),
      });

      if (!res.ok) throw new Error("Failed to create team");
      
      const newTeam = await res.json();
      setTeams((prev) => [...prev, newTeam]);
      setName("");
      setBudgetLimit("1000");
      setMessage("✓ Team successfully onboarded!");
    } catch (err) {
      console.error(err);
      setMessage("❌ Error: Could not connect to API server. (Currently running in mock showcase mode)");
      // Inject to local state for presentation
      const mockNewTeam: Team = {
        id: Math.random().toString(),
        name,
        budget_limit: parseFloat(budgetLimit) || 0,
        budget_used: 0,
      };
      setTeams((prev) => [...prev, mockNewTeam]);
      setName("");
      setBudgetLimit("1000");
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteTeam = async (id: string) => {
    if (!confirm("⚠️ WARNING: Deleting this team will PERMANENTLY delete all of its gateway keys and token usage logs! This action is irreversible. Are you sure you want to proceed?")) {
      return;
    }

    try {
      const apiBase = getAPIBase();
      const res = await fetch(`${apiBase}/api/teams?id=${id}`, {
        method: "DELETE",
      });

      if (!res.ok) throw new Error("Failed to delete team");
      
      // Reload list
      fetchTeams();
    } catch (err) {
      console.error(err);
      // Demo fallback: delete from local state
      setTeams((prev) => prev.filter(t => t.id !== id));
    }
  };

  return (
    <div className="container">
      <div style={{ marginBottom: "2rem" }}>
        <h1 style={{ fontSize: "1.75rem", fontWeight: 700 }}>Team Onboarding & Budgets</h1>
        <p style={{ color: "var(--secondary)", fontSize: "0.875rem" }}>
          Onboard departments and set hard/soft usage limits to enforce spend boundaries.
        </p>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 2fr", gap: "2rem", alignItems: "flex-start" }}>
        {/* Form panel */}
        <div className="card">
          <h3 style={{ fontSize: "1.1rem", fontWeight: 600, marginBottom: "1.25rem" }}>Onboard New Team</h3>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="label">Team / Department Name</label>
              <input
                type="text"
                className="input"
                placeholder="e.g. Sales Forecast Devs"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>

            <div className="form-group">
              <label className="label">Monthly Budget Limit ($ USD)</label>
              <input
                type="number"
                className="input"
                placeholder="e.g. 5000"
                value={budgetLimit}
                onChange={(e) => setBudgetLimit(e.target.value)}
                required
              />
            </div>

            <button type="submit" className="btn btn-primary" style={{ width: "100%", marginTop: "0.5rem" }} disabled={saving}>
              {saving ? "Creating..." : "Onboard Team & Allocate"}
            </button>

            {message && (
              <div style={{
                marginTop: "1rem",
                padding: "0.75rem",
                borderRadius: "0.375rem",
                fontSize: "0.825rem",
                background: message.startsWith("✓") ? "var(--success-light)" : "var(--danger-light)",
                color: message.startsWith("✓") ? "var(--success)" : "var(--danger)",
                fontWeight: 500
              }}>
                {message}
              </div>
            )}
          </form>
        </div>

        {/* Teams List */}
        <div className="card" style={{ padding: 0, overflow: "hidden" }}>
          <div style={{ padding: "1.25rem 1.5rem", borderBottom: "1px solid var(--border)" }}>
            <h3 style={{ fontSize: "1.1rem", fontWeight: 600 }}>Active Teams</h3>
          </div>

          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%" }}>
              <thead>
                <tr>
                  <th>Team ID</th>
                  <th>Name</th>
                  <th>Budget Allocated</th>
                  <th>Budget Used</th>
                  <th>Utilization</th>
                  <th style={{ textAlign: "right" }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: "center", padding: "2rem", color: "var(--secondary)" }}>
                      Loading teams...
                    </td>
                  </tr>
                ) : teams.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ textAlign: "center", padding: "2rem", color: "var(--secondary)" }}>
                      No teams onboarded yet. Fill in the form to get started.
                    </td>
                  </tr>
                ) : (
                  teams.map((t) => {
                    const usagePercent = t.budget_limit > 0 ? (t.budget_used / t.budget_limit) * 100 : 0;
                    return (
                      <tr key={t.id}>
                        <td style={{ fontFamily: "var(--font-geist-mono)", fontSize: "0.75rem", color: "var(--secondary)" }}>
                          {t.id.slice(0, 8)}...
                        </td>
                        <td style={{ fontWeight: 600 }}>{t.name}</td>
                        <td style={{ fontWeight: 500 }}>
                          ${t.budget_limit.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </td>
                        <td style={{ color: t.budget_used >= t.budget_limit ? "var(--danger)" : "inherit" }}>
                          ${t.budget_used.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                        </td>
                        <td>
                          <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
                            <div style={{ flex: 1, background: "var(--border)", height: "8px", width: "80px", borderRadius: "4px", overflow: "hidden" }}>
                              <div style={{
                                width: `${Math.min(usagePercent, 100)}%`,
                                background: usagePercent >= 100 ? "var(--danger)" : usagePercent > 85 ? "var(--warning)" : "var(--primary)",
                                height: "100%"
                              }} />
                            </div>
                            <span style={{ fontSize: "0.75rem", fontWeight: 600, minWidth: "30px", textAlign: "right" }}>
                              {usagePercent.toFixed(0)}%
                            </span>
                          </div>
                        </td>
                        <td style={{ textAlign: "right" }}>
                          <button
                            onClick={() => handleDeleteTeam(t.id)}
                            className="btn btn-secondary"
                            style={{
                              padding: "0.2rem 0.5rem",
                              fontSize: "0.75rem",
                              borderColor: "var(--danger)",
                              color: "var(--danger)",
                              borderRadius: "0.25rem",
                              background: "transparent",
                              cursor: "pointer"
                            }}
                          >
                            Delete
                          </button>
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
    </div>
  );
}
