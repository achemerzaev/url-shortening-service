ALTER TABLE urls
ADD CONSTRAINT unique_url UNIQUE (url),
ADD CONSTRAINT unique_shortcode UNIQUE (ShortCode);