-- name: InsertFile :exec
INSERT INTO files (inode,name,parent,data) VALUES (?,?,?,?);

-- name: SelectOneFileName :one
SELECT * FROM files WHERE name = ? LIMIT 1;

-- name: SelectOneFileInode :one
SELECT * FROM files WHERE inode = ? LIMIT 1;

-- name: SelectFilesParent :many
SELECT * FROM files WHERE parent = ?;

-- name: UpdateFile :exec
UPDATE files SET inode = ?, name = ?, parent = ?, data = ? WHERE inode = ?;

-- name: DeleteFileInode :exec
DELETE FROM files WHERE inode = ?;

-- name: DeleteFileName :exec
DELETE FROM files WHERE name = ?;

-- name: DeleteFileParent :exec
DELETE FROM files WHERE parent = ?;

-- name: InsertDirectory :exec
INSERT INTO directories (inode,name,parent) VALUES (?,?,?);

-- name: SelectOneDirectoryName :one
SELECT * FROM directories WHERE name = ? LIMIT 1;

-- name: SelectOneDirectoryInode :one
SELECT * FROM directories WHERE name = ? LIMIT 1;

-- name: SelectDirectoriesParent :many
SELECT * FROM directories WHERE parent = ?;

-- name: UpdateDirectory :exec
UPDATE directories SET inode = ?, name = ?, parent = ? WHERE inode = ?;

-- name: DeleteDirectoryInode :exec
DELETE FROM directories WHERE inode = ?;

-- name: DeleteDirectoryName :exec
DELETE FROM directories WHERE name = ?;

-- name: DeleteDirectoryParent :exec
DELETE FROM directories WHERE parent = ?;
