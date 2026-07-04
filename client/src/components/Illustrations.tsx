// Small hand-drawn SVG accents — no icon library, keeps the dependency footprint at zero.
// Purely decorative (not data), fine to add freely per foundations §3.2 ("состояния не
// передаются только цветом" — these add a shape/icon cue alongside the text, not instead of it).

export function WaveDivider({ color = 'var(--bg)' }: { color?: string }) {
  return (
    <svg className="wave-divider" viewBox="0 0 1440 60" preserveAspectRatio="none" aria-hidden="true">
      <path
        d="M0,30 C240,60 480,0 720,28 C960,56 1200,4 1440,30 L1440,60 L0,60 Z"
        fill={color}
      />
    </svg>
  )
}

export function EmptyBowlIllustration() {
  return (
    <svg viewBox="0 0 120 120" width="96" height="96" fill="none" aria-hidden="true" className="state-illustration">
      <path d="M40 32 C36 24, 44 20, 40 10" stroke="currentColor" strokeWidth="4" strokeLinecap="round" opacity="0.45" />
      <path d="M60 32 C56 22, 64 18, 60 6" stroke="currentColor" strokeWidth="4" strokeLinecap="round" opacity="0.6" />
      <path d="M80 32 C76 24, 84 20, 80 10" stroke="currentColor" strokeWidth="4" strokeLinecap="round" opacity="0.45" />
      <path
        d="M18 52 C18 52 32 60 60 60 C88 60 102 52 102 52 C102 80 84 98 60 98 C36 98 18 80 18 52 Z"
        fill="currentColor"
      />
      <ellipse cx="60" cy="52" rx="42" ry="11" fill="currentColor" opacity="0.55" />
    </svg>
  )
}

export function AlertPotIllustration() {
  return (
    <svg viewBox="0 0 120 120" width="96" height="96" fill="none" aria-hidden="true" className="state-illustration">
      <circle cx="60" cy="60" r="50" fill="currentColor" opacity="0.12" />
      <rect x="53" y="32" width="14" height="42" rx="7" fill="currentColor" />
      <circle cx="60" cy="86" r="8" fill="currentColor" />
    </svg>
  )
}
