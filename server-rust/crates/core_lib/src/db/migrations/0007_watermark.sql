-- Phase 2.2: 水印支持
-- 为 photos 表添加 watermark_url 列，存储水印图片的 URL

ALTER TABLE photos ADD COLUMN watermark_url TEXT;
