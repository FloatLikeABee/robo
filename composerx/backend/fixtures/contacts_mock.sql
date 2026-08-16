-- Mock contacts for local development (safe to re-run: upserts by unique email).
-- mysql -u USER -p DB_NAME < fixtures/contacts_mock.sql

INSERT INTO email_contacts (name, email, phone, note, created_at, updated_at) VALUES
  ('Alex Morgan', 'alex.morgan@example.com', '+1-415-555-0101', 'Newsletter — product updates', NOW(), NOW()),
  ('Jamie Rivera', 'jamie.rivera@example.com', '+1-212-555-0142', 'Enterprise pilot Q2', NOW(), NOW()),
  ('Sam Okonkwo', 'sam.okonkwo@example.com', '+44-20-7946-0958', 'London office lead', NOW(), NOW()),
  ('Priya Shah', 'priya.shah@example.com', '+91-22-5551-3300', 'Prefers plain-text summaries', NOW(), NOW()),
  ('Chris Nielsen', 'chris.nielsen@example.com', '+45-32-12-34-56', 'Copenhagen distributor', NOW(), NOW()),
  ('Morgan Lee', 'morgan.lee@example.com', '+1-604-555-0187', 'Weekly digest opt-in', NOW(), NOW()),
  ('Taylor Brooks', 'taylor.brooks@example.com', '+1-512-555-0163', 'Event invitations only', NOW(), NOW()),
  ('Jordan Hayes', 'jordan.hayes@example.com', '+1-617-555-0199', 'Renewal reminder Dec', NOW(), NOW()),
  ('Riley Chen', 'riley.chen@example.com', '+86-21-5550-8844', 'Shanghai partner contact', NOW(), NOW()),
  ('Casey Nguyen', 'casey.nguyen@example.com', '+1-714-555-0120', 'Wholesale pricing tier B', NOW(), NOW()),
  ('Quinn Foster', 'quinn.foster@example.com', '+353-1-555-0144', 'Dublin startup week follow-up', NOW(), NOW()),
  ('Avery Patel', 'avery.patel@example.com', '+1-647-555-0175', 'Toronto workshop attendee', NOW(), NOW()),
  ('Reese Walters', 'reese.walters@example.com', '+49-30-5551-9922', 'Berlin meetup 2025', NOW(), NOW()),
  ('Skyler Kim', 'skyler.kim@example.com', '+82-2-555-3300', 'Seoul reseller channel', NOW(), NOW()),
  ('Drew Martinez', 'drew.martinez@example.com', '+34-91-555-7700', 'Madrid EU compliance cc', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  phone = VALUES(phone),
  note = VALUES(note),
  updated_at = NOW();
