-- CREATE TABLE IF NOT EXISTS nodes(
--     inode UNSIGNED BIG INT PRIMARY KEY,
--     name TEXT NOT NULL
-- );

CREATE TABLE IF NOT EXISTS files(
    inode UNSIGNED BIG INT PRIMARY KEY,
    name TEXT NOT NULL,
    parent UNSIGNED BIG INT NOT NULL,
    data BLOB,
    CONSTRAINT fk_parent FOREIGN KEY(parent) REFERENCES dirs(inode)
);

CREATE INDEX IF NOT EXISTS idx_files_name ON files (name);

CREATE TABLE IF NOT EXISTS directories (
    inode UNSIGNED BIG INT PRIMARY KEY,
    name TEXT NOT NULL,
    parent UNSIGNED BIG INT NOT NULL,
    CONSTRAINT fk_parent FOREIGN KEY(parent) REFERENCES dirs(inode)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_directories_inodes ON directories (inode);

CREATE INDEX IF NOT EXISTS idx_directories_name ON directories (name);

INSERT INTO directories (inode,name,parent) VALUES (0,"root",0);