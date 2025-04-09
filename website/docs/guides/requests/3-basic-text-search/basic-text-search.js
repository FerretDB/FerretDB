db.runCommand({ find: 'books', filter: { $text: { $search: 'drama' } }, $db: '{{.Database}}' })
