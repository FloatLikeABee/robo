import { useEffect } from 'react';
import { Link, NavLink, Outlet } from 'react-router-dom';

type Theme = 'light' | 'dark';

function readTheme(): Theme {
  try {
    window.localStorage.setItem('sheetx-theme', 'dark');
  } catch {
    /* ignore */
  }
  return 'dark';
}

const tabClass = (isDark: boolean) => ({ isActive }: { isActive: boolean }) =>
  `shrink-0 inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
    isActive
      ? isDark
        ? 'bg-sky-700/35 text-sky-100 border border-sky-500/40'
        : 'bg-[#d8eef8] text-[#0f4c66] border border-[#b7deee]'
      : isDark
        ? 'text-slate-200 hover:bg-slate-700/70 hover:text-white border border-transparent'
        : 'text-[#31566b] hover:bg-[#e8f5fb] hover:text-[#123f56] border border-transparent'
  }`;

export function Layout() {
  const theme: Theme = readTheme();

  useEffect(() => {
    window.localStorage.setItem('sheetx-theme', theme);
  }, [theme]);

  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);

  const isDark = theme === 'dark';
  const tc = tabClass(isDark);

  return (
    <div
      className={
        isDark
          ? 'min-h-dvh md:h-screen overflow-hidden bg-linear-to-b from-[#14213d] via-[#0f172a] to-[#0b1120] text-slate-100 flex flex-col'
          : 'min-h-dvh md:h-screen overflow-hidden bg-[#e8e4dc] text-slate-900 flex flex-col'
      }
    >
      <header
        className={
          'shrink-0 z-20 border-b ' +
          (isDark ? 'border-slate-700/80 bg-slate-800/95' : 'border-[#c6e3ef] bg-[#eaf6fb]/95')
        }
      >
        <div className="px-3 md:px-4 py-3 flex items-center justify-between gap-2">
          <Link to="/" className="flex items-center gap-2 min-w-0">
            <span className={isDark ? 'text-2xl text-violet-300' : 'text-2xl text-violet-600'}>📄</span>
            <div className="min-w-0">
              <span className={isDark ? 'font-semibold text-white text-lg' : 'font-semibold text-slate-900 text-lg'}>
                Event Logs
              </span>
            </div>
          </Link>
        </div>

        <nav
          className="flex items-center gap-1 px-3 md:px-4 pb-3 overflow-x-auto"
          aria-label="Sections"
        >
          <NavLink to="/events-info" className={tc}>
            <span aria-hidden className="text-base leading-none">📌</span>
            <span>Events &amp; Info</span>
          </NavLink>
        </nav>
      </header>

      <main
        className={
          'flex-1 min-w-0 min-h-0 flex flex-col overflow-hidden px-2 sm:px-3 md:px-4 py-2 md:py-4 ' +
          (isDark ? '' : 'bg-[#f6fbfe]')
        }
      >
        <div className="flex-1 min-h-0 flex flex-col overflow-hidden">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
