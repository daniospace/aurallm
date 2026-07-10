"use client";

import React, { useState } from "react";
import { useAuth } from "../context/AuthContext";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const { user, loading, login, logout, switchRole } = useAuth();
  const [email, setEmail] = useState("");
  const [signingIn, setSigningIn] = useState(false);

  const handleSSOSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) return;

    setSigningIn(true);
    await login(email);
    setSigningIn(false);
  };

  if (loading) {
    return (
      <div style={{
        height: "100vh",
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        alignItems: "center",
        background: "var(--background)",
        gap: "1rem"
      }}>
        <div style={{
          width: "40px",
          height: "40px",
          border: "4px solid var(--border)",
          borderTopColor: "var(--primary)",
          borderRadius: "50%",
          animation: "spin 1s linear infinite"
        }} />
        <style>{`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `}</style>
        <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>
          Securing Bifrost OIDC Handshake...
        </span>
      </div>
    );
  }

  if (!user) {
    return (
      <div style={{
        minHeight: "100vh",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        background: "linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%)",
        padding: "1.5rem"
      }}>
        <div className="card" style={{ maxWidth: "420px", width: "100%", padding: "2.5rem", border: "1px solid rgba(255,255,255,0.08)", background: "rgba(30, 41, 59, 0.7)", backdropFilter: "blur(12px)", boxShadow: "0 20px 25px -5px rgba(0, 0, 0, 0.3)" }}>
          <div style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: "0.5rem", marginBottom: "2rem" }}>
            <div style={{
              background: "linear-gradient(135deg, var(--primary) 0%, #a855f7 100%)",
              color: "#ffffff",
              padding: "0.6rem 0.8rem",
              borderRadius: "0.75rem",
              fontWeight: "bold",
              fontSize: "1.75rem",
              boxShadow: "0 4px 12px rgba(79, 70, 229, 0.4)"
            }}>
              🌉
            </div>
            <h2 style={{ fontSize: "1.5rem", fontWeight: 800, color: "#ffffff", marginTop: "0.75rem" }}>
              Bifrost Enterprise Portal
            </h2>
            <p style={{ color: "#94a3b8", fontSize: "0.825rem", textAlign: "center" }}>
              Commercial-Grade, FinOps-Enforced AI Gateway Console
            </p>
          </div>

          <form onSubmit={handleSSOSubmit}>
            <div className="form-group" style={{ marginBottom: "1.5rem" }}>
              <label className="label" style={{ color: "#cbd5e1" }}>Enterprise SSO Domain Email</label>
              <input
                type="email"
                className="input"
                style={{ background: "rgba(15, 23, 42, 0.6)", border: "1px solid rgba(255,255,255,0.1)", color: "#ffffff" }}
                placeholder="e.g. admin@yourcompany.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={signingIn}
              />
              <span style={{ fontSize: "0.7rem", color: "#94a3b8", marginTop: "0.3rem" }}>
                💡 <strong>Demo Accounts:</strong> Use <code>admin@bifrost.ai</code> for Admin, <code>compliance@bifrost.ai</code> for Compliance, or any email for Developer.
              </span>
            </div>

            <button
              type="submit"
              className="btn btn-primary"
              style={{
                width: "100%",
                padding: "0.75rem",
                fontWeight: 600,
                background: "linear-gradient(135deg, var(--primary) 0%, #a855f7 100%)",
                border: "none",
                boxShadow: "0 4px 6px -1px rgba(79, 70, 229, 0.2)"
              }}
              disabled={signingIn}
            >
              {signingIn ? "Securing SSO Tunnel..." : "Continue with SSO (Okta / SAML)"}
            </button>
          </form>

          <div style={{ marginTop: "2rem", borderTop: "1px solid rgba(255,255,255,0.08)", paddingTop: "1rem", textAlign: "center" }}>
            <span style={{ fontSize: "0.75rem", color: "#64748b" }}>
              🛡️ SOC2 Type II, HIPAA, and GDPR Compliance Enforced
            </span>
          </div>
        </div>
      </div>
    );
  }

  // Active SSO session, display the dashboard and include the SSO RBAC controller at the header level
  return (
    <>
      <div style={{
        background: "var(--primary)",
        color: "#ffffff",
        padding: "0.35rem 1rem",
        fontSize: "0.75rem",
        fontWeight: 600,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        boxShadow: "inset 0 -1px 0 rgba(255,255,255,0.1)"
      }}>
        <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
          <span>🔒 SSO active session: <strong>{user.name}</strong> ({user.email})</span>
          <span style={{
            background: "rgba(255,255,255,0.15)",
            padding: "0.1rem 0.5rem",
            borderRadius: "0.25rem",
            fontSize: "0.7rem",
            textTransform: "uppercase"
          }}>
            Role: {user.role === "admin" ? "FinOps Admin" : user.role === "compliance" ? "Compliance Auditor" : "Developer Lead"}
          </span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
          {/* Quick simulator switcher for presentation demo! */}
          <div style={{ display: "flex", alignItems: "center", gap: "0.35rem" }}>
            <span style={{ fontSize: "0.7rem", opacity: 0.8 }}>⚡ Demo Role Switcher:</span>
            <select
              style={{
                background: "rgba(255,255,255,0.15)",
                color: "#ffffff",
                border: "none",
                fontSize: "0.7rem",
                padding: "0.1rem 0.25rem",
                borderRadius: "0.25rem",
                outline: "none",
                cursor: "pointer"
              }}
              value={user.role}
              onChange={(e) => switchRole(e.target.value)}
            >
              <option value="admin" style={{ color: "black" }}>FinOps Admin</option>
              <option value="developer" style={{ color: "black" }}>Developer Lead</option>
              <option value="compliance" style={{ color: "black" }}>Compliance Auditor</option>
            </select>
          </div>
          <button
            onClick={logout}
            style={{
              background: "transparent",
              border: "none",
              color: "rgba(255,255,255,0.8)",
              cursor: "pointer",
              fontSize: "0.75rem",
              fontWeight: "bold",
              textDecoration: "underline"
            }}
          >
            Logout SSO
          </button>
        </div>
      </div>
      {children}
    </>
  );
}
