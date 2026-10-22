import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { ADMIN_BASE_PATH } from './adminPaths';

const PREFIX = process.env.PUBLIC_URL || '';

const AI = {
  title: 'Dorph AI',
  iconSvg: `${PREFIX}/icons/dorph-ai-icon.svg`,
  iconPng: `${PREFIX}/icons/dorph-ai-192.png`,
  themeColor: '#060a12',
};

const DATA = {
  title: 'DorphNotes',
  iconSvg: `${PREFIX}/icons/dorph-data-icon.svg`,
  iconPng: `${PREFIX}/icons/dorph-data-192.png`,
  themeColor: '#0a0413',
};

function setHref(selector, href) {
  const el = document.querySelector(selector);
  if (el) el.href = href;
}

/**
 * Tab / PWA-ish chrome: favicon and title follow Dorph AI vs Dorph Data surfaces.
 */
export default function DocumentBranding() {
  const { pathname } = useLocation();

  useEffect(() => {
    const isDorphData =
      pathname.startsWith(`${ADMIN_BASE_PATH}/`) ||
      pathname === ADMIN_BASE_PATH ||
      pathname.startsWith('/forms/') ||
      pathname === '/forms';

    const isLogin = pathname === '/login';
    const brand = isDorphData ? DATA : AI;

    document.title = isLogin ? 'Sign in · Dorph AI & DorphNotes' : brand.title;

    setHref('link#dorph-favicon-svg', brand.iconSvg);
    setHref('link#dorph-apple-touch', brand.iconPng);

    const themeMeta = document.querySelector('meta[name="theme-color"]');
    if (themeMeta) themeMeta.setAttribute('content', isLogin ? AI.themeColor : brand.themeColor);
  }, [pathname]);

  return null;
}
