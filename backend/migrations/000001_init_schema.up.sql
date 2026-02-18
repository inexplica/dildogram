-- Миграция 000001: Инициализация схемы базы данных

-- Включаем расширение для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- NULL для SMS авторизации
    first_name VARCHAR(50) NOT NULL DEFAULT '',
    last_name VARCHAR(50) NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    avatar_url VARCHAR(500) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_online BOOLEAN NOT NULL DEFAULT false,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Индексы для пользователей
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_is_online ON users(is_online);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Таблица SMS кодов (для имитации SMS авторизации)
CREATE TABLE sms_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(6) NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sms_codes_phone ON sms_codes(phone);
CREATE INDEX idx_sms_codes_code ON sms_codes(code);
CREATE INDEX idx_sms_codes_expires_at ON sms_codes(expires_at);

-- Таблица чатов (личных и групповых)
CREATE TABLE chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('private', 'group')),
    name VARCHAR(100) NOT NULL DEFAULT '', -- Для групповых чатов
    description TEXT NOT NULL DEFAULT '',
    avatar_url VARCHAR(500) NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Индексы для чатов
CREATE INDEX idx_chats_type ON chats(type);
CREATE INDEX idx_chats_created_by ON chats(created_by);
CREATE INDEX idx_chats_last_message_at ON chats(last_message_at);

-- Таблица участников чатов
CREATE TABLE chat_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(chat_id, user_id)
);

-- Индексы для участников
CREATE INDEX idx_chat_members_chat_id ON chat_members(chat_id);
CREATE INDEX idx_chat_members_user_id ON chat_members(user_id);
CREATE INDEX idx_chat_members_role ON chat_members(role);

-- Таблица сообщений
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'voice')),
    media_url VARCHAR(500),
    reply_to_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    is_edited BOOLEAN NOT NULL DEFAULT false,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'sent' CHECK (status IN ('pending', 'sent', 'delivered', 'read')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Индексы для сообщений
CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_status ON messages(status);
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);

-- Таблица статусов прочтения сообщений
CREATE TABLE message_reads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(message_id, user_id)
);

-- Индексы для прочтений
CREATE INDEX idx_message_reads_message_id ON message_reads(message_id);
CREATE INDEX idx_message_reads_user_id ON message_reads(user_id);

-- Таблица сессий пользователей (опционально)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_info VARCHAR(255) NOT NULL DEFAULT '',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token_hash);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Функция для обновления last_message_at в чате
CREATE OR REPLACE FUNCTION update_chat_last_message()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND NOT NEW.is_deleted) THEN
        UPDATE chats SET last_message_at = NEW.created_at WHERE id = NEW.chat_id;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_chat_last_message_trigger
    AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION update_chat_last_message();

-- Представление для последнего сообщения в чате
CREATE OR REPLACE VIEW chat_last_messages AS
SELECT DISTINCT ON (chat_id)
    chat_id,
    id as message_id,
    content,
    sender_id,
    created_at,
    status
FROM messages
WHERE is_deleted = false
ORDER BY chat_id, created_at DESC;

-- Начальные данные для тестирования (опционально)
-- INSERT INTO users (phone, username, first_name, last_name, bio) VALUES
--     ('+79001234567', 'alice', 'Alice', 'Smith', 'Hello, I am Alice!'),
--     ('+79007654321', 'bob', 'Bob', 'Johnson', 'Bob here');
