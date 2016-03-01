CREATE TABLE "binary" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "hash" varchar(16) DEFAULT NULL,
  "name" varchar(512) DEFAULT NULL,
  "total_parts" INTEGER DEFAULT NULL,
  "posted" timestamp NULL DEFAULT NULL,
  "from" varchar(255) DEFAULT NULL,
  "xref" varchar(1024) DEFAULT NULL,
  "group_name" varchar(255) DEFAULT NULL
);
CREATE TABLE "group" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "active" tinyint(1) DEFAULT NULL,
  "first" INTEGER DEFAULT NULL,
  "last" INTEGER DEFAULT NULL,
  "name" varchar(255) DEFAULT NULL
);
CREATE TABLE "missed_message" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "message_number" INTEGER DEFAULT NULL,
  "group_name" varchar(255) DEFAULT NULL,
  "attempts" INTEGER DEFAULT NULL
);
CREATE TABLE "part" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "hash" varchar(16) DEFAULT NULL,
  "subject" varchar(512) DEFAULT NULL,
  "total_segments" INTEGER DEFAULT NULL,
  "posted" timestamp NULL DEFAULT NULL,
  "from" varchar(255) DEFAULT NULL,
  "xref" varchar(1024) DEFAULT NULL,
  "group_name" varchar(255) DEFAULT NULL,
  "binary_id" INTEGER DEFAULT NULL
);
CREATE TABLE "regex" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "regex" varchar(2048) DEFAULT NULL,
  "description" varchar(255) DEFAULT NULL,
  "status" tinyint(1) DEFAULT NULL,
  "ordinal" INTEGER DEFAULT NULL,
  "group_regex" varchar(255) DEFAULT NULL,
  "kind" varchar(255) DEFAULT NULL
);
CREATE TABLE "collection_regex" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "regex" varchar(2048) DEFAULT NULL,
  "description" varchar(255) DEFAULT NULL,
  "status" tinyint(1) DEFAULT NULL,
  "ordinal" INTEGER DEFAULT NULL,
  "group_regex" varchar(255) DEFAULT NULL,
  "kind" varchar(255) DEFAULT NULL
);
CREATE TABLE "release" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "hash" varchar(255) DEFAULT NULL,
  "created_at" timestamp NULL DEFAULT NULL,
  "posted" timestamp NULL DEFAULT NULL,
  "name" varchar(255) DEFAULT NULL,
  "search_name" varchar(255) DEFAULT NULL,
  "original_name" varchar(255) DEFAULT NULL,
  "from" varchar(255) DEFAULT NULL,
  "status" INTEGER DEFAULT NULL,
  "grabs" INTEGER DEFAULT NULL,
  "size" INTEGER DEFAULT NULL,
  "group_id" INTEGER DEFAULT NULL,
  "category_id" INTEGER DEFAULT NULL,
  "nzb" longtext
);
CREATE TABLE "segment" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "segment" INTEGER DEFAULT NULL,
  "size" INTEGER DEFAULT NULL,
  "message_id" varchar(255) DEFAULT NULL,
  "part_id" INTEGER DEFAULT NULL
);
CREATE INDEX "segment_idx_segment_segment" ON "segment" ("segment");
CREATE INDEX "segment_idx_segment_part_id" ON "segment" ("part_id");
CREATE INDEX "group_name" ON "group" ("name");
CREATE INDEX "group_idx_group_active" ON "group" ("active");
CREATE INDEX "part_idx_part_hash" ON "part" ("hash");
CREATE INDEX "part_idx_part_total_segments" ON "part" ("total_segments");
CREATE INDEX "part_idx_part_group_name" ON "part" ("group_name");
CREATE INDEX "part_idx_part_binary_id" ON "part" ("binary_id");
CREATE INDEX "part_idx_part_posted" ON "part" ("posted");
CREATE INDEX "missed_message_idx_missed_message_group_name" ON "missed_message" ("group_name");
CREATE INDEX "missed_message_idx_missed_message_message_number" ON "missed_message" ("message_number");
CREATE INDEX "release_idx_release_group_id" ON "release" ("group_id");
CREATE INDEX "release_idx_release_category_id" ON "release" ("category_id");
CREATE INDEX "release_idx_release_search_name" ON "release" ("search_name");
CREATE INDEX "binary_idx_binary_name" ON "binary" ("name");
CREATE INDEX "binary_idx_binary_hash" ON "binary" ("hash");
