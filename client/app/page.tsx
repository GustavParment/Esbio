"use client";

import Link from "next/link";
import { useEffect } from "react";
import { useAuth } from "@/lib/contexts/AuthContext";
import { useTheme } from "@/lib/contexts/ThemeContext";
import { useRouter } from "next/navigation";

export default function LandingPage() {
  const { user, loading } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const router = useRouter();

  // Redirect authenticated users to dashboard
  useEffect(() => {
    if (!loading && user) {
      router.push("/dashboard");
    }
  }, [user, loading, router]);

  if (loading) return null;
  if (user) return null;

  return (
    <>
      <style jsx global>{`
        .landing-page {
          --ink: #FFFFFF;
          --ink-2: #F8F9FB;
          --ink-3: #EEF1F5;
          --teal: #0FA898;
          --teal-dim: rgba(15,168,152,0.08);
          --teal-mid: rgba(15,168,152,0.15);
          --l-text: #1A1A2E;
          --text-muted: #5A6578;
          --text-dim: #8A95A8;
          --l-border: rgba(0,0,0,0.08);
          --border-teal: rgba(15,168,152,0.2);
          font-family: 'DM Sans', sans-serif;
          background: var(--ink);
          color: var(--l-text);
          line-height: 1.6;
        }

        .dark .landing-page {
          --ink: #0D1117;
          --ink-2: #1A2233;
          --ink-3: #243045;
          --teal: #1ECAB8;
          --teal-dim: rgba(30,202,184,0.12);
          --teal-mid: rgba(30,202,184,0.25);
          --l-text: #E8EDF5;
          --text-muted: #7A8BA6;
          --text-dim: #4A5A72;
          --l-border: rgba(255,255,255,0.06);
          --border-teal: rgba(30,202,184,0.2);
        }

        @import url('https://fonts.googleapis.com/css2?family=DM+Serif+Display:ital@0;1&family=DM+Sans:wght@300;400;500&display=swap');

        .landing-nav {
          position: fixed; top: 0; left: 0; right: 0; z-index: 100;
          padding: 1.25rem 3rem;
          display: flex; align-items: center; justify-content: space-between;
          background: rgba(255,255,255,0.85);
          backdrop-filter: blur(20px);
          border-bottom: 1px solid var(--l-border);
        }

        .dark .landing-nav {
          background: rgba(13,17,23,0.8);
        }

        .landing-logo {
          font-family: 'DM Serif Display', serif;
          font-size: 1.6rem; color: var(--l-text); letter-spacing: -0.5px;
        }
        .landing-logo span { color: var(--teal); }

        .nav-links { display: flex; gap: 2.5rem; list-style: none; }
        .nav-links a {
          color: var(--text-muted); text-decoration: none;
          font-size: 0.9rem; font-weight: 400; transition: color 0.2s;
        }
        .nav-links a:hover { color: var(--l-text); }

        .nav-cta {
          background: var(--teal); color: var(--ink);
          padding: 0.6rem 1.5rem; border-radius: 100px;
          font-size: 0.875rem; font-weight: 500;
          text-decoration: none; transition: opacity 0.2s;
        }
        .nav-cta:hover { opacity: 0.85; }

        .hero {
          min-height: 100vh; display: flex; flex-direction: column;
          align-items: center; justify-content: center; text-align: center;
          padding: 8rem 2rem 6rem; position: relative; overflow: hidden;
        }

        .hero-glow {
          position: absolute; top: -200px; left: 50%; transform: translateX(-50%);
          width: 800px; height: 800px;
          background: radial-gradient(circle, rgba(15,168,152,0.06) 0%, transparent 70%);
          pointer-events: none;
        }

        .dark .hero-glow {
          background: radial-gradient(circle, rgba(30,202,184,0.08) 0%, transparent 70%);
        }

        .hero-badge {
          display: inline-flex; align-items: center; gap: 0.5rem;
          background: var(--teal-dim); border: 1px solid var(--border-teal);
          padding: 0.4rem 1rem; border-radius: 100px;
          font-size: 0.8rem; color: var(--teal); font-weight: 500;
          margin-bottom: 2rem; animation: fadeUp 0.6s ease both;
        }

        .hero-badge::before {
          content: ''; width: 6px; height: 6px; border-radius: 50%;
          background: var(--teal); animation: pulse 2s infinite;
        }

        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
        @keyframes fadeUp {
          from { opacity: 0; transform: translateY(20px); }
          to { opacity: 1; transform: translateY(0); }
        }

        .landing-h1 {
          font-family: 'DM Serif Display', serif;
          font-size: clamp(3rem, 7vw, 5.5rem);
          line-height: 1.05; letter-spacing: -1px;
          max-width: 900px; animation: fadeUp 0.6s 0.1s ease both;
          color: var(--l-text);
        }
        .landing-h1 em { font-style: italic; color: var(--teal); }

        .hero-sub {
          max-width: 540px; color: var(--text-muted);
          font-size: 1.1rem; font-weight: 300;
          margin-top: 1.5rem; line-height: 1.7;
          animation: fadeUp 0.6s 0.2s ease both;
        }

        .hero-actions {
          display: flex; gap: 1rem; align-items: center;
          margin-top: 2.5rem; animation: fadeUp 0.6s 0.3s ease both;
          flex-wrap: wrap; justify-content: center;
        }

        .btn-primary {
          background: var(--teal); color: var(--ink);
          padding: 0.85rem 2rem; border-radius: 100px;
          font-size: 0.95rem; font-weight: 500;
          text-decoration: none; transition: all 0.2s;
        }
        .btn-primary:hover { opacity: 0.9; transform: translateY(-1px); }

        .btn-ghost {
          color: var(--text-muted); padding: 0.85rem 2rem;
          border-radius: 100px; font-size: 0.95rem;
          text-decoration: none; border: 1px solid var(--l-border); transition: all 0.2s;
        }
        .btn-ghost:hover { border-color: var(--teal); color: var(--l-text); }

        .preview-wrapper {
          width: 100%; max-width: 900px; margin: 4rem auto 0;
          animation: fadeUp 0.8s 0.4s ease both; position: relative;
        }
        .preview-wrapper::before {
          content: ''; position: absolute; inset: -1px; border-radius: 20px;
          background: linear-gradient(135deg, var(--border-teal), transparent 60%); z-index: 0;
        }

        .dashboard-preview {
          position: relative; z-index: 1; background: var(--ink-2);
          border-radius: 20px; border: 1px solid var(--l-border); overflow: hidden;
          box-shadow: 0 20px 60px rgba(0,0,0,0.08);
        }

        .dark .dashboard-preview {
          box-shadow: none;
        }

        .dash-top {
          padding: 1rem 1.5rem; border-bottom: 1px solid var(--l-border);
          display: flex; align-items: center; gap: 0.75rem;
        }
        .dot { width: 10px; height: 10px; border-radius: 50%; }
        .dot-r { background: #FF5F57; }
        .dot-y { background: #FEBC2E; }
        .dot-g { background: #28C840; }

        .dash-inner { display: grid; grid-template-columns: 200px 1fr; min-height: 340px; }

        .dash-sidebar {
          border-right: 1px solid var(--l-border); padding: 1.5rem 1rem;
        }
        .sidebar-label {
          font-size: 0.65rem; color: var(--text-dim);
          text-transform: uppercase; letter-spacing: 1px; margin-bottom: 0.75rem;
        }
        .sidebar-item {
          display: flex; align-items: center; gap: 0.6rem;
          padding: 0.5rem 0.6rem; border-radius: 8px;
          font-size: 0.8rem; color: var(--text-muted); margin-bottom: 0.25rem;
        }
        .sidebar-item.active { background: var(--teal-dim); color: var(--teal); }

        .dash-main { padding: 1.5rem; }
        .dash-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; margin-bottom: 1.5rem; }

        .stat-card {
          background: var(--ink-3); border: 1px solid var(--l-border);
          border-radius: 12px; padding: 1rem;
        }
        .stat-label { font-size: 0.7rem; color: var(--text-dim); margin-bottom: 0.4rem; }
        .stat-value { font-size: 1.4rem; font-weight: 500; color: var(--l-text); }
        .stat-value.teal { color: var(--teal); }
        .stat-change { font-size: 0.7rem; color: #4CAF50; margin-top: 0.25rem; }

        .ai-suggestion {
          background: var(--teal-dim); border: 1px solid var(--border-teal);
          border-radius: 12px; padding: 1rem;
          display: flex; gap: 0.75rem; align-items: flex-start;
        }
        .ai-icon {
          width: 28px; height: 28px; border-radius: 8px;
          background: var(--teal); display: flex;
          align-items: center; justify-content: center;
          font-size: 0.75rem; font-weight: 700; color: var(--ink); flex-shrink: 0;
        }
        .ai-text { font-size: 0.78rem; color: var(--text-muted); line-height: 1.5; }
        .ai-text strong { color: var(--teal); font-weight: 500; }

        .landing-section { padding: 6rem 2rem; }

        .section-label {
          text-align: center; font-size: 0.75rem; text-transform: uppercase;
          letter-spacing: 2px; color: var(--teal); margin-bottom: 1rem;
        }
        .section-title {
          font-family: 'DM Serif Display', serif;
          font-size: clamp(2rem, 4vw, 3rem);
          text-align: center; max-width: 600px; margin: 0 auto 1rem;
          line-height: 1.1; color: var(--l-text);
        }
        .section-sub {
          text-align: center; color: var(--text-muted);
          max-width: 500px; margin: 0 auto 4rem; font-weight: 300; font-size: 1rem;
        }

        .features-grid {
          display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
          gap: 1.5rem; max-width: 1000px; margin: 0 auto;
        }
        .feature-card {
          background: var(--ink-2); border: 1px solid var(--l-border);
          border-radius: 16px; padding: 2rem; transition: border-color 0.2s, transform 0.2s;
        }
        .feature-card:hover { border-color: var(--border-teal); transform: translateY(-3px); }

        .feature-icon {
          width: 44px; height: 44px; border-radius: 12px;
          background: var(--teal-dim); border: 1px solid var(--border-teal);
          display: flex; align-items: center; justify-content: center;
          margin-bottom: 1.25rem; font-size: 1.25rem;
        }
        .feature-card h3 { font-size: 1.05rem; font-weight: 500; margin-bottom: 0.6rem; color: var(--l-text); }
        .feature-card p { color: var(--text-muted); font-size: 0.9rem; line-height: 1.6; font-weight: 300; }

        .how-section {
          background: var(--ink-2);
          border-top: 1px solid var(--l-border); border-bottom: 1px solid var(--l-border);
        }

        .steps { display: flex; flex-direction: column; max-width: 700px; margin: 0 auto; }
        .step {
          display: flex; gap: 2rem; padding: 2.5rem 0;
          border-bottom: 1px solid var(--l-border);
        }
        .step:last-child { border-bottom: none; }

        .step-num {
          font-family: 'DM Serif Display', serif;
          font-size: 3rem; color: var(--text-dim); line-height: 1;
          flex-shrink: 0; width: 3rem;
        }
        .step-content h3 { font-size: 1.1rem; font-weight: 500; margin-bottom: 0.5rem; color: var(--l-text); }
        .step-content p { color: var(--text-muted); font-weight: 300; font-size: 0.95rem; }

        .pricing-grid {
          display: grid; grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
          gap: 1.5rem; max-width: 860px; margin: 0 auto;
        }
        .price-card {
          background: var(--ink-2); border: 1px solid var(--l-border);
          border-radius: 20px; padding: 2rem; transition: transform 0.2s;
        }
        .price-card:hover { transform: translateY(-3px); }
        .price-card.featured {
          border-color: var(--border-teal);
          background: linear-gradient(160deg, rgba(30,202,184,0.06), var(--ink-2));
        }

        .popular-badge {
          display: inline-block; background: var(--teal); color: var(--ink);
          font-size: 0.7rem; font-weight: 500;
          padding: 0.25rem 0.75rem; border-radius: 100px; margin-bottom: 1rem;
        }
        .plan-name { font-size: 0.85rem; color: var(--text-muted); margin-bottom: 0.5rem; }
        .plan-price {
          font-family: 'DM Serif Display', serif;
          font-size: 2.5rem; line-height: 1; margin-bottom: 0.25rem; color: var(--l-text);
        }
        .plan-price span { font-family: 'DM Sans', sans-serif; font-size: 0.9rem; color: var(--text-muted); }
        .plan-desc { color: var(--text-muted); font-size: 0.85rem; margin-bottom: 1.5rem; font-weight: 300; }

        .plan-features { list-style: none; margin-bottom: 2rem; padding: 0; }
        .plan-features li {
          font-size: 0.875rem; color: var(--text-muted);
          padding: 0.4rem 0; display: flex; align-items: center; gap: 0.6rem;
        }
        .plan-features li::before {
          content: ''; width: 5px; height: 5px; border-radius: 50%;
          background: var(--teal); flex-shrink: 0;
        }

        .plan-btn {
          width: 100%; padding: 0.8rem; border-radius: 100px;
          font-size: 0.9rem; font-weight: 500; cursor: pointer;
          border: 1px solid var(--l-border); background: transparent;
          color: var(--l-text); transition: all 0.2s;
          text-decoration: none; display: block; text-align: center;
        }
        .price-card.featured .plan-btn {
          background: var(--teal); color: var(--ink); border-color: var(--teal);
        }
        .plan-btn:hover { opacity: 0.8; }

        .testimonials {
          display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
          gap: 1.5rem; max-width: 900px; margin: 0 auto;
        }
        .testimonial {
          background: var(--ink-2); border: 1px solid var(--l-border);
          border-radius: 16px; padding: 1.75rem; text-align: left;
        }
        .testimonial-text {
          font-style: italic; color: var(--text-muted);
          font-size: 0.95rem; line-height: 1.7; font-weight: 300; margin-bottom: 1.25rem;
        }
        .testimonial-author { display: flex; align-items: center; gap: 0.75rem; }
        .avatar-circle {
          width: 36px; height: 36px; border-radius: 50%;
          background: var(--teal-mid); border: 1px solid var(--border-teal);
          display: flex; align-items: center; justify-content: center;
          font-size: 0.75rem; font-weight: 500; color: var(--teal);
        }
        .author-name { font-size: 0.875rem; font-weight: 500; color: var(--l-text); }
        .author-role { font-size: 0.75rem; color: var(--text-dim); }

        .cta-section {
          text-align: center; padding: 6rem 2rem;
          position: relative; overflow: hidden;
        }
        .cta-glow {
          position: absolute; bottom: -300px; left: 50%; transform: translateX(-50%);
          width: 600px; height: 600px;
          background: radial-gradient(circle, rgba(15,168,152,0.06) 0%, transparent 70%);
          pointer-events: none;
        }

        .dark .cta-glow {
          background: radial-gradient(circle, rgba(30,202,184,0.1) 0%, transparent 70%);
        }
        .cta-section h2 {
          font-family: 'DM Serif Display', serif;
          font-size: clamp(2.5rem, 5vw, 4rem);
          line-height: 1.1; margin-bottom: 1.5rem; color: var(--l-text);
        }
        .cta-section p { color: var(--text-muted); font-weight: 300; margin-bottom: 2.5rem; }

        .landing-footer {
          border-top: 1px solid var(--l-border);
          padding: 2rem 3rem;
          display: flex; justify-content: space-between; align-items: center;
          color: var(--text-dim); font-size: 0.8rem;
          flex-wrap: wrap; gap: 1rem;
        }
        .footer-logo {
          font-family: 'DM Serif Display', serif;
          font-size: 1.2rem; color: var(--l-text);
        }
        .footer-logo span { color: var(--teal); }

        .theme-toggle {
          background: none; border: 1px solid var(--l-border);
          border-radius: 50%; width: 36px; height: 36px;
          display: flex; align-items: center; justify-content: center;
          cursor: pointer; font-size: 1rem; transition: all 0.2s;
        }
        .theme-toggle:hover { border-color: var(--teal); }

        @media (max-width: 768px) {
          .landing-nav { padding: 1rem 1.5rem; }
          .nav-links { display: none; }
          .dash-inner { grid-template-columns: 1fr; }
          .dash-sidebar { display: none; }
          .dash-stats { grid-template-columns: 1fr; }
          .landing-footer { flex-direction: column; text-align: center; }
        }
      `}</style>

      <div className="landing-page">
        {/* NAV */}
        <nav className="landing-nav">
          <img src="/esbio-logo.png" alt="Esbio" className="h-9" />
          <ul className="nav-links">
            <li><a href="#features">Funktioner</a></li>
            <li><a href="#how">Hur det fungerar</a></li>
            <li><a href="#pricing">Priser</a></li>
          </ul>
          <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
            <button
              onClick={toggleTheme}
              className="theme-toggle"
              aria-label="Byt tema"
            >
              {theme === "dark" ? "☀️" : "🌙"}
            </button>
            <Link href="/auth/login" className="nav-cta">Logga in</Link>
          </div>
        </nav>

        {/* HERO */}
        <section className="hero">
          <div className="hero-glow"></div>
          <div className="hero-badge">AI-driven bokföring för svenska företag</div>
          <h1 className="landing-h1">Bokföring som <em>tänker</em><br />åt dig</h1>
          <p className="hero-sub">Esbio hanterar din bokföring automatiskt med AI. Ladda upp kvitton, koppla ditt bankkonto och låt Esbio sköta resten.</p>
          <div className="hero-actions">
            <Link href="/auth/register" className="btn-primary">Testa gratis i 30 dagar</Link>
            <a href="#how" className="btn-ghost">Se hur det fungerar</a>
          </div>

          <div className="preview-wrapper">
            <div className="dashboard-preview">
              <div className="dash-top">
                <div className="dot dot-r"></div>
                <div className="dot dot-y"></div>
                <div className="dot dot-g"></div>
              </div>
              <div className="dash-inner">
                <div className="dash-sidebar">
                  <div className="sidebar-item active">&#9632; Dashboard</div>
                  <div className="sidebar-item">&#9632; Verifikat</div>
                  <div className="sidebar-item">&#9632; Konton</div>
                  <div className="sidebar-item">&#9632; Rapporter</div>
                  <div className="sidebar-item">&#9632; Ester AI</div>
                  <div className="sidebar-item">&#9632; Inställningar</div>
                </div>
                <div className="dash-main">
                  <div className="dash-stats">
                    <div className="stat-card">
                      <div className="stat-label">Verifikat denna månad</div>
                      <div className="stat-value teal">12</div>
                      <div className="stat-change">↑ 3 mot förra mån</div>
                    </div>
                    <div className="stat-card">
                      <div className="stat-label">Totalt konton</div>
                      <div className="stat-value">84</div>
                      <div className="stat-change" style={{ color: "var(--text-dim)" }}>BAS-kontoplan</div>
                    </div>
                    <div className="stat-card">
                      <div className="stat-label">Aktuell period</div>
                      <div className="stat-value">2026-04</div>
                      <div className="stat-change">Aktiv</div>
                    </div>
                  </div>
                  <div className="ai-suggestion">
                    <div className="ai-icon">AI</div>
                    <div className="ai-text">
                      <strong>Ester AI:</strong> Du har 3 okontererade transaktioner. Jag föreslår kontot <strong>5800 – Representation</strong>. Stämmer det? <span style={{ color: "var(--teal)", cursor: "pointer" }}>Bekräfta alla</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* FEATURES */}
        <section id="features" className="landing-section">
          <div className="section-label">Funktioner</div>
          <h2 className="section-title">Allt du behöver, ingenting du inte behöver</h2>
          <p className="section-sub">Esbio är byggt för svenska enskilda firmor och aktiebolag som vill slippa bokföringsträsket.</p>

          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: "var(--teal)"}}><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/></svg>
              </div>
              <h3>Kvittoläsning med AI</h3>
              <p>Ta en bild på kvittot. Esbio läser av belopp, datum och leverantör och skapar verifikaten åt dig automatiskt.</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: "var(--teal)"}}><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
              </div>
              <h3>Automatisk kontering</h3>
              <p>AI:n lär sig ditt företag och konterar transaktioner korrekt direkt. Du godkänner, vi sköter resten.</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: "var(--teal)"}}><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
              </div>
              <h3>Momsrapporter på ett klick</h3>
              <p>Korrekt momsredovisning varje kvartal. Esbio sammanställer rapporten och du godkänner med BankID.</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: "var(--teal)"}}><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
              </div>
              <h3>Fakturering inbyggt</h3>
              <p>Skapa och skicka fakturor direkt i Esbio. Betalningar bokförs automatiskt när kunden betalar.</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: "var(--teal)"}}><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              </div>
              <h3>Bankkoppling</h3>
              <p>Koppla ditt företagskonto så importeras alla transaktioner automatiskt, varje dag.</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{color: "var(--teal)"}}><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
              </div>
              <h3>Årsredovisning med AI</h3>
              <p>Esbio hjälper dig att sammanställa och lämna in din årsredovisning till Bolagsverket utan revisorskostnader.</p>
            </div>
          </div>
        </section>

        {/* ESTER AI DEMO */}
        <section className="landing-section" style={{ textAlign: "center" }}>
          <div className="section-label">Ester AI</div>
          <h2 className="section-title">Din personliga bokföringsassistent</h2>
          <p className="section-sub">Berätta vad du vill göra — Ester sköter resten. Skapa verifikat, schemalägga bokföringar och få svar på dina frågor.</p>

          <div style={{ maxWidth: "900px", margin: "0 auto", display: "flex", flexDirection: "column", gap: "2rem" }}>
            <div style={{ position: "relative" }}>
              <div style={{
                position: "absolute", inset: "-1px", borderRadius: "20px",
                background: "linear-gradient(135deg, var(--border-teal), transparent 60%)",
                zIndex: 0
              }} />
              <img
                src="/ester-demo.png"
                alt="Ester AI demo"
                style={{
                  width: "100%", borderRadius: "20px",
                  border: "1px solid var(--l-border)",
                  position: "relative", zIndex: 1
                }}
              />
            </div>
            <div style={{ position: "relative" }}>
              <div style={{
                position: "absolute", inset: "-1px", borderRadius: "20px",
                background: "linear-gradient(135deg, var(--border-teal), transparent 60%)",
                zIndex: 0
              }} />
              <img
                src="/ester-tasks-demo.png"
                alt="Ester AI schemalagda uppgifter"
                style={{
                  width: "100%", borderRadius: "20px",
                  border: "1px solid var(--l-border)",
                  position: "relative", zIndex: 1
                }}
              />
            </div>
          </div>
        </section>

        {/* HOW IT WORKS */}
        <section id="how" className="landing-section how-section">
          <div className="section-label">Hur det fungerar</div>
          <h2 className="section-title">Tre steg till fri bokföring</h2>
          <p className="section-sub">Kom igång på under tio minuter.</p>

          <div className="steps">
            <div className="step">
              <div className="step-num">01</div>
              <div className="step-content">
                <h3>Koppla ditt bankkonto</h3>
                <p>Esbio ansluter till din bank via Open Banking och importerar transaktioner automatiskt. Inga manuella CSV-filer.</p>
              </div>
            </div>
            <div className="step">
              <div className="step-num">02</div>
              <div className="step-content">
                <h3>AI:n konterar och kategoriserar</h3>
                <p>Esbio analyserar varje transaktion och föreslår rätt konto baserat på ditt företag, din bransch och historiska mönster.</p>
              </div>
            </div>
            <div className="step">
              <div className="step-num">03</div>
              <div className="step-content">
                <h3>Du granskar och godkänner</h3>
                <p>Du behåller kontrollen. Granska AI:ns förslag, gör justeringar om det behövs och godkänn. Esbio sköter resten.</p>
              </div>
            </div>
          </div>
        </section>

        {/* PRICING */}
        <section id="pricing" className="landing-section">
          <div className="section-label">Priser</div>
          <h2 className="section-title">Enkla priser utan överraskningar</h2>
          <p className="section-sub">Alla planer inkluderar AI-kontering, bankkoppling och momsrapporter.</p>

          <div className="pricing-grid">
            <div className="price-card">
              <div className="plan-name">Starter</div>
              <div className="plan-price">199 <span>kr/mån</span></div>
              <div className="plan-desc">Perfekt för enskild firma och frilansare.</div>
              <ul className="plan-features">
                <li>Upp till 100 transaktioner/mån</li>
                <li>AI-kontering</li>
                <li>Bankkoppling</li>
                <li>Momsrapporter</li>
                <li>5 fakturor/mån</li>
              </ul>
              <Link href="/auth/register" className="plan-btn">Kom igång</Link>
            </div>

            <div className="price-card featured">
              <div className="popular-badge">Mest populär</div>
              <div className="plan-name">Tillväxt</div>
              <div className="plan-price">399 <span>kr/mån</span></div>
              <div className="plan-desc">För aktiebolag och växande verksamheter.</div>
              <ul className="plan-features">
                <li>Obegränsade transaktioner</li>
                <li>AI-kontering med inlärning</li>
                <li>Bankkoppling</li>
                <li>Momsrapporter</li>
                <li>Obegränsade fakturor</li>
                <li>Årsredovisning ingår</li>
                <li>Prioriterad support</li>
              </ul>
              <Link href="/auth/register" className="plan-btn">Testa gratis 30 dagar</Link>
            </div>

            <div className="price-card">
              <div className="plan-name">Enterprise</div>
              <div className="plan-price">Offert</div>
              <div className="plan-desc">För bolag med komplexa behov och redovisningsbyråer.</div>
              <ul className="plan-features">
                <li>Allt i Tillväxt</li>
                <li>Flertalet bolag</li>
                <li>API-access</li>
                <li>Dedikerad kontaktperson</li>
                <li>Anpassad onboarding</li>
              </ul>
              <Link href="/auth/register" className="plan-btn">Kontakta oss</Link>
            </div>
          </div>
        </section>

        {/* TESTIMONIALS */}
        <section className="landing-section" style={{ textAlign: "center" }}>
          <div className="section-label">Vad kunderna säger</div>
          <h2 className="section-title" style={{ marginBottom: "3rem" }}>Tusentals svenska företag sparar tid med Esbio</h2>

          <div className="testimonials">
            <div className="testimonial">
              <p className="testimonial-text">&quot;Jag brukade lägga tre timmar i månaden på bokföring. Nu tar det tjugo minuter. Esbio är det bästa jag gjort för mitt företag.&quot;</p>
              <div className="testimonial-author">
                <div className="avatar-circle">AS</div>
                <div>
                  <div className="author-name">Anna Ström</div>
                  <div className="author-role">Grundare, Ström Design AB</div>
                </div>
              </div>
            </div>
            <div className="testimonial">
              <p className="testimonial-text">&quot;AI:n lär sig snabbt. Efter tre veckor konterade den rätt på nästan allt. Momsrapporterna sköter sig helt själva nu.&quot;</p>
              <div className="testimonial-author">
                <div className="avatar-circle">MK</div>
                <div>
                  <div className="author-name">Marcus Karlsson</div>
                  <div className="author-role">VD, Nordisk Konsult AB</div>
                </div>
              </div>
            </div>
            <div className="testimonial">
              <p className="testimonial-text">&quot;Som frilansare ville jag inte betala en revisor 5000 kr för enkla saker. Esbio gör jobbet för en bråkdel av kostnaden.&quot;</p>
              <div className="testimonial-author">
                <div className="avatar-circle">LJ</div>
                <div>
                  <div className="author-name">Linn Johansson</div>
                  <div className="author-role">Frilansfotograf</div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="cta-section">
          <div className="cta-glow"></div>
          <h2>Sluta ägna tid<br /><em style={{ fontStyle: "italic", color: "var(--teal)" }}>åt bokföring</em></h2>
          <p>Testa Esbio gratis i 30 dagar. Inget kreditkort krävs.</p>
          <Link href="/auth/register" className="btn-primary" style={{ fontSize: "1rem", padding: "1rem 2.5rem" }}>Kom igång gratis</Link>
        </section>

        {/* FOOTER */}
        <footer className="landing-footer">
          <img src="/esbio-logo.png" alt="Esbio" className="h-8" />
          <div>© 2025 Parment Software Solutions AB</div>
          <div>Integritetspolicy · Villkor · Support</div>
        </footer>
      </div>
    </>
  );
}
