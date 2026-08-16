import { useState, type ComponentType, type ReactNode, type SVGProps } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../store/auth';
import { PlatformAssistantDrawer } from './PlatformAssistantDrawer';

type IconProps = SVGProps<SVGSVGElement> & { className?: string };

function SvgIcon({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden
    >
      {children}
    </svg>
  );
}

function IconDataAI({ className }: IconProps) {
  return (
    <SvgIcon className={className}>
      <path d="M12 3v3" />
      <path d="M12 18v3" />
      <circle cx="12" cy="12" r="4" />
      <path d="M5.5 7.5 7.5 9.5" />
      <path d="M16.5 14.5 18.5 16.5" />
      <path d="M18.5 7.5 16.5 9.5" />
      <path d="M7.5 14.5 5.5 16.5" />
    </SvgIcon>
  );
}

function IconBookings({ className }: IconProps) {
  return (
    <SvgIcon className={className}>
      <rect x="3" y="5" width="18" height="16" rx="2" />
      <path d="M3 10h18" />
      <path d="M8 3v4" />
      <path d="M16 3v4" />
    </SvgIcon>
  );
}

function IconFlowLog({ className }: IconProps) {
  return (
    <SvgIcon className={className}>
      <path d="M4 6h16" />
      <path d="M4 12h10" />
      <path d="M4 18h7" />
      <path d="M17 15l3 3-3 3" />
    </SvgIcon>
  );
}

const nav: { to: string; label: string; Icon: ComponentType<IconProps> }[] = [
  { to: '/', label: 'Booki AI', Icon: IconDataAI },
  { to: '/bookings', label: 'Bookings', Icon: IconBookings },
  { to: '/flow-log', label: 'Flow log', Icon: IconFlowLog },
];

const tabClass = ({ isActive }: { isActive: boolean }) =>
  `shrink-0 inline-flex items-center gap-2 rounded-xl px-3 py-1.5 text-sm font-medium transition-colors ${
    isActive
      ? 'bg-violet/30 text-text'
      : 'text-muted hover:text-text hover:bg-white/5'
  }`;

export function AppLayout() {
  const logout = useAuth((s) => s.logout);
  const navigate = useNavigate();
  const [assistantOpen, setAssistantOpen] = useState(false);

  return (
    <div className="flex-1 min-h-0 flex flex-col overflow-hidden bg-bg text-text">
      <header className="shrink-0 border-b border-white/5 bg-surface/30">
        <div className="flex items-center justify-between gap-3 px-4 py-3 md:px-8 flex-wrap">
          <div className="flex items-center gap-2.5 min-w-0">
            <img
              src="/booki-logo.svg"
              alt=""
              width={32}
              height={32}
              className="h-8 w-8 shrink-0 rounded-xl object-cover shadow-[0_0_20px_rgba(91,63,214,0.35)]"
            />
            <div className="min-w-0">
              <div className="font-semibold text-text leading-tight">Booki</div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              aria-expanded={assistantOpen}
              aria-controls="booki-ai-assistant-drawer"
              onClick={() => setAssistantOpen((o) => !o)}
              className="px-3 py-1.5 rounded-xl text-xs font-medium bg-violet/30 text-text hover:bg-violet/40"
            >
              AI Assistant
            </button>
            <button
              type="button"
              onClick={() => {
                logout();
                navigate('/login');
              }}
              className="px-3 py-1.5 rounded-xl text-xs font-medium text-muted hover:bg-white/5"
            >
              Sign out
            </button>
          </div>
        </div>

        <nav className="flex gap-1 px-4 pb-3 md:px-8 overflow-x-auto" aria-label="Sections">
          {nav.map(({ to, label, Icon }) => (
            <NavLink key={to} to={to} end={to === '/'} className={tabClass}>
              <Icon className="h-4 w-4 shrink-0" />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
      </header>

      <main className="flex-1 min-h-0 flex flex-col overflow-hidden p-4 md:p-8">
        <div className="flex-1 min-h-0 overflow-y-auto overflow-x-hidden">
          <Outlet />
        </div>
      </main>

      <PlatformAssistantDrawer open={assistantOpen} onClose={() => setAssistantOpen(false)} />
    </div>
  );
}
