# Sample contact data

| File | Use |
|------|-----|
| `contacts_mock.csv` | Contacts → **Import & sync** → CSV import |
| `contacts_mock.sql` | `mysql … < fixtures/contacts_mock.sql` (idempotent upsert by email) |
| `contacts_mock.json` | Serve over HTTP and use **API sync** with that URL, or paste into a mock API |

Emails are `@example.com` placeholders only.

## Built-in email templates JSON

`../data/builtin_email_templates.json` is generated from `frontend/src/lib/starterTemplates.js` (tags + descriptions for AI). From the repo root:

```bash
node <<'NODE'
const fs = require('fs');
const code = fs.readFileSync('frontend/src/lib/starterTemplates.js', 'utf8')
  .replace(/export const STARTER_TEMPLATES/, 'const STARTER_TEMPLATES')
  .replace(/export function[\s\S]*/, '');
eval(code + '; globalThis.__ST = STARTER_TEMPLATES');
const STARTER_TEMPLATES = globalThis.__ST;
const tagById = {
  'account-balance': 'billing.balance',
  'welcome': 'onboarding.welcome',
  'newsletter': 'marketing.newsletter',
  'invoice': 'billing.invoice',
  'appointment': 'scheduling.appointment',
  'promo': 'marketing.promo',
  'order-confirmation': 'commerce.order-confirm',
  'course-announcement': 'education.course',
  'assignment-reminder': 'education.assignment',
  'password-security': 'security.account'
};
const desc = (t) =>
  `${t.description} Merge placeholders use {{curly}} tokens typical for mail merge (e.g. name, company, dates, amounts).`;
const out = STARTER_TEMPLATES.map((t) => ({
  builtin_key: t.id,
  name: t.suggestedName,
  tag: tagById[t.id] || `general.${t.id}`,
  description: desc(t),
  html: t.html
}));
fs.mkdirSync('backend/data', { recursive: true });
fs.writeFileSync('backend/data/builtin_email_templates.json', JSON.stringify(out));
console.log('wrote', out.length, 'entries');
NODE
```
