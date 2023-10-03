db = db.getSiblingDB('main');

db.createCollection("init_collection");
db.init_collection.insert({ initialized: true });
