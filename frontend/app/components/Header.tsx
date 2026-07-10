"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

export default function Header() {
  const pathname = usePathname();

  const navItems = [
    { name: "Dashboard", href: "/" },
    { name: "Teams & Budgets", href: "/teams" },
    { name: "Gateway Keys", href: "/keys" },
    { name: "Providers Config", href: "/providers" },
    { name: "Audit Logs", href: "/usage" },
    { name: "Migration Auditor", href: "/migration" },
  ];

  return (
    <header style={{ background: "var(--panel)", borderBottom: "1px solid var(--border)", position: "sticky", top: 0, zIndex: 50 }}>
      <div className="container" style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "1rem 1.5rem" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
          <div style={{
            background: "linear-gradient(135deg, var(--primary) 0%, #a855f7 100%)",
            color: "#ffffff",
            padding: "0.4rem 0.6rem",
            borderRadius: "0.5rem",
            fontWeight: "bold",
            fontSize: "1.1rem",
            boxShadow: "0 2px 4px rgba(79, 70, 229, 0.25)"
          }}>
            🌉
          </div>
          <span style={{ fontWeight: 700, fontSize: "1.25rem", background: "linear-gradient(to right, var(--primary), #a855f7)", WebkitBackgroundClip: "text", WebkitTextFillColor: "transparent" }}>
            Bifrost
          </span>
          <span style={{ fontSize: "0.75rem", padding: "0.1rem 0.4rem", background: "var(--primary-light)", color: "var(--primary)", borderRadius: "1rem", fontWeight: 600 }}>
            Enterprise
          </span>
        </div>

        <nav style={{ display: "flex", gap: "1rem" }}>
          {navItems.map((item) => {
            const isActive = pathname === item.href || (item.href !== "/" && pathname?.startsWith(item.href));
            return (
              <Link
                key={item.href}
                href={item.href}
                style={{
                  padding: "0.5rem 0.75rem",
                  fontSize: "0.875rem",
                  fontWeight: 500,
                  color: isActive ? "var(--primary)" : "var(--secondary)",
                  background: isActive ? "var(--primary-light)" : "transparent",
                  borderRadius: "0.375rem",
                  transition: "all 0.2s",
                }}
              >
                {item.name}
              </Link>
            );
          })}
        </nav>
      </div>
    </header>
  );
}
