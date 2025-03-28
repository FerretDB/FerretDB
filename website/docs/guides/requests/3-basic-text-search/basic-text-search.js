db.runCommand({ find: "books", filter: { $text: { $search: "drama" } } });
