ALTER TABLE chat add disabledCategories TEXT DEFAULT '';
UPDATE chat SET disabledCategories =
    (CASE WHEN allowFun = 0 THEN 'diversão,' ELSE '' END) ||
    (CASE WHEN allowGames = 0 THEN 'rucoy,' ELSE '' END);
