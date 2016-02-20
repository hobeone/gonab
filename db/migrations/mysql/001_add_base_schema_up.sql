/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `binary` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `hash` varchar(16) DEFAULT NULL,
  `name` varchar(512) DEFAULT NULL,
  `total_parts` int(11) DEFAULT NULL,
  `posted` timestamp NULL DEFAULT NULL,
  `from` varchar(255) DEFAULT NULL,
  `xref` varchar(1024) DEFAULT NULL,
  `group_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_binary_name` (`name`),
  KEY `idx_binary_hash` (`hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `group` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `active` tinyint(1) DEFAULT NULL,
  `first` bigint(20) DEFAULT NULL,
  `last` bigint(20) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `idx_group_active` (`active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `missed_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `message_number` bigint(20) DEFAULT NULL,
  `group_name` varchar(255) DEFAULT NULL,
  `attempts` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_missed_message_group_name` (`group_name`),
  KEY `idx_missed_message_message_number` (`message_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `part` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `hash` varchar(16) DEFAULT NULL,
  `subject` varchar(512) DEFAULT NULL,
  `total_segments` int(11) DEFAULT NULL,
  `posted` timestamp NULL DEFAULT NULL,
  `from` varchar(255) DEFAULT NULL,
  `xref` varchar(1024) DEFAULT NULL,
  `group_name` varchar(255) DEFAULT NULL,
  `binary_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_part_hash` (`hash`),
  KEY `idx_part_total_segments` (`total_segments`),
  KEY `idx_part_group_name` (`group_name`),
  KEY `idx_part_binary_id` (`binary_id`),
  KEY `idx_part_posted` (`posted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `regex` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `regex` varchar(2048) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `status` tinyint(1) DEFAULT NULL,
  `ordinal` int(11) DEFAULT NULL,
  `group_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `release` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `hash` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `posted` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `search_name` varchar(255) DEFAULT NULL,
  `original_name` varchar(255) DEFAULT NULL,
  `from` varchar(255) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `grabs` int(11) DEFAULT NULL,
  `size` bigint(20) DEFAULT NULL,
  `group_id` bigint(20) DEFAULT NULL,
  `category_id` bigint(20) DEFAULT NULL,
  `nzb` longtext,
  KEY `idx_release_group_id` (`group_id`),
  KEY `idx_release_category_id` (`category_id`),
  KEY `idx_release_search_name` (`search_name`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `segment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `segment` int(11) DEFAULT NULL,
  `size` bigint(20) DEFAULT NULL,
  `message_id` varchar(255) DEFAULT NULL,
  `part_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_segment_segment` (`segment`),
  KEY `idx_segment_part_id` (`part_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC;
/*!40101 SET character_set_client = @saved_cs_client */;
