/* global React, THREE */
const { useState, useEffect, useRef } = React;

/* Mocked nav, used by all four directions */
function MockNav() {
  return (
    <nav className="hNav">
      <img className="hNav-mark" src="assets/ma-mark.svg" alt="A & M monogram" />
      <div className="hNav-links">
        <a>FAQs</a>
        <a>Travel</a>
        <a>Things to Do</a>
        <a>Registry</a>
        <a className="hNav-rsvp">RSVP</a>
      </div>
    </nav>
  );
}

/* ====================================================================== */
/* E — Synthesis                                                           */
/* C's sky pushed to full night at the top so the upper half reads as      */
/* space — stars + a quiet orbital element live there. The horizon glows   */
/* warm at the bottom. A's editorial type composition is anchored just     */
/* above the horizon line — a transition from the mundane (warm earth)     */
/* upward into night and orbit.                                            */
/* ====================================================================== */
/* Deterministic, realistic-ish night sky: a dense scatter PLUS a few
   clusters seeded around feature points so density varies the way real
   star fields do (denser in some regions, sparse in others). */
const HERO_E_STARS = (() => {
  let s = 91827;
  const rand = () => { s = (s * 1103515245 + 12345) & 0x7fffffff; return s / 0x7fffffff; };
  const list = [];
  // Background scatter — strong upward bias so density tapers into the
  // horizon glow (the journey: deep space at top → sunset at bottom).
  for (let i = 0; i < 280; i++) {
    list.push({
      x: rand() * 1400,
      y: Math.pow(rand(), 2.4) * 720,
      r: rand() < 0.06 ? 2.4 : rand() < 0.22 ? 1.6 : rand() < 0.55 ? 1.1 : 0.6,
      o: 0.25 + rand() * 0.6,
    });
  }
  // Cluster seeds — all kept high in the sky
  const clusters = [
    [180, 100, 70], [380, 180, 90], [720, 70, 80],
    [1080, 130, 100], [560, 240, 90], [1240, 200, 90],
    [220, 320, 80], [960, 360, 80],
  ];
  for (const [cx, cy, spread] of clusters) {
    const n = 14 + Math.floor(rand() * 10);
    for (let i = 0; i < n; i++) {
      // Gaussian-ish offset (sum-of-uniforms approximation)
      const dx = (rand() + rand() + rand() - 1.5) * spread;
      const dy = (rand() + rand() + rand() - 1.5) * spread * 0.7;
      list.push({
        x: cx + dx, y: cy + dy,
        r: rand() < 0.1 ? 2.0 : rand() < 0.4 ? 1.3 : 0.8,
        o: 0.35 + rand() * 0.55,
      });
    }
  }
  return list;
})();

/* 4-pointed concave sparkle — used for a few "feature" stars */
function HeroESparkle({ cx, cy, size = 12, fill = '#f6e7c8', opacity = 1 }) {
  const L = size, m = size * 0.18;
  const pts = [
    [0, -L], [m, -m], [L, 0], [m, m],
    [0, L], [-m, m], [-L, 0], [-m, -m],
  ].map(([x, y]) => `${cx + x},${cy + y}`).join(' ');
  return <polygon points={pts} fill={fill} opacity={opacity} />;
}

