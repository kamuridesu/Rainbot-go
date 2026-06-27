CREATE TABLE IF NOT EXISTS messages (
    stanzaId VARCHAR(255) NOT NULL,
    chatId VARCHAR(255) NOT NULL,
    senderJid VARCHAR(255) NOT NULL,
    messageText TEXT,
    quotedStanzaId VARCHAR(255),
    createdAt TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (stanzaId, createdAt)
) PARTITION BY RANGE (createdAt);

CREATE TABLE IF NOT EXISTS quotly (
    chatId VARCHAR(255) NOT NULL,
    fileId VARCHAR(255) NOT NULL,
    PRIMARY KEY (chatId, fileId),
    FOREIGN KEY(chatId) REFERENCES chat(chatId)
);

