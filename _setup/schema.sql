-- MySQL Workbench Forward Engineering

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION';

-- -----------------------------------------------------
-- Schema synk
-- -----------------------------------------------------

-- -----------------------------------------------------
-- Schema synk
-- -----------------------------------------------------
CREATE SCHEMA IF NOT EXISTS `synk` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin ;
USE `synk` ;

-- -----------------------------------------------------
-- Table `synk`.`user`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`user` (
  `user_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_name` VARCHAR(200) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NULL DEFAULT NULL,
  `user_email` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `user_pass` VARCHAR(255) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  UNIQUE INDEX `user_email_UNIQUE` (`user_email` ASC) VISIBLE)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`auth`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`auth` (
  `oauth_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `auth_refresh_token` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NULL DEFAULT NULL,
  `auth_platform` ENUM('twitter', 'linkedin', 'instagram') CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `auth_oauth_id` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `user_id` INT UNSIGNED NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`oauth_id`),
  UNIQUE INDEX `oauth_id_UNIQUE` (`oauth_id` ASC) VISIBLE,
  INDEX `fk_user_id_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_auth_user_id`
    FOREIGN KEY (`user_id`)
    REFERENCES `synk`.`user` (`user_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`template`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`template` (
  `template_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `template_name` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `template_content` TEXT CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `template_url_import` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NULL DEFAULT NULL,
  `user_id` INT UNSIGNED NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`template_id`),
  UNIQUE INDEX `template_id_UNIQUE` (`template_id` ASC) VISIBLE,
  INDEX `fk_user_id_idx` (`user_id` ASC) VISIBLE,
  CONSTRAINT `fk_template_user_id`
    FOREIGN KEY (`user_id`)
    REFERENCES `synk`.`user` (`user_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`color`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`color` (
  `color_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `color_name` VARCHAR(50) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `color_hex` VARCHAR(6) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  PRIMARY KEY (`color_id`),
  UNIQUE INDEX `color_id_UNIQUE` (`color_id` ASC) VISIBLE,
  UNIQUE INDEX `color_hex_UNIQUE` (`color_hex` ASC) VISIBLE)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`integration_profile`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`integration_profile` (
  `int_profile_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `color_id` INT UNSIGNED NULL DEFAULT NULL,
  `int_profile_name` VARCHAR(200) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`int_profile_id`),
  UNIQUE INDEX `int_profile_id_UNIQUE` (`int_profile_id` ASC) VISIBLE,
  INDEX `fk_color_id_idx` (`color_id` ASC) VISIBLE,
  CONSTRAINT `fk_integration_profile_color_id`
    FOREIGN KEY (`color_id`)
    REFERENCES `synk`.`color` (`color_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`tag`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`tag` (
  `tag_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tag_label` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `color_id` INT UNSIGNED NULL DEFAULT NULL,
  `template_id` INT UNSIGNED NULL DEFAULT NULL,
  `integration_id` INT UNSIGNED NULL DEFAULT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`tag_id`),
  UNIQUE INDEX `tag_id_UNIQUE` (`tag_id` ASC) VISIBLE,
  INDEX `fk_color_id_idx` (`color_id` ASC) VISIBLE,
  INDEX `fk_template_id_idx` (`template_id` ASC) VISIBLE,
  INDEX `fk_int_profile_id_idx` (`integration_id` ASC) VISIBLE,
  CONSTRAINT `fk_tag_color_id`
    FOREIGN KEY (`color_id`)
    REFERENCES `synk`.`color` (`color_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_tag_template_id`
    FOREIGN KEY (`template_id`)
    REFERENCES `synk`.`template` (`template_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_tag_int_profile_id`
    FOREIGN KEY (`integration_id`)
    REFERENCES `synk`.`integration_profile` (`int_profile_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`integration_credential`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`integration_credential` (
  `int_credential_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `int_credential_name` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `int_credential_type` ENUM('twitter', 'linkedin', 'instagram') CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `int_credential_config` JSON NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`int_credential_id`),
  UNIQUE INDEX `int_credential_id_UNIQUE` (`int_credential_id` ASC) VISIBLE)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`integration_group`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`integration_group` (
  `int_group_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `int_profile_id` INT UNSIGNED NOT NULL,
  `int_credential_id` INT UNSIGNED NOT NULL,
  PRIMARY KEY (`int_group_id`),
  UNIQUE INDEX `int_group_id_UNIQUE` (`int_group_id` ASC) VISIBLE,
  INDEX `fk_int_credential_id_idx` (`int_credential_id` ASC) VISIBLE,
  INDEX `fk_int_profile_id_idx` (`int_profile_id` ASC) VISIBLE,
  CONSTRAINT `fk_integration_group_int_credential_id`
    FOREIGN KEY (`int_credential_id`)
    REFERENCES `synk`.`integration_credential` (`int_credential_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_integration_group_int_profile_id`
    FOREIGN KEY (`int_profile_id`)
    REFERENCES `synk`.`integration_profile` (`int_profile_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_bin;


-- -----------------------------------------------------
-- Table `synk`.`post`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`post` (
  `post_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `post_name` VARCHAR(100) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `post_content` TEXT CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `template_id` INT UNSIGNED NOT NULL,
  `int_profile_id` INT UNSIGNED NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  `deleted_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`post_id`),
  UNIQUE INDEX `post_id_UNIQUE` (`post_id` ASC) VISIBLE,
  INDEX `fk_template_id_idx` (`template_id` ASC) VISIBLE,
  INDEX `fk_int_profile_idx` (`int_profile_id` ASC) VISIBLE,
  CONSTRAINT `fk_post_template_id`
    FOREIGN KEY (`template_id`)
    REFERENCES `synk`.`template` (`template_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_post_int_profile`
    FOREIGN KEY (`int_profile_id`)
    REFERENCES `synk`.`integration_profile` (`int_profile_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;


-- -----------------------------------------------------
-- Table `synk`.`publication`
-- -----------------------------------------------------
CREATE TABLE IF NOT EXISTS `synk`.`publication` (
  `publication_id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `publication_status` ENUM('pending', 'failed', 'published') CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NOT NULL,
  `publication_error_code` VARCHAR(20) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NULL DEFAULT NULL,
  `publication_error_desc` VARCHAR(200) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_bin' NULL DEFAULT NULL,
  `post_id` INT UNSIGNED NOT NULL,
  `int_credential_id` INT UNSIGNED NOT NULL,
  `created_at` DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NULL DEFAULT NULL,
  PRIMARY KEY (`publication_id`),
  UNIQUE INDEX `publication_id_UNIQUE` (`publication_id` ASC) VISIBLE,
  INDEX `fk_post_id_idx` (`post_id` ASC) VISIBLE,
  INDEX `fk_int_credential_id_idx` (`int_credential_id` ASC) VISIBLE,
  CONSTRAINT `fk_publication_post_id`
    FOREIGN KEY (`post_id`)
    REFERENCES `synk`.`post` (`post_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT `fk_publication_int_credential_id`
    FOREIGN KEY (`int_credential_id`)
    REFERENCES `synk`.`integration_credential` (`int_credential_id`)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION)
ENGINE = InnoDB;


SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
