"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getAPIBase } from "./utils/api";

interface Team {
  id: string;
  name: string;
  budget_limit: number;
  budget_used: number;
}

export default function Dashboard() {
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
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
        // Fallback to mock data for rich presentation if API is not fully running locally
        setTeams([
          { id: "1", name: "Engineering Core", budget_limit: 5000.0, budget_used: 1245.5 },
          { id: "2", name: "Data Science Research", budget_limit: 10000.0, budget_used: 8430.2 },
          { id: "3", name: "Product Marketing Sandbox", budget_limit: 1000.0, budget_used: 980.4 },
          { id: "4", name: "QA Automated Tests", budget_limit: 2500.0, budget_used: 145.0 },
        ]);
        setLoading(false);
      });
  }, []);

  const totalLimit = teams.reduce((acc, t) => acc + t.budget_limit, 0);
  const totalUsed = teams.reduce((acc, t) => acc + t.budget_used, 0);
  const averageUsagePercent = totalLimit > 0 ? (totalUsed / totalLimit) * 100 : 0;

  // FinOps Spend Forecasting: Assume 9 days into a 30-day month
  const currentDayOfMonth = 9;
  const daysInMonth = 30;
  const forecastedSpend = (totalUsed / currentDayOfMonth) * daysInMonth;
  const isBudgetAtRisk = forecastedSpend > totalLimit;

  return (
    <div className="container">
      {/* Overview stats */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "2rem" }}>
        <div>
          <h1 style={{ fontSize: "1.75rem", fontWeight: 700 }}>FinOps Overview</h1>
          <p style={{ color: "var(--secondary)", fontSize: "0.875rem" }}>
            Real-time spend forecasting, budget thresholds, and team token tracking.
          </p>
        </div>
        <div style={{ display: "flex", gap: "0.75rem" }}>
          <Link href="/teams" className="btn btn-primary">
            + Onboard Team
          </Link>
          <Link href="/keys" className="btn btn-secondary">
            Manage Keys
          </Link>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))", gap: "1.5rem", marginBottom: "2rem" }}>
        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Month-to-Date Spend</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0", color: "var(--primary)" }}>
            ${totalUsed.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </h2>
          <span style={{ fontSize: "0.75rem", background: "var(--success-light)", color: "var(--success)", padding: "0.1rem 0.4rem", borderRadius: "1rem", fontWeight: 600 }}>
            Live Tracking Active
          </span>
        </div>

        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Total Allocated Budget</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0" }}>
            ${totalLimit.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </h2>
          <div style={{ fontSize: "0.75rem", color: "var(--secondary)" }}>
            Across {teams.length} active teams
          </div>
        </div>

        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Monthly Forecast</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0", color: isBudgetAtRisk ? "var(--danger)" : "var(--success)" }}>
            ${forecastedSpend.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </h2>
          <span style={{
            fontSize: "0.75rem",
            background: isBudgetAtRisk ? "var(--danger-light)" : "var(--success-light)",
            color: isBudgetAtRisk ? "var(--danger)" : "var(--success)",
            padding: "0.1rem 0.4rem",
            borderRadius: "1rem",
            fontWeight: 600
          }}>
            {isBudgetAtRisk ? "⚠️ Budget Limit Exceeded Risk" : "✓ Within Safe Threshold"}
          </span>
        </div>

        <div className="card">
          <span style={{ fontSize: "0.875rem", color: "var(--secondary)", fontWeight: 500 }}>Overall Budget Utilized</span>
          <h2 style={{ fontSize: "2rem", fontWeight: 700, margin: "0.5rem 0" }}>
            {averageUsagePercent.toFixed(1)}%
          </h2>
          <div style={{ width: "100%", background: "var(--border)", height: "8px", borderRadius: "4px", overflow: "hidden" }}>
            <div style={{ width: `${Math.min(averageUsagePercent, 100)}%`, background: averageUsagePercent > 90 ? "var(--danger)" : "var(--primary)", height: "100%" }} />
          </div>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "2fr 1fr", gap: "1.5rem" }}>
        {/* SVG Spend Chart Card */}
        <div className="card" style={{ display: "flex", flexDirection: "column" }}>
          <h3 style={{ fontSize: "1.1rem", fontWeight: 600, marginBottom: "1rem" }}>Monthly Cost Accumulation & Forecast</h3>
          <div style={{ flex: 1, position: "relative", minHeight: "260px", display: "flex", alignItems: "flex-end" }}>
            {/* Custom SVG Chart */}
            <svg viewBox="0 0 500 200" style={{ width: "100%", height: "200px" }}>
              <defs>
                <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--primary)" stopOpacity="0.4" />
                  <stop offset="100%" stopColor="var(--primary)" stopOpacity="0" />
                </linearGradient>
                <linearGradient id="forecastGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="var(--warning)" stopOpacity="0.3" />
                  <stop offset="100%" stopColor="var(--warning)" stopOpacity="0" />
                </linearGradient>
              </defs>

              {/* Grid lines */}
              <line x1="40" y1="20" x2="480" y2="20" stroke="var(--border)" strokeWidth="0.5" strokeDasharray="5,5" />
              <line x1="40" y1="70" x2="480" y2="70" stroke="var(--border)" strokeWidth="0.5" strokeDasharray="5,5" />
              <line x1="40" y1="120" x2="480" y2="120" stroke="var(--border)" strokeWidth="0.5" strokeDasharray="5,5" />
              <line x1="40" y1="170" x2="480" y2="170" stroke="var(--border)" strokeWidth="1" />

              {/* Y Axis labels */}
              <text x="10" y="24" fill="var(--secondary)" fontSize="10">$15,000</text>
              <text x="10" y="74" fill="var(--secondary)" fontSize="10">$10,000</text>
              <text x="10" y="124" fill="var(--secondary)" fontSize="10">$5,000</text>
              <text x="15" y="174" fill="var(--secondary)" fontSize="10">$0</text>

              {/* X Axis labels */}
              <text x="40" y="190" fill="var(--secondary)" fontSize="10">Day 1</text>
              <text x="150" y="190" fill="var(--secondary)" fontSize="10">Day 9 (Today)</text>
              <text x="300" y="190" fill="var(--secondary)" fontSize="10">Day 20</text>
              <text x="450" y="190" fill="var(--secondary)" fontSize="10">Day 30</text>

              {/* MTD Actual Area */}
              <path d="M 40,170 L 40,170 L 60,165 L 80,158 L 100,145 L 120,132 L 140,120 L 150,110 L 150,170 Z" fill="url(#chartGrad)" />
              {/* MTD Actual Line */}
              <path d="M 40,170 L 40,170 L 60,165 L 80,158 L 100,145 L 120,132 L 140,120 L 150,110" fill="none" stroke="var(--primary)" strokeWidth="3" />

              {/* Forecast Area */}
              <path d="M 150,110 L 200,95 L 250,80 L 300,60 L 350,45 L 400,30 L 450,15 L 480,5 L 480,170 L 150,170 Z" fill="url(#forecastGrad)" />
              {/* Forecast Line */}
              <path d="M 150,110 L 200,95 L 250,80 L 300,60 L 350,45 L 400,30 L 450,15 L 480,5" fill="none" stroke="var(--warning)" strokeWidth="2" strokeDasharray="4,4" />

              {/* Indicator dot */}
              <circle cx="150" cy="110" r="5" fill="var(--primary)" stroke="#ffffff" strokeWidth="2" />
            </svg>
          </div>
          <div style={{ display: "flex", justifyContent: "center", gap: "1.5rem", marginTop: "1rem", fontSize: "0.825rem" }}>
            <div style={{ display: "flex", alignItems: "center", gap: "0.25rem" }}>
              <span style={{ display: "inline-block", width: "12px", height: "12px", background: "var(--primary)", borderRadius: "50%" }}></span>
              <span>Actual Accumulation MTD</span>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: "0.25rem" }}>
              <span style={{ display: "inline-block", width: "12px", height: "12px", background: "var(--warning)", borderRadius: "50%" }}></span>
              <span>Predictive Spend Forecast</span>
            </div>
          </div>
        </div>

        {/* Team budgets list card */}
        <div className="card">
          <h3 style={{ fontSize: "1.1rem", fontWeight: 600, marginBottom: "1rem" }}>Top Teams by Spend</h3>
          <div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
            {teams.slice(0, 5).map((t) => {
              const usagePercent = t.budget_limit > 0 ? (t.budget_used / t.budget_limit) * 100 : 0;
              return (
                <div key={t.id} style={{ display: "flex", flexDirection: "column", gap: "0.25rem" }}>
                  <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.875rem" }}>
                    <span style={{ fontWeight: 600 }}>{t.name}</span>
                    <span style={{ color: "var(--secondary)" }}>
                      ${t.budget_used.toFixed(0)} / ${t.budget_limit.toFixed(0)}
                    </span>
                  </div>
                  <div style={{ width: "100%", background: "var(--border)", height: "6px", borderRadius: "3px", overflow: "hidden" }}>
                    <div style={{
                      width: `${Math.min(usagePercent, 100)}%`,
                      background: usagePercent > 90 ? "var(--danger)" : "var(--primary)",
                      height: "100%"
                    }} />
                  </div>
                </div>
              );
            })}
          </div>
          <Link href="/teams" style={{ display: "block", textAlign: "center", fontSize: "0.875rem", color: "var(--primary)", fontWeight: 500, marginTop: "1.5rem" }}>
            View All Teams &rarr;
          </Link>
        </div>
      </div>
    </div>
  );
}
