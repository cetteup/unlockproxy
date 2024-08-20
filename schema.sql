CREATE TABLE `providers`
(
    `id`   int(1) NOT NULL,
    `name` varchar(10) NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `players`
(
    `pid`      int(11) NOT NULL,
    `nick`     varchar(50) NOT NULL,
    `provider` int(1) NOT NULL,
    `imported` datetime    NOT NULL,
    PRIMARY KEY (`pid`, `provider`),
    KEY        `players_providers_FK` (`provider`),
    CONSTRAINT `players_providers_FK` FOREIGN KEY (`provider`) REFERENCES `providers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;