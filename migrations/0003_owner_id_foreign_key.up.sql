ALTER TABLE urls
ADD CONSTRAINT urls_owner_id_fkey
FOREIGN KEY (ownerid) REFERENCES url_users(id)
ON DELETE CASCADE;
