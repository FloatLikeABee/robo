import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { ADMIN_BASE_PATH } from './adminPaths';

const PREFIX = process.env.PUBLIC_URL || '';

const AI = {
  title: 'Morph AI',
  iconSvg: `${PREFIX}/icons/morph-ai-icon.svg`,
  iconPng: `${PREFIX}/icons/morph-ai-192.png`,
  themeColor: '#060a12',
};

const DATA = {
  title: 'Morph Data',
  iconSvg: `${PREFIX}/icons/morph-data-icon.svg`,
  iconPng: `${PREFIX}/icons/morph-data-192.png`,
  themeColor: '#0a0413',
};

function setHref(selector, href) {
  const el = document.querySelector(selector);
  if (el) el.href = href;
}

/**
 * Tab / PWA-ish chrome: favicon and title follow Morph AI vs Morph Data surfaces.
 */
export default function DocumentBranding() {
  const { pathname } = useLocation();

  useEffect(() => {
    const isMorphData =
      pathname.startsWith(`${ADMIN_BASE_PATH}/`) ||
      pathname === ADMIN_BASE_PATH ||
      pathname.startsWith('/forms/') ||
      pathname === '/forms';

    const isLogin = pathname === '/login';
    const brand = isMorphData ? DATA : AI;

    document.title = isLogin ? 'Sign in · Morph AI & Morph Data' : brand.title;

    setHref('link#morph-favicon-svg', brand.iconSvg);
    setHref('link#morph-apple-touch', brand.iconPng);

    const themeMeta = document.querySelector('meta[name="theme-color"]');
    if (themeMeta) themeMeta.setAttribute('content', isLogin ? AI.themeColor : brand.themeColor);
  }, [pathname]);

  return null;
}
