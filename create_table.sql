CREATE DATABASE db_test;
USE db_test;
-- 创建表
CREATE TABLE `admins`
(
    `admin_id` BIGINT NOT NULL PRIMARY KEY,
    `password` TEXT,
    `username` VARCHAR(15) DEFAULT NULL
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

CREATE TABLE `class_infos`
(
    `class_id`   INT         NOT NULL PRIMARY KEY,
    `class_name` VARCHAR(15) NOT NULL
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;
ALTER TABLE `class_infos`
    MODIFY `class_id` INT NOT NULL AUTO_INCREMENT,
    AUTO_INCREMENT = 1;

CREATE TABLE `books`
(
    `book_id`      BIGINT         NOT NULL PRIMARY KEY,
    `name`         VARCHAR(20)    NOT NULL,
    `author`       VARCHAR(15)    NOT NULL,
    `publish`      VARCHAR(20)    NOT NULL,
    `ISBN`         VARCHAR(15)    NOT NULL,
    `introduction` TEXT,
    `language`     VARCHAR(4)     NOT NULL,
    `price`        DECIMAL(10, 2) NOT NULL,
    `pub_date`     VARCHAR(10)    NOT NULL,
    `class_id`     INT            NOT NULL,
    `number`       INT DEFAULT NULL,
    `image`        TEXT
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

ALTER TABLE `books`
    MODIFY `book_id` BIGINT NOT NULL AUTO_INCREMENT,
    AUTO_INCREMENT = 1;

CREATE TABLE `lends`
(
    `ser_num`   BIGINT NOT NULL PRIMARY KEY,
    `book_id`   BIGINT NOT NULL,
    `reader_id` BIGINT NOT NULL,
    `lend_date` DATE DEFAULT NULL,
    `back_date` DATE DEFAULT NULL
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

ALTER TABLE `lends`
    MODIFY `ser_num` BIGINT NOT NULL AUTO_INCREMENT,
    AUTO_INCREMENT = 1;

CREATE TABLE `reserves`
(
    `ser_num`      BIGINT NOT NULL PRIMARY KEY,
    `book_id`      BIGINT NOT NULL,
    `reader_id`    BIGINT NOT NULL,
    `require_date` DATE DEFAULT NULL,
    `accept_date`  DATE DEFAULT NULL
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

ALTER TABLE `reserves`
    MODIFY `ser_num` BIGINT NOT NULL AUTO_INCREMENT,
    AUTO_INCREMENT = 1;

CREATE TABLE `reader_infos`
(
    `reader_id` BIGINT      NOT NULL PRIMARY KEY,
    `name`      VARCHAR(10) NOT NULL,
    `sex`       VARCHAR(2)  NOT NULL,
    `birth`     DATE        NOT NULL,
    `address`   VARCHAR(50) NOT NULL,
    `phone`     VARCHAR(15) NOT NULL
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

ALTER TABLE `reader_infos`
    MODIFY `reader_id` BIGINT NOT NULL AUTO_INCREMENT,
    AUTO_INCREMENT = 10000;

CREATE TABLE `reader_cards`
(
    `reader_id` BIGINT      NOT NULL PRIMARY KEY,
    `username`  VARCHAR(15) NOT NULL,
    `password`  TEXT
) ENGINE = INNODB
  DEFAULT CHARSET = utf8;

-- 添加外键约束
ALTER TABLE `books`
    ADD CONSTRAINT `fk_class_infos_class_id`
        FOREIGN KEY (`class_id`) REFERENCES `class_infos` (`class_id`)
            ON DELETE CASCADE
            ON UPDATE CASCADE;
ALTER TABLE `lends`
    ADD CONSTRAINT `fk_lends_book_id`
        FOREIGN KEY (`book_id`) REFERENCES `books` (`book_id`)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE `lends`
    ADD CONSTRAINT `fk_lends_reader_id`
        FOREIGN KEY (`reader_id`) REFERENCES `reader_infos` (`reader_id`)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE `reserves`
    ADD CONSTRAINT `fk_reserves_book_id`
        FOREIGN KEY (`book_id`) REFERENCES `books` (`book_id`)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE `reserves`
    ADD CONSTRAINT `fk_reserves_reader_id`
        FOREIGN KEY (`reader_id`) REFERENCES `reader_infos` (`reader_id`)
            ON DELETE CASCADE
            ON UPDATE CASCADE;

ALTER TABLE `reader_cards`
    ADD CONSTRAINT `fk_reader_cards_reader_id`
        FOREIGN KEY (`reader_id`) REFERENCES `reader_infos` (`reader_id`)
            ON DELETE CASCADE
            ON UPDATE CASCADE;
COMMIT;