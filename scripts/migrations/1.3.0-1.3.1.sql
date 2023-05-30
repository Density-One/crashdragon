BEGIN TRANSACTION;

ALTER TABLE users ADD COLUMN password TEXT;
INSERT INTO users (name, is_admin, password) VALUES ('TestAdmin', TRUE, crypt('removethisuser', gen_salt('bf')));

END TRANSACTION;