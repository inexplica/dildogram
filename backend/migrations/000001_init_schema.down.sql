-- Откат миграции 000001: Удаление всех таблиц

DROP VIEW IF EXISTS chat_last_messages;
DROP TRIGGER IF EXISTS update_chat_last_message_trigger ON messages;
DROP FUNCTION IF EXISTS update_chat_last_message;
DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;
DROP TRIGGER IF EXISTS update_chats_updated_at ON chats;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column;

DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS message_reads CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
DROP TABLE IF EXISTS chat_members CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS sms_codes CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP EXTENSION IF EXISTS "uuid-ossp";
