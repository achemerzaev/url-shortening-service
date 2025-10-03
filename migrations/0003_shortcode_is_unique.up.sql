ALTER TABLE urls
ADD CONSTRAINT unique_url UNIQUE (Url),
    CONSTRAINT unique_shortcode UNIQUE (ShortCode);