function HeroSynthesis() {
  // Pure render. The whole sky is one static SVG scene; mountains are a
  // separate layered SVG that simulates the Sangre de Cristo ridges.
  return (
    <header className="heroE">
      <div className="heroE-sky" />

      {/* The cosmic scene: nebula clouds, dense varied star field, milky-way
          band, a few feature sparkles, ringed planet, big crescent moon.    */}
      <svg className="heroE-skyfx" viewBox="0 0 1400 900" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
        <defs>
          {/* deep oxblood smolder — much dimmer, almost ember-like */}
          <radialGradient id="hE-neb-rust" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor="#7a2418" stopOpacity="0.22" />
            <stop offset="55%" stopColor="#3a0e22" stopOpacity="0.08" />
            <stop offset="100%" stopColor="#3a0e22" stopOpacity="0" />
          </radialGradient>
          {/* deep violet cloud */}
          <radialGradient id="hE-neb-violet" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor="#3d1e6e" stopOpacity="0.32" />
            <stop offset="100%" stopColor="#3d1e6e" stopOpacity="0" />
          </radialGradient>
          {/* cold ink-blue cloud — distant arm */}
          <radialGradient id="hE-neb-blue" cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor="#162a6e" stopOpacity="0.30" />
            <stop offset="100%" stopColor="#162a6e" stopOpacity="0" />
          </radialGradient>
          {/* milky way — quieter, dustier */}
          <linearGradient id="hE-milky" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stopColor="#cbb89a" stopOpacity="0" />
            <stop offset="45%" stopColor="#cbb89a" stopOpacity="0.04" />
            <stop offset="60%" stopColor="#9a7fb8" stopOpacity="0.05" />
            <stop offset="100%" stopColor="#cbb89a" stopOpacity="0" />
          </linearGradient>
          {/* soften the milky-way band */}
          <filter id="hE-soft" x="-20%" y="-20%" width="140%" height="140%">
            <feGaussianBlur stdDeviation="14" />
          </filter>
        </defs>

        {/* nebula clouds — three colored radial blooms layered to read cosmic */}
        <ellipse cx="980" cy="540" rx="640" ry="380" fill="url(#hE-neb-rust)" />
        <ellipse cx="320" cy="420" rx="540" ry="360" fill="url(#hE-neb-violet)" />
        <ellipse cx="700" cy="220" rx="520" ry="260" fill="url(#hE-neb-blue)" />

        {/* milky-way diagonal — soft luminous band running corner to corner */}
        <g filter="url(#hE-soft)" opacity="0.85">
          <path d="M -100,800 L 1500,80 L 1500,180 L -100,900 Z" fill="url(#hE-milky)" />
        </g>

        {/* dense star field — dimmer, sparser feeling so the dark reads first */}
        {HERO_E_STARS.map((s, i) => (
          <circle key={i} cx={s.x} cy={s.y} r={s.r} fill="#e6d5b3" opacity={s.o * 0.65} />
        ))}

        {/* a few "headline" stars — dim, no halo, just slightly larger */}
        {[
          [240, 180, 2.4], [820, 130, 2.2], [1080, 460, 2.0],
          [560, 380, 1.8], [380, 620, 2.2], [1240, 360, 1.8],
        ].map(([x, y, r], i) => (
          <g key={`b${i}`}>
            <circle cx={x} cy={y} r={r * 2.4} fill="#e6d5b3" opacity="0.05" />
            <circle cx={x} cy={y} r={r} fill="#e6d5b3" opacity="0.85" />
          </g>
        ))}

        {/* ringed planet — far left, dimmer & more shadowed */}
        <g transform="translate(160, 470) rotate(-18)">
          <ellipse cx="0" cy="0" rx="74" ry="16" fill="none" stroke="#cbb89a" strokeOpacity="0.40" strokeWidth="0.8" />
          <ellipse cx="0" cy="0" rx="58" ry="12" fill="none" stroke="#cbb89a" strokeOpacity="0.20" strokeWidth="0.5" />
          <circle cx="0" cy="0" r="24" fill="#d6c39a" opacity="0.85" />
          <circle cx="9" cy="-6" r="22" fill="#04040f" opacity="0.50" />
        </g>
      </svg>

      {/* Sangre de Cristo silhouette — three layered ridges, back to front,
          back ridge faintly lit by the moon, front ridge near-black.        */}
      <svg className="heroE-mountain" viewBox="0 0 1400 260" preserveAspectRatio="none" aria-hidden="true">
        {/* far ridge — softest, slightly tinted by horizon glow */}
        <path
          d="M0,260 L0,170 L60,150 L130,158 L210,128 L300,140 L400,108 L500,130 L600,98 L720,124 L820,100 L920,128 L1040,108 L1140,134 L1240,116 L1320,138 L1400,124 L1400,260 Z"
          fill="#2a1224" opacity="0.85"
        />
        {/* mid ridge — Sangre de Cristo jagged profile */}
        <path
          d="M0,260 L0,200 L40,178 L100,188 L160,156 L230,182 L300,148 L360,176 L440,128 L520,166 L590,140 L660,170 L740,134 L820,168 L900,142 L980,178 L1060,150 L1140,184 L1220,156 L1300,186 L1400,162 L1400,260 Z"
          fill="#170818"
        />
        {/* front ridge — near-black, the foreground horizon */}
        <path
          d="M0,260 L0,228 L80,218 L170,236 L260,212 L360,232 L460,210 L560,234 L660,214 L760,238 L860,216 L960,240 L1060,220 L1180,242 L1280,224 L1400,236 L1400,260 Z"
          fill="#070310"
        />
      </svg>

      <MockNav />

      {/* main type composition — vertical journey: night → horizon */}
      <div className="heroE-stage">

        {/* PAIR 1 — mood line leads into the names */}
        <div className="heroE-zone heroE-zone-intro">
          <span className="heroE-mood">Two stars aligning in the vastness of space</span>
        </div>

        <div className="heroE-zone heroE-zone-names">
          <div className="heroE-lockup heroE-lockup-who">
            <div className="heroE-row heroE-row-top">
              <span className="heroE-word heroE-word-outline">ADRIEN</span>
            </div>
            <div className="heroE-row heroE-row-mid">
              <span className="heroE-amp">&amp;</span>
            </div>
            <div className="heroE-row heroE-row-bot">
              <span className="heroE-word heroE-word-outline">MICHAEL</span>
            </div>
          </div>
        </div>

        {/* DIVIDER 1 — the moon itself breaks pair 1 from pair 2.
            Sits between the names and the "Join them..." line so it
            doesn't overlap any type. */}
        <div className="heroE-divider heroE-divider-moon" aria-hidden="true">
          <svg className="heroE-moon" viewBox="-200 -200 400 400">
            <defs>
              <radialGradient id="hE-moon-halo" cx="50%" cy="50%" r="50%">
                <stop offset="0%"  stopColor="#e8d3a8" stopOpacity="0.22" />
                <stop offset="55%" stopColor="#c8a878" stopOpacity="0.06" />
                <stop offset="100%" stopColor="#c8a878" stopOpacity="0" />
              </radialGradient>
              <mask id="hE-moon-mask">
                <rect x="-200" y="-200" width="400" height="400" fill="black" />
                <circle cx="0" cy="0" r="150" fill="white" />
                <circle cx="55" cy="-20" r="138" fill="black" />
              </mask>
            </defs>
            <circle cx="0" cy="0" r="190" fill="url(#hE-moon-halo)" />
            <rect x="-200" y="-200" width="400" height="400" fill="#d8c39a" opacity="0.85" mask="url(#hE-moon-mask)" />
          </svg>
        </div>

        {/* PAIR 2 — mood line leads into the date */}
        <div className="heroE-zone heroE-zone-invite">
          <span className="heroE-mood">Join them under the high desert moon</span>
        </div>

        <div className="heroE-zone heroE-zone-when">
          <div className="heroE-when">
            <span className="heroE-when-day">Friday</span>
            <span className="heroE-when-date">September 25, 2026</span>
          </div>
        </div>

        {/* DIVIDER — between date and the schedule */}
        <div className="heroE-divider" aria-hidden="true">
          <span className="heroE-divider-rule" />
          <svg className="heroE-divider-star" viewBox="-10 -10 20 20" width="12" height="12">
            <polygon points="0,-9 1.6,-1.6 9,0 1.6,1.6 0,9 -1.6,1.6 -9,0 -1.6,-1.6" fill="#e6d5b3" />
          </svg>
          <span className="heroE-divider-rule" />
        </div>

        {/* PAIR 3 — location & time, the practical beat */}
        <div className="heroE-zone heroE-zone-schedule">
          <div className="heroE-place">Santa Fe · New Mexico</div>
          <div className="heroE-schedule">
            <div className="heroE-sched-col">
              <span className="heroE-sched-time">5pm</span>
              <span className="heroE-sched-label">Ceremony</span>
              <span className="heroE-sched-detail">New Mexico<br/>Museum of Art</span>
            </div>
            <div className="heroE-sched-divider" aria-hidden="true" />
            <div className="heroE-sched-col">
              <span className="heroE-sched-time">7pm</span>
              <span className="heroE-sched-label">Reception</span>
              <span className="heroE-sched-detail">Meow Wolf</span>
            </div>
          </div>
        </div>

      </div>

      {/* Note — anchored at the very bottom, like a footnote */}
      <span className="heroE-note">( Note: This is a 21+ journey )</span>
    </header>
  );
}

window.HeroSynthesis = HeroSynthesis;
