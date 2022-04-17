

CREATE DATABASE IF NOT EXISTS grafana;

use grafana;

CREATE USER 'grafana'@'%' IDENTIFIED BY 'grafana';
GRANT SELECT  ON grafana. * TO 'grafana'@'%';

CREATE TABLE `fffffflllllllllaaaagggggg` (
  `flag` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO fffffflllllllllaaaagggggg(flag) VALUES("*ctf{Upgrade_your_grafAna_now!}");



ALTER TABLE users CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;
ALTER TABLE notes CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci;