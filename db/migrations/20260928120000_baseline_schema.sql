-- The schema as it stood when migrations were introduced. Every statement is
-- IF NOT EXISTS so this is a no-op against the existing dev and production
-- databases, which already have these tables: it exists to give a fresh
-- database the same starting point rather than to change an established one.
--
-- MySQL implicitly commits on DDL, so a surrounding transaction would only
-- give the false impression that a partial failure rolls back.

-- migrate:up transaction:false

CREATE TABLE IF NOT EXISTS `Card_Colors` (
  `id` binary(16) NOT NULL,
  `card_id` binary(16) DEFAULT NULL,
  `color_id` tinyint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_colors_card_id` (`card_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Card_Flavor_Text` (
  `id` binary(16) NOT NULL,
  `text` text,
  `card_id` binary(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_flavor_text_card_id` (`card_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Card_Keywords` (
  `id` binary(16) NOT NULL,
  `keyword` varchar(255) DEFAULT NULL,
  `card_id` binary(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_keywords_card_id` (`card_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Card_Names` (
  `id` binary(16) NOT NULL,
  `card_id` binary(16) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_names_card_id` (`card_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Card_Prints` (
  `id` binary(16) NOT NULL,
  `released` int NOT NULL,
  `card_id` binary(16) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_prints_card_id` (`card_id`,`released`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Card_Subtypes` (
  `id` binary(16) NOT NULL,
  `subtype` varchar(255) DEFAULT NULL,
  `card_id` binary(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_subtypes_card_id` (`card_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Card_Texts` (
  `id` binary(16) NOT NULL,
  `text` text,
  `card_id` binary(16) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_card_texts_card_id` (`card_id`),
  FULLTEXT KEY `idx_text` (`text`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Cards` (
  `id` binary(16) NOT NULL,
  `layout` varchar(255) DEFAULT NULL,
  `front` varchar(255) DEFAULT NULL,
  `back` varchar(255) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `toughness` int DEFAULT NULL,
  `power` int DEFAULT NULL,
  `totalManaCost` int DEFAULT NULL,
  `art` varchar(255) DEFAULT NULL,
  `standard` tinyint(1) DEFAULT '0',
  `future` tinyint(1) DEFAULT '0',
  `historic` tinyint(1) DEFAULT '0',
  `gladiator` tinyint(1) DEFAULT '0',
  `pioneer` tinyint(1) DEFAULT '0',
  `explorer` tinyint(1) DEFAULT '0',
  `modern` tinyint(1) DEFAULT '0',
  `legacy` tinyint(1) DEFAULT '0',
  `pauper` tinyint(1) DEFAULT '0',
  `vintage` tinyint(1) DEFAULT '0',
  `penny` tinyint(1) DEFAULT '0',
  `commander` tinyint(1) DEFAULT '0',
  `oathbreaker` tinyint(1) DEFAULT '0',
  `brawl` tinyint(1) DEFAULT '0',
  `historicbrawl` tinyint(1) DEFAULT '0',
  `alchemy` tinyint(1) DEFAULT '0',
  `paupercommander` tinyint(1) DEFAULT '0',
  `duel` tinyint(1) DEFAULT '0',
  `oldschool` tinyint(1) DEFAULT '0',
  `premodern` tinyint(1) DEFAULT '0',
  `predh` tinyint(1) DEFAULT '0',
  `rarity` tinyint unsigned DEFAULT NULL,
  `manaCost` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `oracle_id` binary(16) DEFAULT NULL,
  `price` int DEFAULT NULL,
  `set_name` varchar(255) DEFAULT NULL,
  `edh_rank` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_cards_id_power` (`id`,`power`),
  KEY `idx_cards_id_toughness` (`id`,`toughness`),
  KEY `idx_cards_id` (`id`),
  KEY `idx_cards_name` (`name`),
  KEY `idx_edh_rank` (`edh_rank`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Colors` (
  `id` tinyint unsigned NOT NULL AUTO_INCREMENT,
  `color` varchar(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_colors_id` (`id`),
  KEY `idx_colors_color` (`color`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Deck_Cards` (
  `id` binary(16) NOT NULL,
  `card_id` binary(16) NOT NULL,
  `deck_id` binary(16) NOT NULL,
  `qty` tinyint unsigned DEFAULT '1',
  `dateCreated` datetime DEFAULT CURRENT_TIMESTAMP,
  `sideboard` tinyint(1) DEFAULT '0',
  `print` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Deck_Groups` (
  `id` binary(16) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `label` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Decks` (
  `id` binary(16) NOT NULL,
  `label` varchar(255) NOT NULL,
  `commander_card_id` binary(16) DEFAULT NULL,
  `user_id` varchar(255) NOT NULL,
  `oathbreaker_card_id` binary(16) DEFAULT NULL,
  `sleeve_id` binary(16) DEFAULT NULL,
  `deck_group_id` binary(16) DEFAULT NULL,
  `budget` int DEFAULT NULL,
  `gamemode` varchar(16) DEFAULT NULL,
  `partner_card_id` binary(16) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Rarities` (
  `id` tinyint unsigned NOT NULL AUTO_INCREMENT,
  `rarity` varchar(50) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Sessions` (
  `session_id` binary(16) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `data` blob NOT NULL,
  `expires` timestamp NOT NULL,
  PRIMARY KEY (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `Sleeves` (
  `id` binary(16) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `image_url` varchar(255) NOT NULL,
  `is_video` tinyint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- migrate:down

-- Deliberately empty. Reversing the baseline means dropping every table in the
-- application, which is never what a rollback should do here.
