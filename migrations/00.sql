CREATE TABLE IF NOT EXISTS chat (
    chatId TEXT PRIMARY KEY NOT NULL,
    isBotEnabled INTEGER DEFAULT 1,
    prefix VARCHAR(1) DEFAULT '/',
    adminOnly INTEGER DEFAULT 0,
    profanityFilterEnabled INTEGER DEFAULT 0,
    customProfanityWords TEXT DEFAULT '',
    warnBanThreshold INTEGER DEFAULT 4,
    allowAdults INTEGER DEFAULT 0,
    allowGames INTEGER DEFAULT 1,
    allowFun INTEGER DEFAULT 1,
    welcomeMessage TEXT DEFAULT '',
    countMessages INTEGER DEFAULT 1
);

CREATE TABLE IF NOT EXISTS member (
    jid TEXT NOT NULL,
    chatId TEXT NOT NULL,
    warns INTEGER DEFAULT 0,
    points INTEGER DEFAULT 0,
    messages INTEGER DEFAULT 0,
    silenced INTEGER DEFAULT 0,
    FOREIGN KEY(chatId) REFERENCES chat(chatId)
);

CREATE TABLE IF NOT EXISTS filter (
    chatId TEXT NOT NULL,
    pattern TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    response TEXT NOT NULL,
    FOREIGN KEY(chatId) REFERENCES chat(chatId)
);
