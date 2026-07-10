"use client";

import React, { createContext, useContext, useState, useEffect } from "react";

interface User {
  name: string;
  email: string;
  role: string; // "admin" | "developer" | "compliance"
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string) => Promise<void>;
  logout: () => void;
  switchRole: (role: string) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check local storage for persistent session
    const storedUser = localStorage.getItem("bifrost_session");
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  const login = async (email: string) => {
    setLoading(true);
    // Simulate OIDC / SAML secure handshake
    await new Promise((resolve) => setTimeout(resolve, 1500));

    let role = "developer";
    let name = "Alex Developer";

    if (email.includes("finops") || email.includes("admin")) {
      role = "admin";
      name = "Sarah FinOps Admin";
    } else if (email.includes("compliance") || email.includes("legal")) {
      role = "compliance";
      name = "Marcus Compliance Auditor";
    }

    const newUser = {
      name,
      email,
      role,
    };

    setUser(newUser);
    localStorage.setItem("bifrost_session", JSON.stringify(newUser));
    setLoading(false);
  };

  const logout = () => {
    setUser(null);
    localStorage.removeItem("bifrost_session");
  };

  const switchRole = (newRole: string) => {
    if (!user) return;
    const updatedUser = { ...user, role: newRole };
    setUser(updatedUser);
    localStorage.setItem("bifrost_session", JSON.stringify(updatedUser));
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, switchRole }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
