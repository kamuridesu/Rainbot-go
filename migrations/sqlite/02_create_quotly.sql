CREATE TABLE IF NOT EXISTS quotly (
    chatId TEXT NOT NULL,
    fileId TEXT NOT NULL,
    PRIMARY KEY (chatId, fileId),
    FOREIGN KEY(chatId) REFERENCES chat(chatId)
);
