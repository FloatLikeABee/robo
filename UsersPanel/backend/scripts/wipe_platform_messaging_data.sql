-- One-time / dev: remove all in-system messaging rows (threads, messages, reads, deliveries).
-- Run against the same database as UsersPanel. Order respects FK dependencies.

SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE msg_message_reads;
TRUNCATE TABLE msg_deliveries;
TRUNCATE TABLE msg_messages;
TRUNCATE TABLE msg_thread_members;
TRUNCATE TABLE msg_threads;
SET FOREIGN_KEY_CHECKS = 1;
