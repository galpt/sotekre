-- phpMyAdmin SQL Dump
-- Sotekre Menu Tree System - Sample Data
-- Import this file to XAMPP/phpMyAdmin to populate the menu tree
--
-- Host: 127.0.0.1
-- Generation Time: Feb 06, 2026
-- Server version: MySQL 8.0+
-- Database: sotekre_dev

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `sotekre_dev`
--

-- --------------------------------------------------------

--
-- Clean existing data (optional - remove if you want to keep existing menus)
--

SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE `menus`;
SET FOREIGN_KEY_CHECKS = 1;

-- --------------------------------------------------------

--
-- Table structure for table `menus`
-- This will be created automatically by the Go application if it doesn't exist
-- But we include it here for reference
--

CREATE TABLE IF NOT EXISTS `menus` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL,
  `url` varchar(255) DEFAULT NULL,
  `parent_id` bigint unsigned DEFAULT NULL,
  `order` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_menus_parent_id` (`parent_id`),
  KEY `idx_menus_order` (`order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Dumping data for table `menus`
-- This matches the structure shown in the Figma design screenshot
--

INSERT INTO `menus` (`id`, `title`, `url`, `parent_id`, `order`, `created_at`, `updated_at`) VALUES
-- Root level: system management
(1, 'system management', '/system-management', NULL, 1, NOW(), NOW()),

-- Level 1: System Management
(2, 'System Management', '/system-management/main', 1, 1, NOW(), NOW()),

-- Level 2: Systems (under System Management)
(3, 'Systems', '/systems', 2, 1, NOW(), NOW()),

-- Level 3: System Code (under Systems)
(4, 'System Code', '/systems/code', 3, 1, NOW(), NOW()),

-- Level 4: Under System Code
(5, 'Code Registration', '/systems/code/registration', 4, 1, NOW(), NOW()),
(6, 'Code Registration - 2', '/systems/code/registration-2', 4, 2, NOW(), NOW()),
(7, 'Properties', '/systems/properties', 4, 3, NOW(), NOW()),

-- Level 3: Menus (under Systems)
(8, 'Menus', '/systems/menus', 3, 2, NOW(), NOW()),

-- Level 4: Under Menus
(9, 'Menu Registration', '/systems/menus/registration', 8, 1, NOW(), NOW()),

-- Level 3: API List (under Systems)
(10, 'API List', '/systems/api', 3, 3, NOW(), NOW()),

-- Level 4: Under API List
(11, 'API Registration', '/systems/api/registration', 10, 1, NOW(), NOW()),
(12, 'API Edit', '/systems/api/edit', 10, 2, NOW(), NOW()),

-- Level 2: Users & Groups (under System Management)
(13, 'Users & Groups', '/users-groups', 2, 2, NOW(), NOW()),

-- Level 3: Users (under Users & Groups)
(14, 'Users', '/users', 13, 1, NOW(), NOW()),

-- Level 4: Under Users
(15, 'User Account Registration', '/users/account-registration', 14, 1, NOW(), NOW()),

-- Level 3: Groups (under Users & Groups)
(16, 'Groups', '/groups', 13, 2, NOW(), NOW()),

-- Level 4: Under Groups
(17, 'User Group Registration', '/groups/registration', 16, 1, NOW(), NOW()),

-- Level 2: 사용자 승인 (User Approval - Korean text as shown in screenshot)
(18, '사용자 승인', '/user-approval', 2, 3, NOW(), NOW()),

-- Level 3: Under User Approval
(19, '사용자 승인 상세', '/user-approval/detail', 18, 1, NOW(), NOW());

-- --------------------------------------------------------

--
-- Update AUTO_INCREMENT for table `menus`
--

ALTER TABLE `menus`
  MODIFY `id` bigint unsigned NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=20;

COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;

--
-- Quick Import Instructions:
-- 1. Open phpMyAdmin (http://localhost/phpmyadmin)
-- 2. Select database 'sotekre_dev' (create if needed)
-- 3. Go to "Import" tab
-- 4. Choose this file (backend/database/sotekre_menus_import.sql)
-- 5. Click "Go" to import
-- 6. Refresh frontend to see the data
--
-- For detailed instructions, see backend/database/README.md
--
