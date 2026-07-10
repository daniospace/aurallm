"use client";

import { useEffect, useState } from "react";
import { getAPIBase } from "../utils/api";

interface Team {
  id: string;
  name: string;
}

interface GeneratedKey {
  key: string;
  key_hash: string;
  team_id: string;
  status: string;
}

export default function KeysPage() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeam, setSelectedTeam] = useState("");
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [newKey, setNewKey] = useState<GeneratedKey | null>(null);
  const [keysList, setKeysList] = useState<{ keyHashDisplay: string; keyHashRaw: string; teamName: string; status: string; date: string }[]>([]);
  const [copied, setCopied] = useState(false);

  const fetchKeys = () => {
    const apiBase = getAPIBase();
    fetch(`${apiBase}/api/keys`, { cache: "no-store" })
      .then((res) => {
        if (!res.ok) throw new Error("Failed to fetch keys");
        return res.json();
      })
      .then((data) => {
        if (Array.isArray(data)) {
          const formatted = data.map((k: any) => ({
            keyHashDisplay: k.key_hash.slice(0, 8) + "..." + k.key_hash.slice(-4),
            keyHashRaw: k.key_hash,
            teamName: k.team_name,
            status: k.status,
            date: new Date(k.created_at).toISOString().split("T")[0],
          }));
          setKeysList(formatted);
        } else {
          setKeysList([]);
        }
      })
      .catch((err) => {
        console.error(err);
        // Fallback mock keys
        setKeysList([
          { keyHashDisplay: "8f731ea2...c412", keyHashRaw: "8f731ea2", teamName: "Engineering Core", status: "active", date: "2026-07-01" },
          { keyHashDisplay: "0a1b2c3d...ef45", keyHashRaw: "0a1b2c3d", teamName: "Data Science Research", status: "active", date: "2026-07-05" },
          { keyHashDisplay: "f9e8d7c6...b5a4", keyHashRaw: "f9e8d7c6", teamName: "Product Marketing Sandbox", status: "suspended", date: "2026-06-15" },
        ]);
      });
  };

  const handleDeleteKey = async (keyHashRaw: string) => {
    if (!confirm("Are you sure you want to revoke and delete this gateway key permanently? This action is irreversible.")) {
      return;
    }

    try {
      const apiBase = getAPIBase();
      const res = await fetch(`${apiBase}/api/keys?key_hash=${keyHashRaw}`, {
        method: "DELETE",
      });

      if (!res.ok) throw new Error("Failed to delete key");
      fetchKeys();
    } catch (err) {
      console.error(err);
      // Demo fallback: delete from local state
      setKeysList((prev) => prev.filter(k => k.keyHashRaw !== keyHashRaw));
    }
  };

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
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        const fallbackTeams = [
          { id: "1", name: "Engineering Core" },
          { id: "2", name: "Data Science Research" },
          { id: "3", name: "Product Marketing Sandbox" },
          { id: "4", name: "QA Automated Tests" },
        ];
        setTeams(fallbackTeams);
        setSelectedTeam("1");
        setLoading(false);
      });

    fetchKeys();
  }, []);

  const handleGenerateKey = async () => {
    if (!selectedTeam) return;
    setGenerating(true);
    setNewKey(null);
    setCopied(false);

    try {
      const apiBase = getAPIBase();
      const res = await fetch(`${apiBase}/api/keys`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ team_id: selectedTeam }),
      });

      if (!res.ok) throw new Error("Failed to generate gateway key");
      
      const keyData = await res.json();
      setNewKey(keyData);
      fetchKeys(); // Reload list dynamically from database
    } catch (err) {
      console.error(err);
      // Demo fallback key generation
      const mockKey = "gw-" + Array.from({ length: 32 }, () => Math.floor(Math.random() * 16).toString(16)).join("");
      const mockHash = "mock_hash_" + Math.random().toString();
      const mockKeyData = {
        key: mockKey,
        key_hash: mockHash,
        team_id: selectedTeam,
        status: "active"
      };
      setNewKey(mockKeyData);

      const targetTeam = teams.find(t => t.id === selectedTeam);
      setKeysList((prev) => [
        {
          keyHashDisplay: mockHash.slice(0, 8) + "..." + mockHash.slice(-4),
          keyHashRaw: mockHash,
          teamName: targetTeam ? targetTeam.name : "Unknown",
          status: "active",
          date: new Date().toISOString().split("T")[0]
        },
        ...prev
      ]);
    } finally {
      setGenerating(false);
    }
  };

  const copyToClipboard = () => {
    if (!newKey) return;
    navigator.clipboard.writeText(newKey.key);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="container">
      <div style={{ marginBottom: "2rem" }}>
        <h1 style={{ fontSize: "1.75rem", fontWeight: 700 }}>Gateway Access Keys</h1>
        <p style={{ color: "var(--secondary)", fontSize: "0.875rem" }}>
          Generate highly secure gateway-level API credentials linked directly to team budgets.
        </p>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 2fr", gap: "2rem", alignItems: "flex-start" }}>
        {/* Left Side: Create Key panel */}
        <div style={{ display: "flex", flexDirection: "column", gap: "1.5rem" }}>
          <div className="card">
            <h3 style={{ fontSize: "1.1rem", fontWeight: 600, marginBottom: "1.25rem" }}>Generate Gateway Key</h3>
            
            <div className="form-group">
              <label className="label">Link to Team / Department</label>
              <select
                className="input"
                style={{ width: "100%", padding: "0.5rem" }}
                value={selectedTeam}
                onChange={(e) => setSelectedTeam(e.target.value)}
                disabled={loading}
              >
                {loading ? (
                  <option>Loading teams...</option>
                ) : teams.length === 0 ? (
                  <option>No teams onboarded</option>
                ) : (
                  teams.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name}
                    </option>
                  ))
                )}
              </select>
            </div>

            <button
              onClick={handleGenerateKey}
              className="btn btn-primary"
              style={{ width: "100%", marginTop: "0.5rem" }}
              disabled={generating || teams.length === 0}
            >
              {generating ? "Generating..." : "Generate Secure Key"}
            </button>
          </div>

          {/* Secure Display Box */}
          {newKey && (
            <div className="card" style={{ border: "1px solid var(--primary)", background: "var(--primary-light)", display: "flex", flexDirection: "column", gap: "0.75rem" }}>
              <span style={{ fontSize: "0.8rem", color: "var(--primary)", fontWeight: 600 }}>🔒 ONCE-ONLY SECRET KEY</span>
              <p style={{ fontSize: "0.75rem", color: "var(--secondary)" }}>
                For security, this key cannot be shown again. Store it safely immediately.
              </p>
              
              <div style={{ display: "flex", gap: "0.5rem", marginTop: "0.25rem" }}>
                <input
                  type="text"
                  className="input"
                  style={{ flex: 1, fontFamily: "var(--font-geist-mono)", fontWeight: "bold", border: "1px solid var(--primary)", color: "var(--foreground)", background: "var(--panel)" }}
                  value={newKey.key}
                  readOnly
                />
                <button onClick={copyToClipboard} className="btn btn-primary" style={{ padding: "0 1rem" }}>
                  {copied ? "Copied!" : "Copy"}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Right Side: Keys List */}
        <div className="card" style={{ padding: 0, overflow: "hidden" }}>
          <div style={{ padding: "1.25rem 1.5rem", borderBottom: "1px solid var(--border)" }}>
            <h3 style={{ fontSize: "1.1rem", fontWeight: 600 }}>Active Credentials</h3>
          </div>

          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%" }}>
              <thead>
                <tr>
                  <th>Key Hash ID</th>
                  <th>Assigned Team</th>
                  <th>Status</th>
                  <th>Created Date</th>
                  <th style={{ textAlign: "right" }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {keysList.map((k, index) => (
                  <tr key={index}>
                    <td style={{ fontFamily: "var(--font-geist-mono)", color: "var(--secondary)", fontSize: "0.825rem" }}>{k.keyHashDisplay}</td>
                    <td style={{ fontWeight: 600 }}>{k.teamName}</td>
                    <td>
                      <span style={{
                        padding: "0.15rem 0.5rem",
                        borderRadius: "1rem",
                        fontSize: "0.75rem",
                        fontWeight: 600,
                        background: k.status === "active" ? "var(--success-light)" : "var(--danger-light)",
                        color: k.status === "active" ? "var(--success)" : "var(--danger)"
                      }}>
                        {k.status}
                      </span>
                    </td>
                    <td>{k.date}</td>
                    <td style={{ textAlign: "right" }}>
                      <button
                        onClick={() => handleDeleteKey(k.keyHashRaw)}
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
                        Revoke
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
