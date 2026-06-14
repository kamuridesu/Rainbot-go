ALTER TABLE filter DROP CONSTRAINT filter_pkey;
ALTER TABLE filter ADD PRIMARY KEY (chatId, pattern);
