import { AcademiApp } from './app.js';

export function syncAppViewportHeight() {
    const viewportHeight = window.visualViewport?.height || window.innerHeight;
    document.documentElement.style.setProperty('--app-height', `${Math.round(viewportHeight)}px`);
    const isDesktop = window.matchMedia('(min-width: 768px)').matches;
    if (isDesktop) {
        const frameH = Math.min(820, Math.round(viewportHeight) - 32);
        document.documentElement.style.setProperty('--frame-height', `${frameH}px`);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    syncAppViewportHeight();
    new AcademiApp();
});

window.addEventListener('resize', syncAppViewportHeight);
window.addEventListener('orientationchange', syncAppViewportHeight);
if (window.visualViewport) {
    window.visualViewport.addEventListener('resize', syncAppViewportHeight);
}

if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        // navigator.serviceWorker.register('/sw.js');
    });
}

document.addEventListener('touchstart', () => {}, { passive: true });

document.addEventListener('touchmove', (e) => {
    if (e.target.closest('.messages-container, .community-feed, .docs-grid, .community-detail-scroll, .doc-modal-body, .community-analysis-preview, .profile-screen .settings-list, .chat-history-list')) {
        return;
    }
}, { passive: true });
