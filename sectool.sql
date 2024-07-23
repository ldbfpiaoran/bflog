/*
 Navicat Premium Data Transfer

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 50744 (5.7.44)
 Source Host           : localhost:3306
 Source Schema         : sectool

 Target Server Type    : MySQL
 Target Server Version : 50744 (5.7.44)
 File Encoding         : 65001

 Date: 06/07/2024 16:15:41
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for dns_rule
-- ----------------------------
DROP TABLE IF EXISTS `dns_rule`;
CREATE TABLE `dns_rule` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `ip_addresses` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for dnslog
-- ----------------------------
DROP TABLE IF EXISTS `dnslog`;
CREATE TABLE `dnslog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `receive_ip` varchar(255) DEFAULT NULL,
  `query_name` varchar(255) DEFAULT NULL,
  `query_type` varchar(255) DEFAULT NULL,
  `created_time` datetime(6) DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=37 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for http_request_log
-- ----------------------------
DROP TABLE IF EXISTS `http_request_log`;
CREATE TABLE `http_request_log` (
  `ID` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `hostname` varchar(255) NOT NULL,
  `timestamp` datetime NOT NULL,
  `remote_addr` varchar(255) NOT NULL,
  `method` varchar(10) NOT NULL,
  `url` text NOT NULL,
  `header` text NOT NULL,
  `body` text NOT NULL,
  `path` text NOT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB AUTO_INCREMENT=58 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for http_response
-- ----------------------------
DROP TABLE IF EXISTS `http_response`;
CREATE TABLE `http_response` (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `method` varchar(255) DEFAULT NULL,
  `path` varchar(255) DEFAULT NULL,
  `status_code` int(11) DEFAULT NULL,
  `redirect_url` varchar(255) DEFAULT NULL,
  `header` json DEFAULT NULL,
  `body` text,
  `create_at` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;
