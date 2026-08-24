-- Phase 2.2: 水印支持 · MariaDB/MySQL 版
ALTER TABLE photos ADD COLUMN watermark_url VARCHAR(768) NULL;
