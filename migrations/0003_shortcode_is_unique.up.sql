ALTER TABLE urls
ADD CONSTRAINT unique_url UNIQUE (Url),
ADD CONSTRAINT unique_shortcode UNIQUE (ShortCode);