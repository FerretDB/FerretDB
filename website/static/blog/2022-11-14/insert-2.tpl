db.league.insertOne({
    club: "PSG",
    points: 30,
    average_age: 30,
    discipline: { red: 5, yellow: 30 },
    qualified: false,
});
db.league.insertMany([
    {
        club: "Arsenal",
        points: 80,
        average_age: 24,
        discipline: { red: 2, yellow: 15 },
        qualified: true,
    },
    {
        club: "Barcelona",
        points: 60,
        average_age: 31,
        discipline: { red: 0, yellow: 7 },
        qualified: false,
    },
    {
        club: "Bayern",
        points: 84,
        average_age: 29,
        discipline: { red: 1, yellow: 20 },
        qualified: true,
    },
]);

##CODE##
