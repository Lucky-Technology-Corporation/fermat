db = db.getSiblingDB('main');

db.createCollection("_swizzle_init_collection");
db.init_collection.insert({ initialized: true });
