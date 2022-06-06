
DROP TABLE IF EXISTS users;

CREATE TABLE public.users
(
    "UserID" serial NOT NULL,
    "Username" character varying[] NOT NULL,
    "UserType" character varying[] NOT NULL,
    PRIMARY KEY ("UserID")
);

ALTER TABLE IF EXISTS public.users
    OWNER to postgres;

-- Table: public.log

-- DROP TABLE IF EXISTS public.log;

CREATE TABLE IF NOT EXISTS public.log
(
    "Id" serial NOT NULL ,
    "State" integer NOT NULL,
    "Message" character varying[] NOT NULL,
    "FromUserID" integer NOT NULL,
    "Value" bytea[] NOT NULL,
    PRIMARY KEY ("Id")
)

ALTER TABLE IF EXISTS public.log
    OWNER to postgres;


DROP ROLE IF EXISTS my_user;
CREATE ROLE my_user LOGIN PASSWORD 'my_password';
GRANT INSERT, SELECT TO my_user;

